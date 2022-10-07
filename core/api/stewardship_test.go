package api_test

import (
	"encoding/hex"
	"io"
	"net/http"
	"testing"

	"github.com/redesblock/mop/core/api"
	"github.com/redesblock/mop/core/flock"
	"github.com/redesblock/mop/core/jsonhttp"
	"github.com/redesblock/mop/core/jsonhttp/jsonhttptest"
	"github.com/redesblock/mop/core/logging"
	statestore "github.com/redesblock/mop/core/statestore/mock"
	"github.com/redesblock/mop/core/steward/mock"
	smock "github.com/redesblock/mop/core/storage/mock"
	"github.com/redesblock/mop/core/tags"
)

func TestStewardship(t *testing.T) {
	var (
		logger         = logging.New(io.Discard, 0)
		statestoreMock = statestore.NewStateStore()
		stewardMock    = &mock.Steward{}
		storer         = smock.NewStorer()
		addr           = flock.NewAddress([]byte{31: 128})
	)
	client, _, _ := newTestServer(t, testServerOptions{
		Storer:  storer,
		Tags:    tags.NewTags(statestoreMock, logger),
		Logger:  logger,
		Steward: stewardMock,
	})

	t.Run("re-upload", func(t *testing.T) {
		jsonhttptest.Request(t, client, http.MethodPut, "/v1/stewardship/"+addr.String(), http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: http.StatusText(http.StatusOK),
				Code:    http.StatusOK,
			}),
		)
		if !stewardMock.LastAddress().Equal(addr) {
			t.Fatalf("\nhave address: %q\nwant address: %q", stewardMock.LastAddress().String(), addr.String())
		}
	})

	t.Run("is-retrievable", func(t *testing.T) {
		jsonhttptest.Request(t, client, http.MethodGet, "/v1/stewardship/"+addr.String(), http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(api.IsRetrievableResponse{IsRetrievable: true}),
		)
		jsonhttptest.Request(t, client, http.MethodGet, "/v1/stewardship/"+hex.EncodeToString([]byte{}), http.StatusNotFound,
			jsonhttptest.WithExpectedJSONResponse(&jsonhttp.StatusResponse{
				Code:    http.StatusNotFound,
				Message: http.StatusText(http.StatusNotFound),
			}),
		)
	})
}
