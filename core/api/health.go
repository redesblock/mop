package api

import (
	"net/http"
	"runtime/debug"

	ver "github.com/redesblock/mop"
	"github.com/redesblock/mop/core/api/jsonhttp"
)

type statusResponse struct {
	Status          string `json:"status"`
	Version         string `json:"version"`
	APIVersion      string `json:"apiVersion"`
	DebugAPIVersion string `json:"debugApiVersion"`
}

func (s *Service) healthHandler(w http.ResponseWriter, r *http.Request) {
	status := s.probe.Healthy()
	jsonhttp.OK(w, statusResponse{
		Status:          status.String(),
		Version:         ver.Version,
		APIVersion:      Version,
		DebugAPIVersion: Version,
	})
}

func (s *Service) freeHandler(w http.ResponseWriter, r *http.Request) {
	debug.FreeOSMemory()
	jsonhttp.OK(w, "ok")
}
