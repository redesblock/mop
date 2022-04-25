package debugapi_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/jsonhttp/jsonhttptest"
	topmock "github.com/redesblock/hop/core/topology/mock"
)

type topologyResponse struct {
	Topology string `json:"topology"`
}

func TestTopologyOK(t *testing.T) {
	marshalFunc := func() ([]byte, error) {
		return json.Marshal(topologyResponse{Topology: "abcd"})
	}
	testServer := newTestServer(t, testServerOptions{
		TopologyOpts: []topmock.Option{topmock.WithMarshalJSONFunc(marshalFunc)},
	})

	jsonhttptest.ResponseDirect(t, testServer.Client, http.MethodGet, "/topology", nil, http.StatusOK, topologyResponse{
		Topology: "abcd",
	})
}

func TestTopologyError(t *testing.T) {
	marshalFunc := func() ([]byte, error) {
		return nil, errors.New("error")
	}
	testServer := newTestServer(t, testServerOptions{
		TopologyOpts: []topmock.Option{topmock.WithMarshalJSONFunc(marshalFunc)},
	})

	jsonhttptest.ResponseDirect(t, testServer.Client, http.MethodGet, "/topology", nil, http.StatusInternalServerError, jsonhttp.StatusResponse{
		Message: "error",
		Code:    http.StatusInternalServerError,
	})
}