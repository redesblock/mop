package api_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/redesblock/hop/core/api"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/jsonhttp/jsonhttptest"
	"github.com/redesblock/hop/core/storage/mock"
	"github.com/redesblock/hop/core/swarm"
)

// TestHop tests that the data upload api responds as expected when uploading,
// downloading and requesting a resource that cannot be found.
func TestHop(t *testing.T) {
	var (
		resource   = "/hop"
		content    = []byte("foo")
		expHash    = "2387e8e7d8a48c2a9339c97c1dc3461a9a7aa07e994c5cb8b38fd7c1b3e6ea48"
		mockStorer = mock.NewStorer()
		client     = newTestServer(t, testServerOptions{
			Storer: mockStorer,
		})
	)

	t.Run("upload", func(t *testing.T) {
		jsonhttptest.ResponseDirect(t, client, http.MethodPost, resource, bytes.NewReader(content), http.StatusOK, api.HopPostResponse{
			Hash: swarm.MustParseHexAddress(expHash),
		})
	})

	t.Run("download", func(t *testing.T) {
		resp := request(t, client, http.MethodGet, resource+"/"+expHash, nil, http.StatusOK)
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(data, content) {
			t.Fatalf("data mismatch. got %s, want %s", string(data), string(content))
		}
	})

	t.Run("not found", func(t *testing.T) {
		jsonhttptest.ResponseDirect(t, client, http.MethodGet, resource+"/abcd", nil, http.StatusNotFound, jsonhttp.StatusResponse{
			Message: "not found",
			Code:    http.StatusNotFound,
		})
	})
}
