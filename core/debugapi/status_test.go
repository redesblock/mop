package debugapi_test

import (
	"net/http"
	"testing"

	"github.com/redesblock/hop/core/debugapi"
	"github.com/redesblock/hop/core/jsonhttp/jsonhttptest"
)

func TestHealth(t *testing.T) {
	testServer := newTestServer(t, testServerOptions{})

	jsonhttptest.ResponseDirect(t, testServer.Client, http.MethodGet, "/health", nil, http.StatusOK, debugapi.StatusResponse{
		Status: "ok",
	})
}

func TestReadiness(t *testing.T) {
	testServer := newTestServer(t, testServerOptions{})

	jsonhttptest.ResponseDirect(t, testServer.Client, http.MethodGet, "/readiness", nil, http.StatusOK, debugapi.StatusResponse{
		Status: "ok",
	})
}
