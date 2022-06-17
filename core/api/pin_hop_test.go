package api_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/ethersphere/manifest/mantaray"
	"github.com/redesblock/hop/core/api"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/jsonhttp/jsonhttptest"
	"github.com/redesblock/hop/core/logging"
	statestore "github.com/redesblock/hop/core/statestore/mock"
	"github.com/redesblock/hop/core/storage/mock"
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/tags"
	"github.com/redesblock/hop/core/traversal"
)

func TestPinHopHandler(t *testing.T) {
	var (
		dirUploadResource     = "/dirs"
		pinHopResource        = "/pin/hop"
		pinHopAddressResource = func(addr string) string { return pinHopResource + "/" + addr }
		pinChunksResource     = "/pin/chunks"

		mockStorer       = mock.NewStorer()
		mockStatestore   = statestore.NewStateStore()
		traversalService = traversal.NewService(mockStorer)
		logger           = logging.New(ioutil.Discard, 0)
		client, _, _     = newTestServer(t, testServerOptions{
			Storer:    mockStorer,
			Traversal: traversalService,
			Tags:      tags.NewTags(mockStatestore, logger),
		})
	)

	var (
		obfuscationKey   = make([]byte, 32)
		obfuscationKeyFn = func(p []byte) (n int, err error) {
			n = copy(p, obfuscationKey)
			return
		}
	)
	mantaray.SetObfuscationKeyFn(obfuscationKeyFn)

	t.Run("pin-hop-1", func(t *testing.T) {
		files := []f{
			{
				data: []byte("<h1>Swarm"),
				name: "index.html",
				dir:  "",
			},
		}

		tarReader := tarFiles(t, files)

		rootHash := "efc4c4cb45f346416eaad92bc0a34c7a92fc042c2cdd8f713345c5fadb235706"

		// verify directory tar upload response
		jsonhttptest.Request(t, client, http.MethodPost, dirUploadResource, http.StatusOK,
			jsonhttptest.WithRequestBody(tarReader),
			jsonhttptest.WithRequestHeader("Content-Type", api.ContentTypeTar),
			jsonhttptest.WithExpectedJSONResponse(api.FileUploadResponse{
				Reference: swarm.MustParseHexAddress(rootHash),
			}),
		)

		jsonhttptest.Request(t, client, http.MethodPost, pinHopAddressResource(rootHash), http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: http.StatusText(http.StatusOK),
				Code:    http.StatusOK,
			}),
		)

		expectedChunkCount := 7

		var respBytes []byte

		jsonhttptest.Request(t, client, http.MethodGet, pinChunksResource, http.StatusOK,
			jsonhttptest.WithPutResponseBody(&respBytes),
		)

		read := bytes.NewReader(respBytes)

		// get the reference as everytime it will change because of random encryption key
		var resp api.ListPinnedChunksResponse
		err := json.NewDecoder(read).Decode(&resp)
		if err != nil {
			t.Fatal(err)
		}

		if expectedChunkCount != len(resp.Chunks) {
			t.Fatalf("expected to find %d pinned chunks, got %d", expectedChunkCount, len(resp.Chunks))
		}
	})

	t.Run("unpin-hop-1", func(t *testing.T) {
		rootHash := "efc4c4cb45f346416eaad92bc0a34c7a92fc042c2cdd8f713345c5fadb235706"

		jsonhttptest.Request(t, client, http.MethodDelete, pinHopAddressResource(rootHash), http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: http.StatusText(http.StatusOK),
				Code:    http.StatusOK,
			}),
		)

		jsonhttptest.Request(t, client, http.MethodGet, pinChunksResource, http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(api.ListPinnedChunksResponse{
				Chunks: []api.PinnedChunk{},
			}),
		)
	})

}
