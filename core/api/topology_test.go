package api_test

import (
	"net/http"
	"testing"

	"github.com/redesblock/mop/core/api/jsonhttp/jsonhttptest"
)

func TestTopologyOK(t *testing.T) {
	testServer, _, _, _ := newTestServer(t, testServerOptions{DebugAPI: true})

	var body []byte
	opts := jsonhttptest.WithPutResponseBody(&body)
	jsonhttptest.Request(t, testServer, http.MethodGet, "/topology", http.StatusOK, opts)

	if len(body) == 0 {
		t.Error("empty response")
	}
}
