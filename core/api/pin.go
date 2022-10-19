package api

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/storer/storage"
)

// pinRootHash pins root hash of given reference. This method is idempotent.
func (s *Service) pinRootHash(w http.ResponseWriter, r *http.Request) {
	ref, err := cluster.ParseHexAddress(mux.Vars(r)["reference"])
	if err != nil {
		s.logger.Debug("pins root hash: parse reference string failed", "string", mux.Vars(r)["reference"], "error", err)
		s.logger.Error(nil, "pins root hash: parse reference string failed")
		jsonhttp.BadRequest(w, "parse reference string failed")
		return
	}

	has, err := s.pinning.HasPin(ref)
	if err != nil {
		s.logger.Debug("pins root hash: has pins failed", "chunk_address", ref, "error", err)
		s.logger.Error(nil, "pins root hash: has pins failed")
		jsonhttp.InternalServerError(w, "pins root hash: checking of tracking pins failed")
		return
	}
	if has {
		jsonhttp.OK(w, nil)
		return
	}

	switch err = s.pinning.CreatePin(r.Context(), ref, true); {
	case errors.Is(err, storage.ErrNotFound):
		jsonhttp.NotFound(w, nil)
		return
	case err != nil:
		s.logger.Debug("pins root hash: create pins failed", "chunk_address", ref, "error", err)
		s.logger.Error(nil, "pins root hash: create pins failed")
		jsonhttp.InternalServerError(w, "pins root hash: creation of tracking pins failed")
		return
	}

	jsonhttp.Created(w, nil)
}

// unpinRootHash unpin's an already pinned root hash. This method is idempotent.
func (s *Service) unpinRootHash(w http.ResponseWriter, r *http.Request) {
	ref, err := cluster.ParseHexAddress(mux.Vars(r)["reference"])
	if err != nil {
		s.logger.Debug("unpin root hash: parse reference string failed", "string", mux.Vars(r)["reference"], "error", err)
		s.logger.Error(nil, "unpin root hash: parse reference string failed")
		jsonhttp.BadRequest(w, "parse reference string failed")
		return
	}

	has, err := s.pinning.HasPin(ref)
	if err != nil {
		s.logger.Debug("unpin root hash: has pins failed", "chunk_address", ref, "error", err)
		s.logger.Error(nil, "unpin root hash: has pins failed")
		jsonhttp.InternalServerError(w, "pins root hash: checking of tracking pins")
		return
	}
	if !has {
		jsonhttp.NotFound(w, nil)
		return
	}

	if err := s.pinning.DeletePin(r.Context(), ref); err != nil {
		s.logger.Debug("unpin root hash: delete pins failed", "chunk_address", ref, "error", err)
		s.logger.Error(nil, "unpin root hash: delete pins failed")
		jsonhttp.InternalServerError(w, "unpin root hash: deletion of pins failed")
		return
	}

	jsonhttp.OK(w, nil)
}

// getPinnedRootHash returns back the given reference if its root hash is pinned.
func (s *Service) getPinnedRootHash(w http.ResponseWriter, r *http.Request) {
	ref, err := cluster.ParseHexAddress(mux.Vars(r)["reference"])
	if err != nil {
		s.logger.Debug("pinned root hash: parse reference string failed", "string", mux.Vars(r)["reference"], "error", err)
		s.logger.Error(nil, "pinned root hash: parse reference string failed")
		jsonhttp.BadRequest(w, "parse reference string failed")
		return
	}

	has, err := s.pinning.HasPin(ref)
	if err != nil {
		s.logger.Debug("pinned root hash: has pins failed", "chunk_address", ref, "error", err)
		s.logger.Error(nil, "pinned root hash: has pins failed")
		jsonhttp.InternalServerError(w, "pinned root hash: check reference failed")
		return
	}

	if !has {
		jsonhttp.NotFound(w, nil)
		return
	}

	jsonhttp.OK(w, struct {
		Reference cluster.Address `json:"reference"`
	}{
		Reference: ref,
	})
}

// listPinnedRootHashes lists all the references of the pinned root hashes.
func (s *Service) listPinnedRootHashes(w http.ResponseWriter, r *http.Request) {
	pinned, err := s.pinning.Pins()
	if err != nil {
		s.logger.Debug("list pinned root references: unable to list references", "error", err)
		s.logger.Error(nil, "list pinned root references: unable to list references")
		jsonhttp.InternalServerError(w, "list pinned root references failed")
		return
	}

	jsonhttp.OK(w, struct {
		References []cluster.Address `json:"references"`
	}{
		References: pinned,
	})
}
