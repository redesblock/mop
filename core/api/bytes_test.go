package api_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/redesblock/mop/core/api"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/api/jsonhttp/jsonhttptest"
	"github.com/redesblock/mop/core/cluster"
	mockbatchstore "github.com/redesblock/mop/core/incentives/voucher/batchstore/mock"
	mockpost "github.com/redesblock/mop/core/incentives/voucher/mock"
	"github.com/redesblock/mop/core/log"
	pinning "github.com/redesblock/mop/core/pins/mock"
	statestore "github.com/redesblock/mop/core/storer/statestore/mock"
	"github.com/redesblock/mop/core/storer/storage/mock"
	"github.com/redesblock/mop/core/tags"
	"gitlab.com/nolash/go-mockbytes"
)

// TestBytes tests that the data upload api responds as expected when uploading,
// downloading and requesting a resource that cannot be found.
func TestBytes(t *testing.T) {
	const (
		resource = "/bytes"
		expHash  = "29a5fb121ce96194ba8b7b823a1f9c6af87e1791f824940a53b5a7efe3f790d9"
	)

	var (
		storerMock      = mock.NewStorer()
		pinningMock     = pinning.NewServiceMock()
		logger          = log.Noop
		client, _, _, _ = newTestServer(t, testServerOptions{
			Storer:  storerMock,
			Tags:    tags.NewTags(statestore.NewStateStore(), log.Noop),
			Pinning: pinningMock,
			Logger:  logger,
			Post:    mockpost.New(mockpost.WithAcceptAll()),
		})
	)

	g := mockbytes.New(0, mockbytes.MockTypeStandard).WithModulus(255)
	content, err := g.SequentialBytes(cluster.ChunkSize * 2)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("upload", func(t *testing.T) {
		chunkAddr := cluster.MustParseHexAddress(expHash)
		jsonhttptest.Request(t, client, http.MethodPost, resource, http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.ClusterDeferredUploadHeader, "true"),
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(bytes.NewReader(content)),
			jsonhttptest.WithExpectedJSONResponse(api.BytesPostResponse{
				Reference: chunkAddr,
			}),
		)

		has, err := storerMock.Has(context.Background(), chunkAddr)
		if err != nil {
			t.Fatal(err)
		}
		if !has {
			t.Fatal("storer check root chunk address: have none; want one")
		}

		refs, err := pinningMock.Pins()
		if err != nil {
			t.Fatal("unable to get pinned references")
		}
		if have, want := len(refs), 0; have != want {
			t.Fatalf("root pins count mismatch: have %d; want %d", have, want)
		}
	})

	t.Run("upload-with-pins", func(t *testing.T) {
		var res api.BytesPostResponse
		jsonhttptest.Request(t, client, http.MethodPost, resource, http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.ClusterDeferredUploadHeader, "true"),
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(bytes.NewReader(content)),
			jsonhttptest.WithRequestHeader(api.ClusterPinHeader, "true"),
			jsonhttptest.WithUnmarshalJSONResponse(&res),
		)
		reference := res.Reference

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

	t.Run("download", func(t *testing.T) {
		resp := request(t, client, http.MethodGet, resource+"/"+expHash, nil, http.StatusOK)
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(data, content) {
			t.Fatalf("data mismatch. got %s, want %s", string(data), string(content))
		}
	})

	t.Run("head", func(t *testing.T) {
		resp := request(t, client, http.MethodHead, resource+"/"+expHash, nil, http.StatusOK)
		if int(resp.ContentLength) != len(content) {
			t.Fatalf("length %d want %d", resp.ContentLength, len(content))
		}
	})
	t.Run("head with compression", func(t *testing.T) {
		resp := jsonhttptest.Request(t, client, http.MethodHead, resource+"/"+expHash, http.StatusOK,
			jsonhttptest.WithRequestHeader("Accept-Encoding", "gzip"),
		)
		val, err := strconv.Atoi(resp.Get("Content-Length"))
		if err != nil {
			t.Fatal(err)
		}
		if val != len(content) {
			t.Fatalf("length %d want %d", val, len(content))
		}
	})

	t.Run("not found", func(t *testing.T) {
		jsonhttptest.Request(t, client, http.MethodGet, resource+"/0xabcd", http.StatusNotFound,
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: "Not Found",
				Code:    http.StatusNotFound,
			}),
		)
	})

	t.Run("internal error", func(t *testing.T) {
		jsonhttptest.Request(t, client, http.MethodGet, resource+"/abcd", http.StatusInternalServerError,
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: "api download: joiner failed",
				Code:    http.StatusInternalServerError,
			}),
		)
	})

	t.Run("upload multipart error", func(t *testing.T) {
		jsonhttptest.Request(t, client, http.MethodPost, resource, http.StatusBadRequest,
			jsonhttptest.WithRequestHeader(api.ClusterDeferredUploadHeader, "true"),
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestHeader(api.ContentTypeHeader, "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW"),
			jsonhttptest.WithRequestBody(bytes.NewReader(content)),
		)
	})
}

func TestBytesInvalidStamp(t *testing.T) {
	const (
		resource = "/bytes"
		expHash  = "29a5fb121ce96194ba8b7b823a1f9c6af87e1791f824940a53b5a7efe3f790d9"
	)

	var (
		storerMock        = mock.NewStorer()
		pinningMock       = pinning.NewServiceMock()
		logger            = log.Noop
		retBool           = false
		retErr      error = nil
		existsFn          = func(id []byte) (bool, error) {
			return retBool, retErr
		}
		client, _, _, _ = newTestServer(t, testServerOptions{
			Storer:     storerMock,
			Tags:       tags.NewTags(statestore.NewStateStore(), log.Noop),
			Pinning:    pinningMock,
			Logger:     logger,
			Post:       mockpost.New(mockpost.WithAcceptAll()),
			BatchStore: mockbatchstore.New(mockbatchstore.WithExistsFunc(existsFn)),
		})
	)

	g := mockbytes.New(0, mockbytes.MockTypeStandard).WithModulus(255)
	content, err := g.SequentialBytes(cluster.ChunkSize * 2)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("upload, batch doesn't exist", func(t *testing.T) {
		chunkAddr := cluster.MustParseHexAddress(expHash)
		jsonhttptest.Request(t, client, http.MethodPost, resource, http.StatusBadRequest,
			jsonhttptest.WithRequestHeader(api.ClusterDeferredUploadHeader, "true"),
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(bytes.NewReader(content)),
		)

		has, err := storerMock.Has(context.Background(), chunkAddr)
		if err != nil {
			t.Fatal(err)
		}
		if has {
			t.Fatal("storer check root chunk address: have ont; want none")
		}

		refs, err := pinningMock.Pins()
		if err != nil {
			t.Fatal("unable to get pinned references")
		}
		if have, want := len(refs), 0; have != want {
			t.Fatalf("root pins count mismatch: have %d; want %d", have, want)
		}
	})

	// throw back an error
	retErr = errors.New("err happened")

	t.Run("upload, batch exists error", func(t *testing.T) {
		chunkAddr := cluster.MustParseHexAddress(expHash)
		jsonhttptest.Request(t, client, http.MethodPost, resource, http.StatusBadRequest,
			jsonhttptest.WithRequestHeader(api.ClusterDeferredUploadHeader, "true"),
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(bytes.NewReader(content)),
		)

		has, err := storerMock.Has(context.Background(), chunkAddr)
		if err != nil {
			t.Fatal(err)
		}
		if has {
			t.Fatal("storer check root chunk address: have ont; want none")
		}
	})
}
