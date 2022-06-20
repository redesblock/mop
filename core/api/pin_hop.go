package api

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/traversal"
)

// pinHop is used to pin an already uploaded content.
func (s *server) pinHop(w http.ResponseWriter, r *http.Request) {
	addr, err := swarm.ParseHexAddress(mux.Vars(r)["address"])
	if err != nil {
		s.logger.Debugf("pin hop: parse address: %v", err)
		s.logger.Error("pin hop: parse address")
		jsonhttp.BadRequest(w, "bad address")
		return
	}

	has, err := s.storer.Has(r.Context(), addr)
	if err != nil {
		s.logger.Debugf("pin hop: localstore has: %v", err)
		s.logger.Error("pin hop: store")
		jsonhttp.InternalServerError(w, err)
		return
	}

	if !has {
		_, err := s.storer.Get(r.Context(), storage.ModeGetRequest, addr)
		if err != nil {
			s.logger.Debugf("pin chunk: netstore get: %v", err)
			s.logger.Error("pin chunk: netstore")

			jsonhttp.NotFound(w, nil)
			return
		}
	}

	ctx := r.Context()

	chunkAddressFn := s.pinChunkAddressFn(ctx, addr)

	err = s.traversal.TraverseManifestAddresses(ctx, addr, chunkAddressFn)
	if err != nil {
		s.logger.Debugf("pin hop: traverse chunks: %v, addr %s", err, addr)

		if errors.Is(err, traversal.ErrInvalidType) {
			s.logger.Error("pin hop: invalid type")
			jsonhttp.BadRequest(w, "invalid type")
			return
		}

		s.logger.Error("pin hop: cannot pin")
		jsonhttp.InternalServerError(w, "cannot pin")
		return
	}

	jsonhttp.OK(w, nil)
}

// unpinHop removes pinning from content.
func (s *server) unpinHop(w http.ResponseWriter, r *http.Request) {
	addr, err := swarm.ParseHexAddress(mux.Vars(r)["address"])
	if err != nil {
		s.logger.Debugf("pin hop: parse address: %v", err)
		s.logger.Error("pin hop: parse address")
		jsonhttp.BadRequest(w, "bad address")
		return
	}

	has, err := s.storer.Has(r.Context(), addr)
	if err != nil {
		s.logger.Debugf("pin hop: localstore has: %v", err)
		s.logger.Error("pin hop: store")
		jsonhttp.InternalServerError(w, err)
		return
	}

	if !has {
		jsonhttp.NotFound(w, nil)
		return
	}

	ctx := r.Context()

	chunkAddressFn := s.unpinChunkAddressFn(ctx, addr)

	err = s.traversal.TraverseManifestAddresses(ctx, addr, chunkAddressFn)
	if err != nil {
		s.logger.Debugf("pin hop: traverse chunks: %v, addr %s", err, addr)

		if errors.Is(err, traversal.ErrInvalidType) {
			s.logger.Error("pin hop: invalid type")
			jsonhttp.BadRequest(w, "invalid type")
			return
		}

		s.logger.Error("pin hop: cannot unpin")
		jsonhttp.InternalServerError(w, "cannot unpin")
		return
	}

	jsonhttp.OK(w, nil)
}
