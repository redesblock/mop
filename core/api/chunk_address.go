package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/storer/storage"
)

func (s *Service) hasChunkHandler(w http.ResponseWriter, r *http.Request) {
	str := mux.Vars(r)["address"]
	addr, err := cluster.ParseHexAddress(str)
	if err != nil {
		s.logger.Debug("has chunk: parse chunk address string failed", "string", str, "error", err)
		jsonhttp.BadRequest(w, "bad address")
		return
	}

	has, err := s.storer.Has(r.Context(), addr)
	if err != nil {
		s.logger.Debug("has chunk: has chunk failed", "chunk_address", addr, "error", err)
		jsonhttp.BadRequest(w, err)
		return
	}

	if !has {
		jsonhttp.NotFound(w, nil)
		return
	}
	jsonhttp.OK(w, nil)
}

func (s *Service) removeChunk(w http.ResponseWriter, r *http.Request) {
	str := mux.Vars(r)["address"]
	addr, err := cluster.ParseHexAddress(str)
	if err != nil {
		s.logger.Debug("remove chunk: parse chunk address string failed", "string", str, "error", err)
		jsonhttp.BadRequest(w, "bad address")
		return
	}

	has, err := s.storer.Has(r.Context(), addr)
	if err != nil {
		s.logger.Debug("remove chunk: has chunk failed", "chunk_address", addr, "error", err)
		jsonhttp.BadRequest(w, err)
		return
	}

	if !has {
		jsonhttp.OK(w, nil)
		return
	}

	err = s.storer.Set(r.Context(), storage.ModeSetRemove, addr)
	if err != nil {
		s.logger.Debug("remove chunk: remove chunk failed", "chunk_address", addr, "error", err)
		jsonhttp.InternalServerError(w, err)
		return
	}
	jsonhttp.OK(w, nil)
}
