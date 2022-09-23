package debugapi_test

import (
	"github.com/redesblock/mop/cmd/version"
	"net/http"
	"testing"

	"github.com/redesblock/mop/core/api"
	"github.com/redesblock/mop/core/debugapi"
	"github.com/redesblock/mop/core/jsonhttp/jsonhttptest"
)

func TestHealth(t *testing.T) {
	testServer := newTestServer(t, testServerOptions{})

	jsonhttptest.Request(t, testServer.Client, http.MethodGet, "/health", http.StatusOK,
		jsonhttptest.WithExpectedJSONResponse(debugapi.StatusResponse{
			Status:          "ok",
			Version:         version.Version,
			APIVersion:      api.Version,
			DebugAPIVersion: debugapi.Version,
		}),
	)
}

func TestReadiness(t *testing.T) {
	testServer := newTestServer(t, testServerOptions{})

	jsonhttptest.Request(t, testServer.Client, http.MethodGet, "/readiness", http.StatusOK,
		jsonhttptest.WithExpectedJSONResponse(debugapi.StatusResponse{
			Status:          "ok",
			Version:         version.Version,
			APIVersion:      api.Version,
			DebugAPIVersion: debugapi.Version,
		}),
	)
}
