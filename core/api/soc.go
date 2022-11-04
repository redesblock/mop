package api

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/chunk/cac"
	"github.com/redesblock/mop/core/chunk/soc"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/incentives/voucher"
)

type socPostResponse struct {
	Reference cluster.Address `json:"reference"`
}

func (s *Service) socUploadHandler(w http.ResponseWriter, r *http.Request) {
	str := mux.Vars(r)["owner"]
	owner, err := hex.DecodeString(str)
	if err != nil {
		s.logger.Debug("soc upload: parse owner string failed", "string", str, "error", err)
		s.logger.Error(nil, "soc upload: parse owner string failed")
		jsonhttp.BadRequest(w, "bad owner")
		return
	}
	str = mux.Vars(r)["id"]
	id, err := hex.DecodeString(mux.Vars(r)["id"])
	if err != nil {
		s.logger.Debug("soc upload: parse id string failed", "string", str, "error", err)
		s.logger.Error(nil, "soc upload: parse id string failed")
		jsonhttp.BadRequest(w, "bad id")
		return
	}

	sigStr := r.URL.Query().Get("sig")
	if sigStr == "" {
		s.logger.Debug("soc upload: empty sig string")
		s.logger.Error(nil, "soc upload: empty sig string")
		jsonhttp.BadRequest(w, "empty signature")
		return
	}

	sig, err := hex.DecodeString(sigStr)
	if err != nil {
		s.logger.Debug("soc upload: decode sig string failed", "string", sigStr, "error", err)
		s.logger.Error(nil, "soc upload: decode sig string failed")
		jsonhttp.BadRequest(w, "bad signature")
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		if jsonhttp.HandleBodyReadError(err, w) {
			return
		}
		s.logger.Debug("soc upload: read body failed", "error", err)
		s.logger.Error(nil, "soc upload: read body failed")
		jsonhttp.InternalServerError(w, "cannot read chunk data")
		return
	}

	if len(data) < cluster.SpanSize {
		s.logger.Debug("soc upload: chunk data too short")
		s.logger.Error(nil, "soc upload: chunk data too short")
		jsonhttp.BadRequest(w, "short chunk data")
		return
	}

	if len(data) > cluster.ChunkSize+cluster.SpanSize {
		s.logger.Debug("soc upload: chunk data exceeds required length", "required_length", cluster.ChunkSize+cluster.SpanSize)
		s.logger.Error(nil, "soc upload: chunk data exceeds required length")
		jsonhttp.RequestEntityTooLarge(w, "payload too large")
		return
	}

	ch, err := cac.NewWithDataSpan(data)
	if err != nil {
		s.logger.Debug("soc upload: create content addressed chunk failed", "error", err)
		s.logger.Error(nil, "soc upload: create content addressed chunk failed")
		jsonhttp.BadRequest(w, "chunk data error")
		return
	}

	ss, err := soc.NewSigned(id, ch, owner, sig)
	if err != nil {
		s.logger.Debug("soc upload: create soc failed", "id", id, "owner", owner, "error", err)
		s.logger.Error(nil, "soc upload: create soc failed")
		jsonhttp.Unauthorized(w, "invalid address")
		return
	}

	sch, err := ss.Chunk()
	if err != nil {
		s.logger.Debug("soc upload: read chunk data failed", "error", err)
		s.logger.Error(nil, "soc upload: read chunk data failed")
		jsonhttp.InternalServerError(w, "cannot read chunk data")
		return
	}

	if !soc.Valid(sch) {
		s.logger.Debug("soc upload: invalid chunk", "error", err)
		s.logger.Error(nil, "soc upload: invalid chunk")
		jsonhttp.Unauthorized(w, "invalid chunk")
		return
	}

	ctx := r.Context()

	has, err := s.storer.Has(ctx, sch.Address())
	if err != nil {
		s.logger.Debug("soc upload: has check failed", "chunk_address", sch.Address(), "error", err)
		s.logger.Error(nil, "soc upload: has check failed")
		jsonhttp.InternalServerError(w, "storage error")
		return
	}
	if has {
		s.logger.Error(nil, "soc upload: chunk already exists")
		jsonhttp.Conflict(w, "chunk already exists")
		return
	}
	batch, err := requestVoucherBatchId(r)
	if err != nil {
		s.logger.Debug("soc upload: parse voucher batch id failed", "error", err)
		s.logger.Error(nil, "soc upload: parse voucher batch id failed")
		jsonhttp.BadRequest(w, "invalid voucher batch id")
		return
	}

	i, err := s.post.GetStampIssuer(batch)
	if err != nil {
		s.logger.Debug("soc upload: get voucher batch issuer failed", "batch_id", fmt.Sprintf("%x", batch), "error", err)
		s.logger.Error(nil, "soc upload: get voucher batch issue")
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
	stamp, err := stamper.Stamp(sch.Address())
	if err != nil {
		s.logger.Debug("soc upload: stamp failed", "chunk_address", sch.Address(), "error", err)
		s.logger.Error(nil, "soc upload: stamp failed")
		switch {
		case errors.Is(err, voucher.ErrBucketFull):
			jsonhttp.PaymentRequired(w, "batch is overissued")
		default:
			jsonhttp.InternalServerError(w, "stamp error")
		}
		return
	}
	sch = sch.WithStamp(stamp)
	_, err = s.storer.Put(ctx, requestModePut(r), sch)
	if err != nil {
		s.logger.Debug("soc upload: write chunk failed", "chunk_address", sch.Address(), "error", err)
		s.logger.Error(nil, "soc upload: write chunk failed")
		jsonhttp.BadRequest(w, "chunk write error")
		return
	}

	if strings.ToLower(r.Header.Get(ClusterPinHeader)) == "true" {
		if err := s.pinning.CreatePin(ctx, sch.Address(), false); err != nil {
			s.logger.Debug("soc upload: create pins failed", "chunk_address", sch.Address(), "error", err)
			s.logger.Error(nil, "soc upload: create pins failed")
			jsonhttp.InternalServerError(w, "soc upload: creation of pins failed")
			return
		}
	}

	jsonhttp.Created(w, chunkAddressResponse{Reference: sch.Address()})
}
