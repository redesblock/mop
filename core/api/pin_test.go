package api_test

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/redesblock/mop/core/api"
	"github.com/redesblock/mop/core/flock"
	"github.com/redesblock/mop/core/jsonhttp"
	"github.com/redesblock/mop/core/jsonhttp/jsonhttptest"
	"github.com/redesblock/mop/core/logging"
	pinning "github.com/redesblock/mop/core/pinning/mock"
	mockpost "github.com/redesblock/mop/core/postage/mock"
	statestore "github.com/redesblock/mop/core/statestore/mock"
	"github.com/redesblock/mop/core/storage/mock"
	testingc "github.com/redesblock/mop/core/storage/testing"
	"github.com/redesblock/mop/core/tags"
	"github.com/redesblock/mop/core/traversal"
)

func checkPinHandlers(t *testing.T, client *http.Client, rootHash string, createPin bool) {
	t.Helper()

	const pinsBasePath = "/pins"

	var (
		pinsReferencePath        = pinsBasePath + "/" + rootHash
		pinsInvalidReferencePath = pinsBasePath + "/" + "838d0a193ecd1152d1bb1432d5ecc02398533b2494889e23b8bd5ace30ac2zzz"
		pinsUnknownReferencePath = pinsBasePath + "/" + "838d0a193ecd1152d1bb1432d5ecc02398533b2494889e23b8bd5ace30ac2ccc"
	)

	jsonhttptest.Request(t, client, http.MethodGet, pinsInvalidReferencePath, http.StatusBadRequest)

	jsonhttptest.Request(t, client, http.MethodGet, pinsUnknownReferencePath, http.StatusNotFound,
		jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
			Message: http.StatusText(http.StatusNotFound),
			Code:    http.StatusNotFound,
		}),
	)

	if createPin {
		jsonhttptest.Request(t, client, http.MethodPost, pinsReferencePath, http.StatusCreated,
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: http.StatusText(http.StatusCreated),
				Code:    http.StatusCreated,
			}),
		)
	}

	jsonhttptest.Request(t, client, http.MethodGet, pinsReferencePath, http.StatusOK,
		jsonhttptest.WithExpectedJSONResponse(struct {
			Reference flock.Address `json:"reference"`
		}{
			Reference: flock.MustParseHexAddress(rootHash),
		}),
	)

	jsonhttptest.Request(t, client, http.MethodGet, pinsBasePath, http.StatusOK,
		jsonhttptest.WithExpectedJSONResponse(struct {
			References []flock.Address `json:"references"`
		}{
			References: []flock.Address{flock.MustParseHexAddress(rootHash)},
		}),
	)

	jsonhttptest.Request(t, client, http.MethodDelete, pinsReferencePath, http.StatusOK)

	jsonhttptest.Request(t, client, http.MethodGet, pinsReferencePath, http.StatusNotFound,
		jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
			Message: http.StatusText(http.StatusNotFound),
			Code:    http.StatusNotFound,
		}),
	)
}

func TestPinHandlers(t *testing.T) {
	var (
		storerMock   = mock.NewStorer()
		client, _, _ = newTestServer(t, testServerOptions{
			Storer:    storerMock,
			Traversal: traversal.New(storerMock),
			Tags:      tags.NewTags(statestore.NewStateStore(), logging.New(io.Discard, 0)),
			Pinning:   pinning.NewServiceMock(),
			Logger:    logging.New(io.Discard, 5),
			Post:      mockpost.New(mockpost.WithAcceptAll()),
		})
	)

	t.Run("bytes", func(t *testing.T) {
		const rootHash = "838d0a193ecd1152d1bb1432d5ecc02398533b2494889e23b8bd5ace30ac2aeb"
		jsonhttptest.Request(t, client, http.MethodPost, "/bytes", http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.FlockPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(strings.NewReader("this is a simple text")),
			jsonhttptest.WithExpectedJSONResponse(api.MopUploadResponse{
				Reference: flock.MustParseHexAddress(rootHash),
			}),
		)
		checkPinHandlers(t, client, rootHash, true)
	})

	t.Run("mop", func(t *testing.T) {
		tarReader := tarFiles(t, []f{{
			data: []byte("<h1>Flock"),
			name: "index.html",
			dir:  "",
		}})
		rootHash := "350dd938021b8c68d6de9e23003e57219301061b6c0bb1a3c9ea537a8b246e4c"
		jsonhttptest.Request(t, client, http.MethodPost, "/mop", http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.FlockPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(tarReader),
			jsonhttptest.WithRequestHeader("Content-Type", api.ContentTypeTar),
			jsonhttptest.WithRequestHeader(api.FlockCollectionHeader, "true"),
			jsonhttptest.WithRequestHeader(api.FlockPinHeader, "true"),
			jsonhttptest.WithExpectedJSONResponse(api.MopUploadResponse{
				Reference: flock.MustParseHexAddress(rootHash),
			}),
		)
		checkPinHandlers(t, client, rootHash, false)

		header := jsonhttptest.Request(t, client, http.MethodPost, "/mop?name=somefile.txt", http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.FlockPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestHeader("Content-Type", "text/plain"),
			jsonhttptest.WithRequestHeader(api.FlockEncryptHeader, "true"),
			jsonhttptest.WithRequestHeader(api.FlockPinHeader, "true"),
			jsonhttptest.WithRequestBody(strings.NewReader("this is a simple text")),
		)

		rootHash = strings.Trim(header.Get("ETag"), "\"")
		checkPinHandlers(t, client, rootHash, false)
	})

	t.Run("chunk", func(t *testing.T) {
		var (
			chunk    = testingc.GenerateTestRandomChunk()
			rootHash = chunk.Address().String()
		)
		jsonhttptest.Request(t, client, http.MethodPost, "/chunks", http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.FlockPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(bytes.NewReader(chunk.Data())),
			jsonhttptest.WithExpectedJSONResponse(api.ChunkAddressResponse{
				Reference: chunk.Address(),
			}),
		)
		checkPinHandlers(t, client, rootHash, true)
	})
}
