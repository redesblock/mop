package debugapi

import (
	"github.com/redesblock/mop/cmd/version"
	"net/http"

	"github.com/redesblock/mop/core/api"
	"github.com/redesblock/mop/core/jsonhttp"
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
