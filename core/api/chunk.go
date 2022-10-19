package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/redesblock/mop/core/chunk/cac"

	"github.com/gorilla/mux"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/incentives/voucher"
	"github.com/redesblock/mop/core/mctx"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/redesblock/mop/core/tags"
)

type chunkAddressResponse struct {
	Reference cluster.Address `json:"reference"`
}

func (s *Service) processUploadRequest(
	r *http.Request,
) (ctx context.Context, tag *tags.Tag, putter storage.Putter, waitFn func() error, err error) {

	if h := r.Header.Get(ClusterTagHeader); h != "" {
		tag, err = s.getTag(h)
		if err != nil {
			s.logger.Debug("chunk upload: get tag failed", "error", err)
			s.logger.Error(nil, "chunk upload: get tag failed")
			return nil, nil, nil, nil, errors.New("cannot get tag")
		}

		// add the tag to the context if it exists
		ctx = mctx.SetTag(r.Context(), tag)
	} else {
		ctx = r.Context()
	}

	putter, wait, err := s.newStamperPutter(r)
	if err != nil {
		s.logger.Debug("chunk upload: putter failed", "error", err)
		s.logger.Error(nil, "chunk upload: putter failed")
		switch {
		case errors.Is(err, voucher.ErrNotFound):
			return nil, nil, nil, nil, errors.New("batch not found")
		case errors.Is(err, voucher.ErrNotUsable):
			return nil, nil, nil, nil, errors.New("batch not usable")
		}
		return nil, nil, nil, nil, err
	}

	return ctx, tag, putter, wait, nil
}

func (s *Service) chunkUploadHandler(w http.ResponseWriter, r *http.Request) {
	ctx, tag, putter, wait, err := s.processUploadRequest(r)
	if err != nil {
		jsonhttp.BadRequest(w, err.Error())
		return
	}

	if tag != nil {
		err = tag.Inc(tags.StateSplit)
		if err != nil {
			s.logger.Debug("chunk upload: increment tag failed", "error", err)
			s.logger.Error(nil, "chunk upload: increment tag failed")
			jsonhttp.InternalServerError(w, "increment tag")
			return
		}
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		if jsonhttp.HandleBodyReadError(err, w) {
			return
		}
		s.logger.Debug("chunk upload: read chunk data failed", "error", err)
		s.logger.Error(nil, "chunk upload: read chunk data failed")
		jsonhttp.InternalServerError(w, "cannot read chunk data")
		return
	}

	if len(data) < cluster.SpanSize {
		s.logger.Debug("chunk upload: insufficient data length")
		s.logger.Error(nil, "chunk upload: insufficient data length")
		jsonhttp.BadRequest(w, "insufficient data length")
		return
	}

	chunk, err := cac.NewWithDataSpan(data)
	if err != nil {
		s.logger.Debug("chunk upload: create chunk failed", "error", err)
		s.logger.Error(nil, "chunk upload: create chunk error")
		jsonhttp.InternalServerError(w, "create chunk error")
		return
	}

	seen, err := putter.Put(ctx, requestModePut(r), chunk)
	if err != nil {
		s.logger.Debug("chunk upload: write chunk failed", "chunk_address", chunk.Address(), "error", err)
		s.logger.Error(nil, "chunk upload: write chunk failed")
		switch {
		case errors.Is(err, voucher.ErrBucketFull):
			jsonhttp.PaymentRequired(w, "batch is overissued")
		default:
			jsonhttp.InternalServerError(w, "chunk write error")
		}
		return
	} else if len(seen) > 0 && seen[0] && tag != nil {
		err := tag.Inc(tags.StateSeen)
		if err != nil {
			s.logger.Debug("chunk upload: increment tag failed", "error", err)
			s.logger.Error(nil, "chunk upload: increment tag failed")
			jsonhttp.BadRequest(w, "increment tag")
			return
		}
	}

	if tag != nil {
		// indicate that the chunk is stored
		err = tag.Inc(tags.StateStored)
		if err != nil {
			s.logger.Debug("chunk upload: increment tag failed", "error", err)
			s.logger.Error(nil, "chunk upload: increment tag failed")
			jsonhttp.InternalServerError(w, "increment tag failed")
			return
		}
		w.Header().Set(ClusterTagHeader, fmt.Sprint(tag.Uid))
	}

	if strings.ToLower(r.Header.Get(ClusterPinHeader)) == "true" {
		if err := s.pinning.CreatePin(ctx, chunk.Address(), false); err != nil {
			s.logger.Debug("chunk upload: pins creation failed", "chunk_address", chunk.Address(), "error", err)
			s.logger.Error(nil, "chunk upload: pins creation failed")
			err = s.storer.Set(ctx, storage.ModeSetUnpin, chunk.Address())
			if err != nil {
				s.logger.Debug("chunk upload: pins deletion failed", "chunk_address", chunk.Address(), "error", err)
				s.logger.Error(nil, "chunk upload: pins deletion failed")
			}
			jsonhttp.InternalServerError(w, "chunk upload: creation of pins failed")
			return
		}
	}

	if err = wait(); err != nil {
		s.logger.Debug("chunk upload: chainsync chunk failed", "error", err)
		s.logger.Error(nil, "chunk upload: chainsync chunk failed")
		jsonhttp.InternalServerError(w, "chunk upload: chainsync failed")
		return
	}

	w.Header().Set("Access-Control-Expose-Headers", ClusterTagHeader)
	jsonhttp.Created(w, chunkAddressResponse{Reference: chunk.Address()})
}

func (s *Service) chunkGetHandler(w http.ResponseWriter, r *http.Request) {
	loggerV1 := s.logger.V(1).Build()

	nameOrHex := mux.Vars(r)["address"]
	ctx := r.Context()

	address, err := s.resolveNameOrAddress(nameOrHex)
	if err != nil {
		s.logger.Debug("chunk get: parse chunk address string failed", "string", nameOrHex, "error", err)
		s.logger.Error(nil, "chunk get: parse chunk address string failed")
		jsonhttp.NotFound(w, nil)
		return
	}

	chunk, err := s.storer.Get(ctx, storage.ModeGetRequest, address)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			loggerV1.Debug("chunk get: chunk not found", "address", address)
			jsonhttp.NotFound(w, "chunk get: chunk not found")
			return

		}
		s.logger.Debug("chunk get: read chunk failed", "chunk_address", address, "error", err)
		s.logger.Error(nil, "chunk get: read chunk failed")
		jsonhttp.InternalServerError(w, "read chunk failed")
		return
	}
	w.Header().Set(contentTypeHeader, "binary/octet-stream")
	_, _ = io.Copy(w, bytes.NewReader(chunk.Data()))
}
