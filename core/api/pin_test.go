package api_test

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/redesblock/mop/core/api"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/api/jsonhttp/jsonhttptest"
	"github.com/redesblock/mop/core/cluster"
	mockpost "github.com/redesblock/mop/core/incentives/voucher/mock"
	"github.com/redesblock/mop/core/log"
	pinning "github.com/redesblock/mop/core/pins/mock"
	statestore "github.com/redesblock/mop/core/storer/statestore/mock"
	"github.com/redesblock/mop/core/storer/storage/mock"
	testingc "github.com/redesblock/mop/core/storer/storage/testing"
	"github.com/redesblock/mop/core/tags"
	"github.com/redesblock/mop/core/traverser"
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
			Reference cluster.Address `json:"reference"`
		}{
			Reference: cluster.MustParseHexAddress(rootHash),
		}),
	)

	jsonhttptest.Request(t, client, http.MethodGet, pinsBasePath, http.StatusOK,
		jsonhttptest.WithExpectedJSONResponse(struct {
			References []cluster.Address `json:"references"`
		}{
			References: []cluster.Address{cluster.MustParseHexAddress(rootHash)},
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
		logger          = log.Noop
		storerMock      = mock.NewStorer()
		client, _, _, _ = newTestServer(t, testServerOptions{
			Storer:    storerMock,
			Traversal: traverser.New(storerMock),
			Tags:      tags.NewTags(statestore.NewStateStore(), logger),
			Pinning:   pinning.NewServiceMock(),
			Logger:    logger,
			Post:      mockpost.New(mockpost.WithAcceptAll()),
		})
	)

	t.Run("bytes", func(t *testing.T) {
		const rootHash = "838d0a193ecd1152d1bb1432d5ecc02398533b2494889e23b8bd5ace30ac2aeb"
		jsonhttptest.Request(t, client, http.MethodPost, "/bytes", http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.ClusterDeferredUploadHeader, "true"),
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(strings.NewReader("this is a simple text")),
			jsonhttptest.WithExpectedJSONResponse(api.MopUploadResponse{
				Reference: cluster.MustParseHexAddress(rootHash),
			}),
		)
		checkPinHandlers(t, client, rootHash, true)
	})

	t.Run("mop", func(t *testing.T) {
		tarReader := tarFiles(t, []f{{
			data: []byte("<h1>Cluster"),
			name: "index.html",
			dir:  "",
		}})
		rootHash := "734b93933ed5d0a26ccfebf52a2d250c4f432a02c330fb2d49ce17e6ad46484f"
		jsonhttptest.Request(t, client, http.MethodPost, "/mop", http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.ClusterDeferredUploadHeader, "true"),
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(tarReader),
			jsonhttptest.WithRequestHeader(api.ContentTypeHeader, api.ContentTypeTar),
			jsonhttptest.WithRequestHeader(api.ClusterCollectionHeader, "true"),
			jsonhttptest.WithRequestHeader(api.ClusterPinHeader, "true"),
			jsonhttptest.WithExpectedJSONResponse(api.MopUploadResponse{
				Reference: cluster.MustParseHexAddress(rootHash),
			}),
		)
		checkPinHandlers(t, client, rootHash, false)

		header := jsonhttptest.Request(t, client, http.MethodPost, "/mop?name=somefile.txt", http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.ClusterDeferredUploadHeader, "true"),
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestHeader(api.ContentTypeHeader, "text/plain"),
			jsonhttptest.WithRequestHeader(api.ClusterEncryptHeader, "true"),
			jsonhttptest.WithRequestHeader(api.ClusterPinHeader, "true"),
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
			jsonhttptest.WithRequestHeader(api.ClusterDeferredUploadHeader, "true"),
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(bytes.NewReader(chunk.Data())),
			jsonhttptest.WithExpectedJSONResponse(api.ChunkAddressResponse{
				Reference: chunk.Address(),
			}),
		)
		checkPinHandlers(t, client, rootHash, true)
	})
}
