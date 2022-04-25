package api

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
)

// Presence of this header in the HTTP request indicates the chunk needs to be pinned.
const PinHeaderName = "x-swarm-pin"

func (s *server) chunkUploadHandler(w http.ResponseWriter, r *http.Request) {
	addr := mux.Vars(r)["addr"]
	ctx := r.Context()

	address, err := swarm.ParseHexAddress(addr)
	if err != nil {
		s.Logger.Debugf("chunk: parse chunk address %s: %v", addr, err)
		s.Logger.Error("chunk: error uploading chunk")
		jsonhttp.BadRequest(w, "invalid chunk address")
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.Logger.Debugf("chunk: read chunk data error: %v, addr %s", err, address)
		s.Logger.Error("chunk: read chunk data error")
		jsonhttp.InternalServerError(w, "cannot read chunk data")
		return

	}

	_, err = s.Storer.Put(ctx, storage.ModePutUpload, swarm.NewChunk(address, data))
	if err != nil {
		s.Logger.Debugf("chunk: chunk write error: %v, addr %s", err, address)
		s.Logger.Error("chunk: chunk write error")
		jsonhttp.BadRequest(w, "chunk write error")
		return
	}

	// Check if this chunk needs to pinned and pin it
	pinHeaderValues := r.Header.Get(PinHeaderName)
	if pinHeaderValues != "" && strings.ToLower(pinHeaderValues) == "true" {
		err = s.Storer.Set(ctx, storage.ModeSetPin, address)
		if err != nil {
			s.Logger.Debugf("chunk: chunk pinning error: %v, addr %s", err, address)
			s.Logger.Error("chunk: chunk pinning error")
			jsonhttp.InternalServerError(w, "cannot pin chunk")
			return
		}
	}

	jsonhttp.OK(w, nil)
}

func (s *server) chunkGetHandler(w http.ResponseWriter, r *http.Request) {
	addr := mux.Vars(r)["addr"]
	ctx := r.Context()

	address, err := swarm.ParseHexAddress(addr)
	if err != nil {
		s.Logger.Debugf("chunk: parse chunk address %s: %v", addr, err)
		s.Logger.Error("chunk: parse chunk address error")
		jsonhttp.BadRequest(w, "invalid chunk address")
		return
	}

	chunk, err := s.Storer.Get(ctx, storage.ModeGetRequest, address)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			s.Logger.Trace("chunk: chunk not found. addr %s", address)
			jsonhttp.NotFound(w, "chunk not found")
			return

		}
		s.Logger.Debugf("chunk: chunk read error: %v ,addr %s", err, address)
		s.Logger.Error("chunk: chunk read error")
		jsonhttp.InternalServerError(w, "chunk read error")
		return
	}
	w.Header().Set("Content-Type", "binary/octet-stream")
	_, _ = io.Copy(w, bytes.NewReader(chunk.Data()))
}
