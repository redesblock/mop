package debugapi

import (
	"net/http"

	"github.com/redesblock/hop/core/jsonhttp"
)

func (s *Service) reserveStateHandler(w http.ResponseWriter, _ *http.Request) {
	jsonhttp.OK(w, s.batchStore.GetReserveState())
}

// chainStateHandler returns the current chain state.
func (s *Service) chainStateHandler(w http.ResponseWriter, _ *http.Request) {
	jsonhttp.OK(w, s.batchStore.GetChainState())
}
