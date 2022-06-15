package debugapi

import (
	"net/http"

	"github.com/redesblock/hop/core/jsonhttp"
)

type statusResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func (s *server) statusHandler(w http.ResponseWriter, r *http.Request) {
	jsonhttp.OK(w, statusResponse{
		Status: "ok",
	})
}
