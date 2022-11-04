package api_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/redesblock/mop/core/api"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/api/jsonhttp/jsonhttptest"
	mockpost "github.com/redesblock/mop/core/incentives/voucher/mock"
	"github.com/redesblock/mop/core/log"
	statestore "github.com/redesblock/mop/core/storer/statestore/mock"
	"github.com/redesblock/mop/core/storer/storage/mock"
	testingc "github.com/redesblock/mop/core/storer/storage/testing"
	"github.com/redesblock/mop/core/tags"
)

func TestGatewayMode(t *testing.T) {
	logger := log.Noop
	chunk := testingc.GenerateTestRandomChunk()
	client, _, _, _ := newTestServer(t, testServerOptions{
		Storer:      mock.NewStorer(),
		Tags:        tags.NewTags(statestore.NewStateStore(), logger),
		Logger:      logger,
		GatewayMode: true,
		Post:        mockpost.New(mockpost.WithAcceptAll()),
	})

	forbiddenResponseOption := jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
		Message: http.StatusText(http.StatusForbidden),
		Code:    http.StatusForbidden,
	})

	t.Run("pins endpoints", func(t *testing.T) {
		path := "/pins/0773a91efd6547c754fc1d95fb1c62c7d1b47f959c2caa685dfec8736da95c1c"
		jsonhttptest.Request(t, client, http.MethodGet, path, http.StatusForbidden, forbiddenResponseOption)
		jsonhttptest.Request(t, client, http.MethodPost, path, http.StatusForbidden, forbiddenResponseOption)
		jsonhttptest.Request(t, client, http.MethodDelete, path, http.StatusForbidden, forbiddenResponseOption)
		jsonhttptest.Request(t, client, http.MethodGet, "/pins", http.StatusForbidden, forbiddenResponseOption)
	})

	t.Run("tags endpoints", func(t *testing.T) {
		path := "/tags/42"
		jsonhttptest.Request(t, client, http.MethodGet, path, http.StatusForbidden, forbiddenResponseOption)
		jsonhttptest.Request(t, client, http.MethodDelete, path, http.StatusForbidden, forbiddenResponseOption)
		jsonhttptest.Request(t, client, http.MethodPatch, path, http.StatusForbidden, forbiddenResponseOption)
		jsonhttptest.Request(t, client, http.MethodGet, "/tags", http.StatusForbidden, forbiddenResponseOption)
	})

	t.Run("psser endpoints", func(t *testing.T) {
		jsonhttptest.Request(t, client, http.MethodPost, "/psser/send/test-topic/ab", http.StatusForbidden, forbiddenResponseOption)
		jsonhttptest.Request(t, client, http.MethodGet, "/psser/subscribe/test-topic", http.StatusForbidden, forbiddenResponseOption)
	})

	t.Run("pins", func(t *testing.T) {
		headerOption := jsonhttptest.WithRequestHeader(api.ClusterPinHeader, "true")

		forbiddenResponseOption := jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
			Message: "pins is disabled",
			Code:    http.StatusForbidden,
		})

		// should work without pins
		jsonhttptest.Request(t, client, http.MethodPost, "/chunks", http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.ClusterDeferredUploadHeader, "true"),
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(bytes.NewReader(chunk.Data())),
		)

		jsonhttptest.Request(t, client, http.MethodPost, "/chunks/0773a91efd6547c754fc1d95fb1c62c7d1b47f959c2caa685dfec8736da95c1c", http.StatusForbidden, forbiddenResponseOption, headerOption)

		jsonhttptest.Request(t, client, http.MethodPost, "/bytes", http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.ClusterDeferredUploadHeader, "true"),
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(bytes.NewReader(chunk.Data())),
		) // should work without pins
		jsonhttptest.Request(t, client, http.MethodPost, "/bytes", http.StatusForbidden, forbiddenResponseOption, headerOption)
		jsonhttptest.Request(t, client, http.MethodPost, "/files", http.StatusForbidden, forbiddenResponseOption, headerOption)
		jsonhttptest.Request(t, client, http.MethodPost, "/dirs", http.StatusForbidden, forbiddenResponseOption, headerOption)
	})

	t.Run("encryption", func(t *testing.T) {
		headerOption := jsonhttptest.WithRequestHeader(api.ClusterEncryptHeader, "true")

		forbiddenResponseOption := jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
			Message: "encryption is disabled",
			Code:    http.StatusForbidden,
		})

		jsonhttptest.Request(t, client, http.MethodPost, "/bytes", http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.ClusterDeferredUploadHeader, "true"),
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(bytes.NewReader(chunk.Data())),
		) // should work without pins
		jsonhttptest.Request(t, client, http.MethodPost, "/bytes", http.StatusForbidden, forbiddenResponseOption, headerOption)
		jsonhttptest.Request(t, client, http.MethodPost, "/files", http.StatusForbidden, forbiddenResponseOption, headerOption)
		jsonhttptest.Request(t, client, http.MethodPost, "/dirs", http.StatusForbidden, forbiddenResponseOption, headerOption)
	})
}
