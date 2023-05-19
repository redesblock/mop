package api

import (
	"errors"
	"net/http"

	"github.com/redesblock/mop/core/resolver"

	"github.com/gorilla/mux"
	"github.com/redesblock/mop/core/api/jsonhttp"
)

//	wardenshipPutHandler re-uploads root hash and all of its underlying
//
// associated chunks to the network.
func (s *Service) wardenshipPutHandler(w http.ResponseWriter, r *http.Request) {
	nameOrHex := mux.Vars(r)["address"]
	address, err := s.resolveNameOrAddress(nameOrHex)
	switch {
	case errors.Is(err, resolver.ErrParse), errors.Is(err, resolver.ErrInvalidContentHash):
		s.logger.Debug("wardenship put: parse address string failed", "string", nameOrHex, "error", err)
		s.logger.Error(nil, "wardenship put: invalid address")
		jsonhttp.BadRequest(w, "invalid address")
		return
	case errors.Is(err, resolver.ErrNotFound):
		s.logger.Debug("wardenship put: address not found", "string", nameOrHex, "error", err)
		s.logger.Error(nil, "wardenship put: address not found")
		jsonhttp.NotFound(w, "address not found")
		return
	case errors.Is(err, resolver.ErrServiceNotAvailable):
		s.logger.Debug("wardenship put: service unavailable", "string", nameOrHex, "error", err)
		s.logger.Error(nil, "wardenship put: service unavailable")
		jsonhttp.InternalServerError(w, "wardenship put: resolver service unavailable")
		return
	case err != nil:
		s.logger.Debug("wardenship put: resolve address or name string failed", "string", nameOrHex, "error", err)
		s.logger.Error(nil, "wardenship put: resolve address or name string failed")
		jsonhttp.InternalServerError(w, "wardenship put: resolve name or address")
		return
	}
	err = s.warden.Reupload(r.Context(), address)
	if err != nil {
		s.logger.Debug("wardenship put: re-upload failed", "chunk_address", address, "error", err)
		s.logger.Error(nil, "wardenship put: re-upload failed")
		jsonhttp.InternalServerError(w, "wardenship put: re-upload failed")
		return
	}
	jsonhttp.OK(w, nil)
}

type isRetrievableResponse struct {
	IsRetrievable bool `json:"isRetrievable"`
}

// wardenshipGetHandler checks whether the content on the given address is retrievable.
func (s *Service) wardenshipGetHandler(w http.ResponseWriter, r *http.Request) {
	nameOrHex := mux.Vars(r)["address"]
	address, err := s.resolveNameOrAddress(nameOrHex)
	if err != nil {
		s.logger.Debug("wardenship get: parse address string failed", "string", nameOrHex, "error", err)
		s.logger.Error(nil, "wardenship get: parse address string failed")
		jsonhttp.NotFound(w, nil)
		return
	}
	res, err := s.warden.IsRetrievable(r.Context(), address)
	if err != nil {
		s.logger.Debug("wardenship get: is retrievable check failed", "chunk_address", address, "error", err)
		s.logger.Error(nil, "wardenship get: is retrievable")
		jsonhttp.InternalServerError(w, "wardenship get: is retrievable check failed")
		return
	}
	jsonhttp.OK(w, isRetrievableResponse{
		IsRetrievable: res,
	})
}
