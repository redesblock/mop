package api_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/redesblock/hop/core/api"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/jsonhttp/jsonhttptest"
	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/storage/mock"
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/tags"
	mockbytes "gitlab.com/nolash/go-mockbytes"
)

// TestBytes tests that the data upload api responds as expected when uploading,
// downloading and requesting a resource that cannot be found.
func TestBytes(t *testing.T) {
	var (
		resource   = "/bytes"
		targets    = "0x222"
		expHash    = "29a5fb121ce96194ba8b7b823a1f9c6af87e1791f824940a53b5a7efe3f790d9"
		mockStorer = mock.NewStorer()
		client     = newTestServer(t, testServerOptions{
			Storer: mockStorer,
			Tags:   tags.NewTags(),
			Logger: logging.New(ioutil.Discard, 5),
		})
	)
	g := mockbytes.New(0, mockbytes.MockTypeStandard).WithModulus(255)
	content, err := g.SequentialBytes(swarm.ChunkSize * 2)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("upload", func(t *testing.T) {
		jsonhttptest.ResponseDirect(t, client, http.MethodPost, resource, bytes.NewReader(content), http.StatusOK, api.BytesPostResponse{
			Reference: swarm.MustParseHexAddress(expHash),
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

	t.Run("download-with-targets", func(t *testing.T) {
		resp := request(t, client, http.MethodGet, resource+"/"+expHash+"?targets="+targets, nil, http.StatusOK)

		if resp.Header.Get(api.TargetsRecoveryHeader) != targets {
			t.Fatalf("targets mismatch. got %s, want %s", resp.Header.Get(api.TargetsRecoveryHeader), targets)
		}
	})

	t.Run("not found", func(t *testing.T) {
		jsonhttptest.ResponseDirect(t, client, http.MethodGet, resource+"/abcd", nil, http.StatusNotFound, jsonhttp.StatusResponse{
			Message: "not found",
			Code:    http.StatusNotFound,
		})
	})
}
