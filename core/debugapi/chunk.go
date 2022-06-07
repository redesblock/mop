package debugapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
)

func (s *server) hasChunkHandler(w http.ResponseWriter, r *http.Request) {
	addr, err := swarm.ParseHexAddress(mux.Vars(r)["address"])
	if err != nil {
		s.Logger.Debugf("debug api: parse chunk address: %v", err)
		jsonhttp.BadRequest(w, "bad address")
		return
	}

	has, err := s.Storer.Has(r.Context(), addr)
	if err != nil {
		s.Logger.Debugf("debug api: localstore has: %v", err)
		jsonhttp.BadRequest(w, err)
		return
	}

	if !has {
		jsonhttp.NotFound(w, nil)
		return
	}
	jsonhttp.OK(w, nil)
}

func (s *server) removeChunk(w http.ResponseWriter, r *http.Request) {
	addr, err := swarm.ParseHexAddress(mux.Vars(r)["address"])
	if err != nil {
		s.Logger.Debugf("debug api: parse chunk address: %v", err)
		jsonhttp.BadRequest(w, "bad address")
		return
	}

	has, err := s.Storer.Has(r.Context(), addr)
	if err != nil {
		s.Logger.Debugf("debug api: localstore remove: %v", err)
		jsonhttp.BadRequest(w, err)
		return
	}

	if !has {
		jsonhttp.OK(w, nil)
		return
	}

	err = s.Storer.Set(r.Context(), storage.ModeSetRemove, addr)
	if err != nil {
		s.Logger.Debugf("debug api: localstore remove: %v", err)
		jsonhttp.InternalServerError(w, err)
		return
	}
	jsonhttp.OK(w, nil)
}
