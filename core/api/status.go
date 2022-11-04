package api

import (
	"net/http"

	ver "github.com/redesblock/mop"
	"github.com/redesblock/mop/core/api/jsonhttp"
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
		Version:         ver.Version,
		APIVersion:      Version,
		DebugAPIVersion: Version,
	})
}
