package debugapi

import (
	"net/http"

	"github.com/redesblock/hop/core/jsonhttp"
)

func (s *Service) reserveStateHandler(w http.ResponseWriter, r *http.Request) {
	jsonhttp.OK(w, s.batchStore.GetReserveState())
}
