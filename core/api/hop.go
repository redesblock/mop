package api

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash"
	"io"
	"io/ioutil"
	"net/http"

	bmtlegacy "github.com/ethersphere/bmt/legacy"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/sha3"

	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
)

type HopPostResponse struct {
	Hash swarm.Address `json:"hash"`
}

func hashFunc() hash.Hash {
	return sha3.NewLegacyKeccak256()
}

func (s *server) hopUploadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.Logger.Debugf("hop: read error: %v", err)
		s.Logger.Error("hop: read error")
		jsonhttp.InternalServerError(w, "cannot read request")
		return
	}

	p := bmtlegacy.NewTreePool(hashFunc, swarm.Branches, bmtlegacy.PoolSize)
	hasher := bmtlegacy.New(p)
	span := len(data)
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(span))
	data = append(b, data...)
	err = hasher.SetSpan(int64(span))
	if err != nil {
		s.Logger.Debugf("hop: hasher set span: %v", err)
		s.Logger.Error("hop: hash data error")
		jsonhttp.InternalServerError(w, "cannot hash data")
		return
	}
	_, err = hasher.Write(data[8:])
	if err != nil {
		s.Logger.Debugf("hop: hasher write: %v", err)
		s.Logger.Error("hop: hash data error")
		jsonhttp.InternalServerError(w, "cannot hash data")
		return
	}
	addr := swarm.NewAddress(hasher.Sum(nil))
	_, err = s.Storer.Put(ctx, storage.ModePutUpload, swarm.NewChunk(addr, data[8:]))
	if err != nil {
		s.Logger.Debugf("hop: write error: %v, addr %s", err, addr)
		s.Logger.Error("hop: write error")
		jsonhttp.InternalServerError(w, "write error")
		return
	}
	jsonhttp.OK(w, HopPostResponse{Hash: addr})
}

func (s *server) hopGetHandler(w http.ResponseWriter, r *http.Request) {
	addr := mux.Vars(r)["address"]
	ctx := r.Context()

	address, err := swarm.ParseHexAddress(addr)
	if err != nil {
		s.Logger.Debugf("hop: parse address %s: %v", addr, err)
		s.Logger.Error("hop: parse address error")
		jsonhttp.BadRequest(w, "invalid address")
		return
	}

	chunk, err := s.Storer.Get(ctx, storage.ModeGetRequest, address)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			s.Logger.Trace("hop: not found. addr %s", address)
			jsonhttp.NotFound(w, "not found")
			return

		}
		s.Logger.Debugf("hop: read error: %v ,addr %s", err, address)
		s.Logger.Error("hop: read error")
		jsonhttp.InternalServerError(w, "read error")
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = io.Copy(w, bytes.NewReader(chunk.Data()))
}
