package debugapi_test

import (
	"net/http"
	"testing"

	"github.com/redesblock/hop/core/debugapi"
	"github.com/redesblock/hop/core/jsonhttp/jsonhttptest"
)

func TestHealth(t *testing.T) {
	testServer := newTestServer(t, testServerOptions{})

	jsonhttptest.Request(t, testServer.Client, http.MethodGet, "/health", http.StatusOK,
		jsonhttptest.WithExpectedJSONResponse(debugapi.StatusResponse{
			Status: "ok",
		}),
	)
}

func TestReadiness(t *testing.T) {
	testServer := newTestServer(t, testServerOptions{})

	jsonhttptest.Request(t, testServer.Client, http.MethodGet, "/readiness", http.StatusOK,
		jsonhttptest.WithExpectedJSONResponse(debugapi.StatusResponse{
			Status: "ok",
		}),
	)
}
