package api

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/redesblock/mop/core/chunk/trojan"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/incentives/voucher"
)

const (
	writeDeadline   = 4 * time.Second // write deadline. should be smaller than the shutdown timeout on api close
	targetMaxLength = 3               // max target length in bytes, in order to prevent grieving by excess computation
)

func (s *Service) pssPostHandler(w http.ResponseWriter, r *http.Request) {
	topicVar := mux.Vars(r)["topic"]
	topic := trojan.NewTopic(topicVar)

	targetsVar := mux.Vars(r)["targets"]
	var targets trojan.Targets
	tgts := strings.Split(targetsVar, ",")

	for _, v := range tgts {
		target, err := hex.DecodeString(v)
		if err != nil {
			s.logger.Debug("psser post: decode target string failed", "string", target, "error", err)
			s.logger.Error(nil, "psser post: decode target string failed", "string", target)
			jsonhttp.BadRequest(w, "target is not valid hex string")
			return
		}
		if len(target) > targetMaxLength {
			s.logger.Debug("psser post: invalid target string length", "string", target, "length", len(target))
			s.logger.Error(nil, "psser post: invalid target string length", "string", target, "length", len(target))
			jsonhttp.BadRequest(w, fmt.Sprintf("hex string target exceeds max length of %d", targetMaxLength*2))
			return
		}
		targets = append(targets, target)
	}

	recipientQueryString := r.URL.Query().Get("recipient")
	var recipient *ecdsa.PublicKey
	if recipientQueryString == "" {
		// use topic-based encryption
		privkey := crypto.Secp256k1PrivateKeyFromBytes(topic[:])
		recipient = &privkey.PublicKey
	} else {
		var err error
		recipient, err = trojan.ParseRecipient(recipientQueryString)
		if err != nil {
			s.logger.Debug("psser post: parse recipient string failed", "string", recipientQueryString, "error", err)
			s.logger.Error(nil, "psser post: parse recipient string failed")
			jsonhttp.BadRequest(w, "psser recipient: invalid format")
			return
		}
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Debug("psser post: read body failed", "error", err)
		s.logger.Error(nil, "psser post: read body failed")
		jsonhttp.InternalServerError(w, "psser send failed")
		return
	}
	batch, err := requestVoucherBatchId(r)
	if err != nil {
		s.logger.Debug("psser post: decode voucher batch id failed", "error", err)
		s.logger.Error(nil, "psser post: decode voucher batch id failed")
		jsonhttp.BadRequest(w, "invalid voucher batch id")
		return
	}
	i, err := s.post.GetStampIssuer(batch)
	if err != nil {
		s.logger.Debug("psser post: get voucher batch issuer failed", "batch_id", fmt.Sprintf("%x", batch), "error", err)
		s.logger.Error(nil, "psser post: get voucher batch issuer failed")
		switch {
		case errors.Is(err, voucher.ErrNotFound):
			jsonhttp.BadRequest(w, "batch not found")
		case errors.Is(err, voucher.ErrNotUsable):
			jsonhttp.BadRequest(w, "batch not usable yet")
		default:
			jsonhttp.BadRequest(w, "voucher stamp issuer")
		}
		return
	}
	stamper := voucher.NewStamper(i, s.signer)

	err = s.pss.Send(r.Context(), topic, payload, stamper, recipient, targets)
	if err != nil {
		s.logger.Debug("psser post: send payload failed", "topic", fmt.Sprintf("%x", topic), "error", err)
		s.logger.Error(nil, "psser post: send payload failed")
		switch {
		case errors.Is(err, voucher.ErrBucketFull):
			jsonhttp.PaymentRequired(w, "batch is overissued")
		default:
			jsonhttp.InternalServerError(w, "psser send failed")
		}
		return
	}

	jsonhttp.Created(w, nil)
}

func (s *Service) pssWsHandler(w http.ResponseWriter, r *http.Request) {

	upgrader := websocket.Upgrader{
		ReadBufferSize:  cluster.ChunkSize,
		WriteBufferSize: cluster.ChunkSize,
		CheckOrigin:     s.checkOrigin,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Debug("psser ws: upgrade failed", "error", err)
		s.logger.Error(nil, "psser ws: upgrade failed")
		jsonhttp.InternalServerError(w, "psser ws: upgrade failed")
		return
	}

	t := mux.Vars(r)["topic"]
	s.wsWg.Add(1)
	go s.pumpWs(conn, t)
}

func (s *Service) pumpWs(conn *websocket.Conn, t string) {
	defer s.wsWg.Done()

	var (
		dataC  = make(chan []byte)
		gone   = make(chan struct{})
		topic  = trojan.NewTopic(t)
		ticker = time.NewTicker(s.WsPingPeriod)
		err    error
	)
	defer func() {
		ticker.Stop()
		_ = conn.Close()
	}()
	cleanup := s.pss.Register(topic, func(_ context.Context, m []byte) {
		dataC <- m
	})

	defer cleanup()

	conn.SetCloseHandler(func(code int, text string) error {
		s.logger.Debug("psser ws: client gone", "code", code, "message", text)
		close(gone)
		return nil
	})

	for {
		select {
		case b := <-dataC:
			err = conn.SetWriteDeadline(time.Now().Add(writeDeadline))
			if err != nil {
				s.logger.Debug("psser ws: set write deadline failed", "error", err)
				return
			}

			err = conn.WriteMessage(websocket.BinaryMessage, b)
			if err != nil {
				s.logger.Debug("psser ws: write message failed", "error", err)
				return
			}

		case <-s.quit:
			// shutdown
			err = conn.SetWriteDeadline(time.Now().Add(writeDeadline))
			if err != nil {
				s.logger.Debug("psser ws: set write deadline failed", "error", err)
				return
			}
			err = conn.WriteMessage(websocket.CloseMessage, []byte{})
			if err != nil {
				s.logger.Debug("psser ws: write close message failed", "error", err)
			}
			return
		case <-gone:
			// client gone
			return
		case <-ticker.C:
			err = conn.SetWriteDeadline(time.Now().Add(writeDeadline))
			if err != nil {
				s.logger.Debug("psser ws: set write deadline failed", "error", err)
				return
			}
			if err = conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				// error encountered while pinging client. client probably gone
				return
			}
		}
	}
}
