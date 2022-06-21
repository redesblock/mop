package api_test

import (
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

func TestPinHopHandler(t *testing.T) {
	var (
		dirUploadResource     = "/hop"
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
			Logger:    logger,
		})
	)

	t.Run("pin-hop-1", func(t *testing.T) {
		files := []f{
			{
				data: []byte("<h1>Swarm"),
				name: "index.html",
				dir:  "",
			},
		}

		tarReader := tarFiles(t, files)

		rootHash := "9e178dbd1ed4b748379e25144e28dfb29c07a4b5114896ef454480115a56b237"

		// verify directory tar upload response
		jsonhttptest.Request(t, client, http.MethodPost, dirUploadResource, http.StatusOK,
			jsonhttptest.WithRequestBody(tarReader),
			jsonhttptest.WithRequestHeader("Content-Type", api.ContentTypeTar),
			jsonhttptest.WithRequestHeader(api.SwarmCollectionHeader, "True"),
			jsonhttptest.WithExpectedJSONResponse(api.HopUploadResponse{
				Reference: swarm.MustParseHexAddress(rootHash),
			}),
		)

		jsonhttptest.Request(t, client, http.MethodPost, pinHopAddressResource(rootHash), http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: http.StatusText(http.StatusOK),
				Code:    http.StatusOK,
			}),
		)

		expectedChunkCount := 3

		// get the reference as everytime it will change because of random encryption key
		var resp api.ListPinnedChunksResponse

		jsonhttptest.Request(t, client, http.MethodGet, pinChunksResource, http.StatusOK,
			jsonhttptest.WithUnmarshalJSONResponse(&resp),
		)

		if expectedChunkCount != len(resp.Chunks) {
			t.Fatalf("expected to find %d pinned chunks, got %d", expectedChunkCount, len(resp.Chunks))
		}
	})

	t.Run("unpin-hop-1", func(t *testing.T) {
		rootHash := "9e178dbd1ed4b748379e25144e28dfb29c07a4b5114896ef454480115a56b237"

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
