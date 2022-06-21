package api_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/redesblock/hop/core/api"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/jsonhttp/jsonhttptest"
	"github.com/redesblock/hop/core/logging"
	pinning "github.com/redesblock/hop/core/pinning/mock"
	mockpost "github.com/redesblock/hop/core/postage/mock"
	statestore "github.com/redesblock/hop/core/statestore/mock"
	"github.com/redesblock/hop/core/storage/mock"
	testingc "github.com/redesblock/hop/core/storage/testing"
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/tags"
	"github.com/redesblock/hop/core/traversal"
)

func checkPinHandlers(t *testing.T, client *http.Client, rootHash string) {
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

	jsonhttptest.Request(t, client, http.MethodPost, pinsReferencePath, http.StatusCreated,
		jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
			Message: http.StatusText(http.StatusCreated),
			Code:    http.StatusCreated,
		}),
	)

	jsonhttptest.Request(t, client, http.MethodGet, pinsReferencePath, http.StatusOK,
		jsonhttptest.WithExpectedJSONResponse(struct {
			Reference swarm.Address `json:"reference"`
		}{
			Reference: swarm.MustParseHexAddress(rootHash),
		}),
	)

	jsonhttptest.Request(t, client, http.MethodGet, pinsBasePath, http.StatusOK,
		jsonhttptest.WithExpectedJSONResponse(struct {
			References []swarm.Address `json:"references"`
		}{
			References: []swarm.Address{swarm.MustParseHexAddress(rootHash)},
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
			Traversal: traversal.NewService(storerMock),
			Tags:      tags.NewTags(statestore.NewStateStore(), logging.New(ioutil.Discard, 0)),
			Pinning:   pinning.NewServiceMock(),
			Logger:    logging.New(ioutil.Discard, 5),
			Post:      mockpost.New(mockpost.WithAcceptAll()),
		})
	)

	t.Run("bytes", func(t *testing.T) {
		const rootHash = "838d0a193ecd1152d1bb1432d5ecc02398533b2494889e23b8bd5ace30ac2aeb"
		jsonhttptest.Request(t, client, http.MethodPost, "/bytes", http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.SwarmPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(strings.NewReader("this is a simple text")),
			jsonhttptest.WithExpectedJSONResponse(api.HopUploadResponse{
				Reference: swarm.MustParseHexAddress(rootHash),
			}),
		)
		checkPinHandlers(t, client, rootHash)
	})

	t.Run("hop", func(t *testing.T) {
		tarReader := tarFiles(t, []f{{
			data: []byte("<h1>Swarm"),
			name: "index.html",
			dir:  "",
		}})
		rootHash := "9e178dbd1ed4b748379e25144e28dfb29c07a4b5114896ef454480115a56b237"
		jsonhttptest.Request(t, client, http.MethodPost, "/hop", http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.SwarmPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(tarReader),
			jsonhttptest.WithRequestHeader("Content-Type", api.ContentTypeTar),
			jsonhttptest.WithRequestHeader(api.SwarmCollectionHeader, "True"),
			jsonhttptest.WithExpectedJSONResponse(api.HopUploadResponse{
				Reference: swarm.MustParseHexAddress(rootHash),
			}),
		)
		checkPinHandlers(t, client, rootHash)

		rootHash = "dd13a5a6cc9db3ef514d645e6719178dbfb1a90b49b9262cafce35b0d27cf245"
		jsonhttptest.Request(t, client, http.MethodPost, "/hop?name=somefile.txt", http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.SwarmPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestHeader("Content-Type", "text/plain"),
			jsonhttptest.WithRequestBody(strings.NewReader("this is a simple text")),
			jsonhttptest.WithExpectedJSONResponse(api.HopUploadResponse{
				Reference: swarm.MustParseHexAddress(rootHash),
			}),
		)
		checkPinHandlers(t, client, rootHash)
	})

	t.Run("chunk", func(t *testing.T) {
		var (
			chunk    = testingc.GenerateTestRandomChunk()
			rootHash = chunk.Address().String()
		)
		jsonhttptest.Request(t, client, http.MethodPost, "/chunks", http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.SwarmPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(bytes.NewReader(chunk.Data())),
			jsonhttptest.WithExpectedJSONResponse(api.ChunkAddressResponse{
				Reference: chunk.Address(),
			}),
		)
		checkPinHandlers(t, client, rootHash)
	})
}
