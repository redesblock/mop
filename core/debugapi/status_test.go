package debugapi_test

import (
	"github.com/redesblock/hop/cmd/version"
	"net/http"
	"testing"

	"github.com/redesblock/hop/core/api"
	"github.com/redesblock/hop/core/debugapi"
	"github.com/redesblock/hop/core/jsonhttp/jsonhttptest"
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
