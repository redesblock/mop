package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redesblock/hop/core/jsonhttp"
)

//  stewardshipPutHandler re-uploads root hash and all of its underlying
// associated chunks to the network.
func (s *server) stewardshipPutHandler(w http.ResponseWriter, r *http.Request) {
	nameOrHex := mux.Vars(r)["address"]
	address, err := s.resolveNameOrAddress(nameOrHex)
	if err != nil {
		s.logger.Debugf("stewardship put: parse address %s: %v", nameOrHex, err)
		s.logger.Error("stewardship put: parse address")
		jsonhttp.NotFound(w, nil)
		return
	}
	err = s.steward.Reupload(r.Context(), address)
	if err != nil {
		s.logger.Debugf("stewardship put: re-upload %s: %v", address, err)
		s.logger.Error("stewardship put: re-upload")
		jsonhttp.InternalServerError(w, nil)
		return
	}
	jsonhttp.OK(w, nil)
}
