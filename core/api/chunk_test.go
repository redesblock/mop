package api_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	mockpost "github.com/redesblock/mop/core/incentives/voucher/mock"
	"github.com/redesblock/mop/core/log"
	pinning "github.com/redesblock/mop/core/pins/mock"
	statestore "github.com/redesblock/mop/core/storer/statestore/mock"

	"github.com/redesblock/mop/core/tags"

	"github.com/redesblock/mop/core/api"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/api/jsonhttp/jsonhttptest"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/redesblock/mop/core/storer/storage/mock"
	testingc "github.com/redesblock/mop/core/storer/storage/testing"
)

// TestChunkUploadDownload uploads a chunk to an API that verifies the chunk according
// to a given validator, then tries to download the uploaded data.
func TestChunkUploadDownload(t *testing.T) {
	var (
		chunksEndpoint  = "/chunks"
		chunksResource  = func(a cluster.Address) string { return "/chunks/" + a.String() }
		chunk           = testingc.GenerateTestRandomChunk()
		statestoreMock  = statestore.NewStateStore()
		logger          = log.Noop
		tag             = tags.NewTags(statestoreMock, logger)
		storerMock      = mock.NewStorer()
		pinningMock     = pinning.NewServiceMock()
		client, _, _, _ = newTestServer(t, testServerOptions{
			Storer:  storerMock,
			Pinning: pinningMock,
			Tags:    tag,
			Post:    mockpost.New(mockpost.WithAcceptAll()),
		})
	)

	t.Run("empty chunk", func(t *testing.T) {
		jsonhttptest.Request(t, client, http.MethodPost, chunksEndpoint, http.StatusBadRequest,
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: "insufficient data length",
				Code:    http.StatusBadRequest,
			}),
		)
	})

	t.Run("ok", func(t *testing.T) {
		jsonhttptest.Request(t, client, http.MethodPost, chunksEndpoint, http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.ClusterDeferredUploadHeader, "true"),
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(bytes.NewReader(chunk.Data())),
			jsonhttptest.WithExpectedJSONResponse(api.ChunkAddressResponse{Reference: chunk.Address()}),
		)

		has, err := storerMock.Has(context.Background(), chunk.Address())
		if err != nil {
			t.Fatal(err)
		}
		if !has {
			t.Fatal("storer check root chunk reference: have none; want one")
		}

		// try to fetch the same chunk
		endpoint := chunksResource(chunk.Address())
		resp := request(t, client, http.MethodGet, endpoint, nil, http.StatusOK)
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(chunk.Data(), data) {
			t.Fatal("data retrieved doesnt match uploaded content")
		}
	})

	t.Run("pins-invalid-value", func(t *testing.T) {
		jsonhttptest.Request(t, client, http.MethodPost, chunksEndpoint, http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.ClusterDeferredUploadHeader, "true"),
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(bytes.NewReader(chunk.Data())),
			jsonhttptest.WithExpectedJSONResponse(api.ChunkAddressResponse{Reference: chunk.Address()}),
			jsonhttptest.WithRequestHeader(api.ClusterPinHeader, "invalid-pins"),
		)

		// Also check if the chunk is NOT pinned
		if storerMock.GetModeSet(chunk.Address()) == storage.ModeSetPin {
			t.Fatal("chunk should not be pinned")
		}
	})
	t.Run("pins-header-missing", func(t *testing.T) {
		jsonhttptest.Request(t, client, http.MethodPost, chunksEndpoint, http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.ClusterDeferredUploadHeader, "true"),
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(bytes.NewReader(chunk.Data())),
			jsonhttptest.WithExpectedJSONResponse(api.ChunkAddressResponse{Reference: chunk.Address()}),
		)

		// Also check if the chunk is NOT pinned
		if storerMock.GetModeSet(chunk.Address()) == storage.ModeSetPin {
			t.Fatal("chunk should not be pinned")
		}
	})
	t.Run("pins-ok", func(t *testing.T) {
		reference := chunk.Address()
		jsonhttptest.Request(t, client, http.MethodPost, chunksEndpoint, http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.ClusterDeferredUploadHeader, "true"),
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(bytes.NewReader(chunk.Data())),
			jsonhttptest.WithExpectedJSONResponse(api.ChunkAddressResponse{Reference: reference}),
			jsonhttptest.WithRequestHeader(api.ClusterPinHeader, "True"),
		)

		has, err := storerMock.Has(context.Background(), reference)
		if err != nil {
			t.Fatal(err)
		}
		if !has {
			t.Fatal("storer check root chunk reference: have none; want one")
		}

		refs, err := pinningMock.Pins()
		if err != nil {
			t.Fatal(err)
		}
		if have, want := len(refs), 1; have != want {
			t.Fatalf("root pins count mismatch: have %d; want %d", have, want)
		}
		if have, want := refs[0], reference; !have.Equal(want) {
			t.Fatalf("root pins reference mismatch: have %q; want %q", have, want)
		}
	})
}

func TestHasChunkHandler(t *testing.T) {
	mockStorer := mock.NewStorer()
	testServer, _, _, _ := newTestServer(t, testServerOptions{
		Storer: mockStorer,
	})

	key := cluster.MustParseHexAddress("aabbcc")
	value := []byte("data data data")

	_, err := mockStorer.Put(context.Background(), storage.ModePutUpload, cluster.NewChunk(key, value))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("ok", func(t *testing.T) {
		jsonhttptest.Request(t, testServer, http.MethodHead, "/chunks/"+key.String(), http.StatusOK,
			jsonhttptest.WithNoResponseBody())
	})

	t.Run("not found", func(t *testing.T) {
		jsonhttptest.Request(t, testServer, http.MethodHead, "/chunks/abbbbb", http.StatusNotFound,
			jsonhttptest.WithNoResponseBody())
	})

	t.Run("bad address", func(t *testing.T) {
		jsonhttptest.Request(t, testServer, http.MethodHead, "/chunks/abcd1100zz", http.StatusBadRequest,
			jsonhttptest.WithNoResponseBody())
	})

	t.Run("remove-chunk", func(t *testing.T) {
		jsonhttptest.Request(t, testServer, http.MethodDelete, "/chunks/"+key.String(), http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: http.StatusText(http.StatusOK),
				Code:    http.StatusOK,
			}),
		)
		yes, err := mockStorer.Has(context.Background(), key)
		if err != nil {
			t.Fatal(err)
		}
		if yes {
			t.Fatalf("The chunk %s is not deleted", key.String())
		}
	})

	t.Run("remove-not-present-chunk", func(t *testing.T) {
		notPresentChunkAddress := "deadbeef"
		jsonhttptest.Request(t, testServer, http.MethodDelete, "/chunks/"+notPresentChunkAddress, http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: http.StatusText(http.StatusOK),
				Code:    http.StatusOK,
			}),
		)
		yes, err := mockStorer.Has(context.Background(), cluster.NewAddress([]byte(notPresentChunkAddress)))
		if err != nil {
			t.Fatal(err)
		}
		if yes {
			t.Fatalf("The chunk %s is not deleted", notPresentChunkAddress)
		}
	})
}
