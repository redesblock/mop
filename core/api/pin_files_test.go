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
	statestore "github.com/redesblock/hop/core/statestore/mock"
	"github.com/redesblock/hop/core/storage/mock"
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/tags"
	"github.com/redesblock/hop/core/traversal"
)

func TestPinFilesHandler(t *testing.T) {
	var (
		fileUploadResource      = "/hop"
		pinFilesResource        = "/pin/files"
		pinFilesAddressResource = func(addr string) string { return pinFilesResource + "/" + addr }
		pinChunksResource       = "/pin/chunks"

		simpleData = []byte("this is a simple text")

		mockStorer       = mock.NewStorer()
		mockStatestore   = statestore.NewStateStore()
		traversalService = traversal.NewService(mockStorer)
		logger           = logging.New(ioutil.Discard, 0)
		client, _, _     = newTestServer(t, testServerOptions{
			Storer:    mockStorer,
			Traversal: traversalService,
			Tags:      tags.NewTags(mockStatestore, logger),
			Logger:    logger,
		})
	)

	t.Run("pin-file-1", func(t *testing.T) {
		rootHash := "dd13a5a6cc9db3ef514d645e6719178dbfb1a90b49b9262cafce35b0d27cf245"
		metadataHash := "0cc878d32c96126d47f63fbe391114ee1438cd521146fc975dea1546d302b6c0"
		metadataHash2 := "a14d1ef845307c634e9ec74539bd668d0d1b37f37de4128939d57098135850da"
		contentHash := "838d0a193ecd1152d1bb1432d5ecc02398533b2494889e23b8bd5ace30ac2aeb"

		jsonhttptest.Request(t, client, http.MethodPost,
			fileUploadResource+"?name=somefile.txt", http.StatusOK,
			jsonhttptest.WithRequestBody(bytes.NewReader(simpleData)),
			jsonhttptest.WithExpectedJSONResponse(api.HopUploadResponse{
				Reference: swarm.MustParseHexAddress(rootHash),
			}),
			jsonhttptest.WithRequestHeader("Content-Type", "text/plain"),
		)

		jsonhttptest.Request(t, client, http.MethodPost, pinFilesAddressResource(rootHash), http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: http.StatusText(http.StatusOK),
				Code:    http.StatusOK,
			}),
		)

		hashes := map[string]int{
			rootHash:      1,
			metadataHash:  1,
			metadataHash2: 1,
			contentHash:   1,
		}

		actualResponse := api.ListPinnedChunksResponse{
			Chunks: []api.PinnedChunk{},
		}

		jsonhttptest.Request(t, client, http.MethodGet, pinChunksResource, http.StatusOK,
			jsonhttptest.WithUnmarshalJSONResponse(&actualResponse),
		)
		if len(actualResponse.Chunks) != len(hashes) {
			t.Fatalf("Response chunk count mismatch Expected: %d Found: %d",
				len(hashes), len(actualResponse.Chunks))
		}
		for _, v := range actualResponse.Chunks {
			if counter, ok := hashes[v.Address.String()]; !ok {
				t.Fatalf("found unexpected hash %s", v.Address.String())
			} else if uint64(counter) != v.PinCounter {
				t.Fatalf("found unexpected pin counter: Expected: %d, Found: %d",
					counter, v.PinCounter)
			}
		}
	})

	t.Run("unpin-file-1", func(t *testing.T) {
		rootHash := "dd13a5a6cc9db3ef514d645e6719178dbfb1a90b49b9262cafce35b0d27cf245"

		jsonhttptest.Request(t, client, http.MethodDelete, pinFilesAddressResource(rootHash), http.StatusOK,
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
