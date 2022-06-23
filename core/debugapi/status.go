package debugapi

import (
	"github.com/redesblock/hop/cmd/version"
	"net/http"

	"github.com/redesblock/hop/core/api"
	"github.com/redesblock/hop/core/jsonhttp"
)

type statusResponse struct {
	Status          string `json:"status"`
	Version         string `json:"version"`
	APIVersion      string `json:"apiVersion"`
	DebugAPIVersion string `json:"debugApiVersion"`
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	jsonhttp.OK(w, statusResponse{
		Status:          "ok",
		Version:         version.Version,
		APIVersion:      api.Version,
		DebugAPIVersion: Version,
	})
}
