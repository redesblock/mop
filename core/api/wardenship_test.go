package api_test

import (
	"encoding/hex"
	"net/http"
	"testing"

	"github.com/redesblock/mop/core/api"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/api/jsonhttp/jsonhttptest"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/resolver"
	resolverMock "github.com/redesblock/mop/core/resolver/mock"
	statestore "github.com/redesblock/mop/core/storer/statestore/mock"
	smock "github.com/redesblock/mop/core/storer/storage/mock"
	"github.com/redesblock/mop/core/tags"
	"github.com/redesblock/mop/core/warden/mock"
)

func Testwardenship(t *testing.T) {
	var (
		logger         = log.Noop
		statestoreMock = statestore.NewStateStore()
		stewardMock    = &mock.Steward{}
		storer         = smock.NewStorer()
		addr           = cluster.NewAddress([]byte{31: 128})
	)
	client, _, _, _ := newTestServer(t, testServerOptions{
		Storer:  storer,
		Tags:    tags.NewTags(statestoreMock, logger),
		Logger:  logger,
		Steward: stewardMock,
	})

	t.Run("re-upload", func(t *testing.T) {
		jsonhttptest.Request(t, client, http.MethodPut, "/v1/wardenship/"+addr.String(), http.StatusOK,
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
		jsonhttptest.Request(t, client, http.MethodGet, "/v1/wardenship/"+addr.String(), http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(api.IsRetrievableResponse{IsRetrievable: true}),
		)
		jsonhttptest.Request(t, client, http.MethodGet, "/v1/wardenship/"+hex.EncodeToString([]byte{}), http.StatusNotFound,
			jsonhttptest.WithExpectedJSONResponse(&jsonhttp.StatusResponse{
				Code:    http.StatusNotFound,
				Message: http.StatusText(http.StatusNotFound),
			}),
		)
	})
}

func TestwardenshipInputValidations(t *testing.T) {
	var (
		logger         = log.Noop
		statestoreMock = statestore.NewStateStore()
		stewardMock    = &mock.Steward{}
		storer         = smock.NewStorer()
	)
	client, _, _, _ := newTestServer(t, testServerOptions{
		Storer:  storer,
		Tags:    tags.NewTags(statestoreMock, logger),
		Logger:  logger,
		Steward: stewardMock,
		Resolver: resolverMock.NewResolver(
			resolverMock.WithResolveFunc(
				func(string) (cluster.Address, error) {
					return cluster.Address{}, resolver.ErrParse
				},
			),
		),
	})
	for _, tt := range []struct {
		name            string
		reference       string
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:            "correct reference",
			reference:       "1e477b015af480e387fbf5edd90f1685a30c0e3ba88eeb3871b326b816a542da",
			expectedStatus:  http.StatusOK,
			expectedMessage: http.StatusText(http.StatusOK),
		},
		{
			name:            "reference not found",
			reference:       "1e477b015af480e387fbf5edd90f1685a30c0e3ba88eeb3871b326b816a542d/",
			expectedStatus:  http.StatusNotFound,
			expectedMessage: http.StatusText(http.StatusNotFound),
		},
		{
			name:            "incorrect reference",
			reference:       "xc0f6",
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "invalid address",
		},
	} {
		t.Run("input validation -"+tt.name, func(t *testing.T) {
			jsonhttptest.Request(t, client, http.MethodPut, "/v1/wardenship/"+tt.reference, tt.expectedStatus,
				jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
					Message: tt.expectedMessage,
					Code:    tt.expectedStatus,
				}),
			)
		})
	}
}
