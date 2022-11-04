package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/chunk/cac"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/incentives/voucher"
	"github.com/redesblock/mop/core/mctx"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/redesblock/mop/core/tags"
)

const streamReadTimeout = 15 * time.Minute

var successWsMsg = []byte{}

func (s *Service) chunkUploadStreamHandler(w http.ResponseWriter, r *http.Request) {

	_, tag, putter, wait, err := s.processUploadRequest(r)
	if err != nil {
		jsonhttp.BadRequest(w, err.Error())
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  cluster.ChunkSize,
		WriteBufferSize: cluster.ChunkSize,
		CheckOrigin:     s.checkOrigin,
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Debug("chunk upload: upgrade failed", "error", err)
		s.logger.Error(nil, "chunk upload: upgrade failed")
		jsonhttp.BadRequest(w, "upgrade failed")
		return
	}

	cctx := context.Background()
	if tag != nil {
		cctx = mctx.SetTag(cctx, tag)
	}

	s.wsWg.Add(1)
	go s.handleUploadStream(
		cctx,
		c,
		tag,
		putter,
		requestModePut(r),
		strings.ToLower(r.Header.Get(ClusterPinHeader)) == "true",
		wait,
	)
}

func (s *Service) handleUploadStream(
	ctx context.Context,
	conn *websocket.Conn,
	tag *tags.Tag,
	putter storage.Putter,
	mode storage.ModePut,
	pin bool,
	wait func() error,
) {
	defer s.wsWg.Done()

	var (
		gone = make(chan struct{})
		err  error
	)
	defer func() {
		_ = conn.Close()
		if err = wait(); err != nil {
			s.logger.Error(err, "chunk upload stream: syncing chunks failed")
		}
	}()

	conn.SetCloseHandler(func(code int, text string) error {
		s.logger.Debug("chunk upload stream: client gone", "code", code, "message", text)
		close(gone)
		return nil
	})

	sendMsg := func(msgType int, buf []byte) error {
		err := conn.SetWriteDeadline(time.Now().Add(writeDeadline))
		if err != nil {
			return err
		}
		err = conn.WriteMessage(msgType, buf)
		if err != nil {
			return err
		}
		return nil
	}

	sendErrorClose := func(code int, errmsg string) {
		err := conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(code, errmsg),
			time.Now().Add(writeDeadline),
		)
		if err != nil {
			s.logger.Error(err, "chunk upload stream: failed sending close message")
		}
	}

	for {
		select {
		case <-s.quit:
			// shutdown
			sendErrorClose(websocket.CloseGoingAway, "node shutting down")
			return
		case <-gone:
			// client gone
			return
		default:
			// if there is no indication to stop, go ahead and read the next message
		}

		err = conn.SetReadDeadline(time.Now().Add(streamReadTimeout))
		if err != nil {
			s.logger.Debug("chunk upload stream: set read deadline failed", "error", err)
			s.logger.Error(nil, "chunk upload stream: set read deadline failed")
			return
		}

		mt, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Debug("chunk upload stream: read message failed", "error", err)
				s.logger.Error(nil, "chunk upload stream: read message failed")
			}
			return
		}

		if mt != websocket.BinaryMessage {
			s.logger.Debug("chunk upload stream: unexpected message received from client", "message_type", mt)
			s.logger.Error(nil, "chunk upload stream: unexpected message received from client")
			sendErrorClose(websocket.CloseUnsupportedData, "invalid message")
			return
		}

		if tag != nil {
			err = tag.Inc(tags.StateSplit)
			if err != nil {
				s.logger.Debug("chunk upload stream: incrementing tag failed", "error", err)
				s.logger.Error(nil, "chunk upload stream: incrementing tag failed")
				sendErrorClose(websocket.CloseInternalServerErr, "incrementing tag failed")
				return
			}
		}

		if len(msg) < cluster.SpanSize {
			s.logger.Debug("chunk upload stream: insufficient data")
			s.logger.Error(nil, "chunk upload stream: insufficient data")
			return
		}

		chunk, err := cac.NewWithDataSpan(msg)
		if err != nil {
			s.logger.Debug("chunk upload stream: create chunk failed", "error", err)
			s.logger.Error(nil, "chunk upload stream: create chunk failed")
			return
		}

		seen, err := putter.Put(ctx, mode, chunk)
		if err != nil {
			s.logger.Debug("chunk upload stream: write chunk failed", "address", chunk.Address(), "error", err)
			s.logger.Error(nil, "chunk upload stream: write chunk failed")
			switch {
			case errors.Is(err, voucher.ErrBucketFull):
				sendErrorClose(websocket.CloseInternalServerErr, "batch is overissued")
			default:
				sendErrorClose(websocket.CloseInternalServerErr, "chunk write error")
			}
			return
		} else if len(seen) > 0 && seen[0] && tag != nil {
			err := tag.Inc(tags.StateSeen)
			if err != nil {
				s.logger.Debug("chunk upload stream: increment tag failed", "error", err)
				s.logger.Error(nil, "chunk upload stream: increment tag")
				sendErrorClose(websocket.CloseInternalServerErr, "incrementing tag failed")
				return
			}
		}

		if tag != nil {
			// indicate that the chunk is stored
			err = tag.Inc(tags.StateStored)
			if err != nil {
				s.logger.Debug("chunk upload stream: increment tag failed", "error", err)
				s.logger.Error(nil, "chunk upload stream: increment tag failed")
				sendErrorClose(websocket.CloseInternalServerErr, "incrementing tag failed")
				return
			}
		}

		if pin {
			if err := s.pinning.CreatePin(ctx, chunk.Address(), false); err != nil {
				s.logger.Debug("chunk upload stream: pins creation failed", "chunk_address", chunk.Address(), "error", err)
				s.logger.Error(nil, "chunk upload stream: pins creation failed")
				// since we already increment the pins counter because of the ModePut, we need
				// to delete the pins here to prevent the pins counter from never going to 0
				err = s.storer.Set(ctx, storage.ModeSetUnpin, chunk.Address())
				if err != nil {
					s.logger.Debug("chunk upload stream: pins deletion failed", "chunk_address", chunk.Address(), "error", err)
					s.logger.Error(nil, "chunk upload stream: pins deletion failed")
				}
				sendErrorClose(websocket.CloseInternalServerErr, "failed creating pins")
				return
			}
		}

		err = sendMsg(websocket.BinaryMessage, successWsMsg)
		if err != nil {
			s.logger.Debug("chunk upload stream: sending success message failed", "error", err)
			s.logger.Error(nil, "chunk upload stream: sending success message failed")
			return
		}
	}
}
