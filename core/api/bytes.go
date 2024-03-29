package api

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/chunk/cac"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/incentives/voucher"
	"github.com/redesblock/mop/core/mctx"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/redesblock/mop/core/tags"
	"github.com/redesblock/mop/core/tracer"
	"github.com/redesblock/mop/core/util/ioutil"
)

type bytesPostResponse struct {
	Reference cluster.Address `json:"reference"`
}

// bytesUploadHandler handles upload of raw binary data of arbitrary length.
func (s *Service) bytesUploadHandler(w http.ResponseWriter, r *http.Request) {
	logger := tracer.NewLoggerWithTraceID(r.Context(), s.logger)

	putter, wait, err := s.newStamperPutter(r)
	if err != nil {
		logger.Debug("bytes upload: get putter failed", "error", err)
		logger.Error(nil, "bytes upload: get putter failed")
		jsonhttp.BadRequest(w, nil)
		return
	}

	if strings.Contains(strings.ToLower(r.Header.Get(contentTypeHeader)), "multipart/form-data") {
		logger.Error(nil, "bytes upload: multipart uploads are not supported on this endpoint")
		jsonhttp.BadRequest(w, "multipart uploads not supported")
		return
	}

	tag, created, err := s.getOrCreateTag(r.Header.Get(ClusterTagHeader))
	if err != nil {
		logger.Debug("bytes upload: get or create tag failed", "error", err)
		logger.Error(nil, "bytes upload: get or create tag failed")
		jsonhttp.InternalServerError(w, "cannot get or create tag")
		return
	}

	if !created {
		// only in the case when tag is sent via header (i.e. not created by this request)
		if estimatedTotalChunks := requestCalculateNumberOfChunks(r); estimatedTotalChunks > 0 {
			err = tag.IncN(tags.TotalChunks, estimatedTotalChunks)
			if err != nil {
				s.logger.Debug("bytes upload: increment tag failed", "error", err)
				s.logger.Error(nil, "bytes upload: increment tag failed")
				jsonhttp.InternalServerError(w, "increment tag")
				return
			}
		}
	}

	// Add the tag to the context
	ctx := mctx.SetTag(r.Context(), tag)
	p := requestPipelineFn(putter, r)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	pr := ioutil.TimeoutReader(ctx, r.Body, time.Minute, func(n uint64) {
		logger.Error(nil, "bytes upload: idle read timeout exceeded")
		logger.Debug("bytes upload: idle read timeout exceeded", "bytes_read", n)
		cancel()
	})
	address, err := p(ctx, pr, nil)
	if err != nil {
		logger.Debug("bytes upload: split write all failed", "error", err)
		logger.Error(nil, "bytes upload: split write all failed")
		switch {
		case errors.Is(err, voucher.ErrBucketFull):
			jsonhttp.PaymentRequired(w, "batch is overissued")
		default:
			jsonhttp.InternalServerError(w, "bytes upload: split write all failed")
		}
		return
	}
	if err = wait(); err != nil {
		logger.Debug("bytes upload: chainsync chunks failed", "error", err)
		logger.Error(nil, "bytes upload: chainsync chunks failed")
		jsonhttp.InternalServerError(w, "bytes upload: chainsync chunks failed")
		return
	}

	if created {
		_, err = tag.DoneSplit(address)
		if err != nil {
			logger.Debug("bytes upload: done split failed", "error", err)
			logger.Error(nil, "bytes upload: done split failed")
			jsonhttp.InternalServerError(w, "bytes upload: done split filed")
			return
		}
	}

	if strings.ToLower(r.Header.Get(ClusterPinHeader)) == "true" {
		if err := s.pinning.CreatePin(ctx, address, false); err != nil {
			logger.Debug("bytes upload: pins creation failed", "address", address, "error", err)
			logger.Error(nil, "bytes upload: pins creation failed")
			jsonhttp.InternalServerError(w, "bytes upload: create ping failed")
			return
		}
	}

	w.Header().Set(ClusterTagHeader, fmt.Sprint(tag.Uid))
	w.Header().Set("Access-Control-Expose-Headers", ClusterTagHeader)
	jsonhttp.Created(w, bytesPostResponse{
		Reference: address,
	})
}

// bytesGetHandler handles retrieval of raw binary data of arbitrary length.
func (s *Service) bytesGetHandler(w http.ResponseWriter, r *http.Request) {
	logger := tracer.NewLoggerWithTraceID(r.Context(), s.logger)
	nameOrHex := mux.Vars(r)["address"]

	address, err := s.resolveNameOrAddress(nameOrHex)
	if err != nil {
		logger.Debug("bytes: parse address string failed", nameOrHex, err)
		logger.Error(nil, "bytes: parse address string failed")
		jsonhttp.NotFound(w, nil)
		return
	}

	additionalHeaders := http.Header{
		contentTypeHeader: {"application/octet-stream"},
	}

	s.downloadHandler(w, r, address, additionalHeaders, true)
}

func (s *Service) bytesHeadHandler(w http.ResponseWriter, r *http.Request) {
	logger := tracer.NewLoggerWithTraceID(r.Context(), s.logger)
	nameOrHex := mux.Vars(r)["address"]

	address, err := s.resolveNameOrAddress(nameOrHex)
	if err != nil {
		logger.Debug("bytes: parse address string failed", "string", nameOrHex, "error", err)
		logger.Error(nil, "bytes: parse address string failed")
		w.WriteHeader(http.StatusBadRequest) // HEAD requests do not write a body
		return
	}
	ch, err := s.storer.Get(r.Context(), storage.ModeGetRequest, address)
	if err != nil {
		logger.Debug("bytes: get root chunk failed", "chunk_address", address, "error", err)
		logger.Error(nil, "bytes: get rook chunk failed")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Add("Access-Control-Expose-Headers", "Accept-Ranges, Content-Encoding")
	w.Header().Add(contentTypeHeader, "application/octet-stream")
	var span int64

	if cac.Valid(ch) {
		span = int64(binary.LittleEndian.Uint64(ch.Data()[:cluster.SpanSize]))
	} else {
		// soc
		span = int64(len(ch.Data()))
	}
	w.Header().Add("Content-Length", strconv.FormatInt(span, 10))
	w.WriteHeader(http.StatusOK) // HEAD requests do not write a body
}
