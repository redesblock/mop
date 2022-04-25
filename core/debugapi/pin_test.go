package debugapi_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/jsonhttp/jsonhttptest"
	"github.com/redesblock/hop/core/storage/mock"
	"github.com/redesblock/hop/core/storage/mock/validator"
	"github.com/redesblock/hop/core/swarm"
)

// TestPinChunkHandler checks for pinning, unpinning and listing of chunks.
// It also check other edgw cases like chunk not present and checking for pinning,
// invalid chunk address case etc. This test case has to be run in sequence and
// it assumes some state of the DB before another case is run.
func TestPinChunkHandler(t *testing.T) {
	resource := func(addr swarm.Address) string { return "/chunk/" + addr.String() }
	hash := swarm.MustParseHexAddress("aabbcc")
	data := []byte("bbaatt")
	mockValidator := validator.NewMockValidator(hash, data)
	mockValidatingStorer := mock.NewValidatingStorer(mockValidator)
	debugTestServer := newTestServer(t, testServerOptions{
		Storer: mockValidatingStorer,
	})
	// This server is used to store chunks
	hopTestServer := newHopTestServer(t, testServerOptions{
		Storer: mockValidatingStorer,
	})

	// bad chunk address
	t.Run("pin-bad-address", func(t *testing.T) {
		jsonhttptest.ResponseDirect(t, debugTestServer.Client, http.MethodPost, "/chunks-pin/abcd1100zz", nil, http.StatusBadRequest, jsonhttp.StatusResponse{
			Message: "bad address",
			Code:    http.StatusBadRequest,
		})
	})

	// pin a chunk which is not existing
	t.Run("pin-absent-chunk", func(t *testing.T) {
		jsonhttptest.ResponseDirect(t, debugTestServer.Client, http.MethodPost, "/chunks-pin/123456", nil, http.StatusNotFound, jsonhttp.StatusResponse{
			Message: http.StatusText(http.StatusNotFound),
			Code:    http.StatusNotFound,
		})
	})

	// unpin on a chunk which is not pinned
	t.Run("unpin-while-not-pinned", func(t *testing.T) {
		// Post a chunk
		jsonhttptest.ResponseDirect(t, hopTestServer, http.MethodPost, resource(hash), bytes.NewReader(data), http.StatusOK, jsonhttp.StatusResponse{
			Message: http.StatusText(http.StatusOK),
			Code:    http.StatusOK,
		})

		jsonhttptest.ResponseDirect(t, debugTestServer.Client, http.MethodDelete, "/chunks-pin/"+hash.String(), nil, http.StatusBadRequest, jsonhttp.StatusResponse{
			Message: "chunk is not yet pinned",
			Code:    http.StatusBadRequest,
		})
	})

	// pin a existing chunk first time
	t.Run("pin-chunk-1", func(t *testing.T) {
		// Post a chunk
		jsonhttptest.ResponseDirect(t, hopTestServer, http.MethodPost, resource(hash), bytes.NewReader(data), http.StatusOK, jsonhttp.StatusResponse{
			Message: http.StatusText(http.StatusOK),
			Code:    http.StatusOK,
		})

		jsonhttptest.ResponseDirect(t, debugTestServer.Client, http.MethodPost, "/chunks-pin/"+hash.String(), nil, http.StatusOK, jsonhttp.StatusResponse{
			Message: http.StatusText(http.StatusOK),
			Code:    http.StatusOK,
		})

		// Check is the chunk is pinned once
		jsonhttptest.ResponseDirectWithJson(t, debugTestServer.Client, http.MethodGet, "/chunks-pin/"+hash.String(), nil, http.StatusOK, jsonhttp.StatusResponse{
			Message: `{"Address":"aabbcc","PinCounter":1}`,
			Code:    http.StatusOK,
		})

	})

	// pin a existing chunk second time
	t.Run("pin-chunk-2", func(t *testing.T) {
		jsonhttptest.ResponseDirect(t, debugTestServer.Client, http.MethodPost, "/chunks-pin/"+hash.String(), nil, http.StatusOK, jsonhttp.StatusResponse{
			Message: http.StatusText(http.StatusOK),
			Code:    http.StatusOK,
		})

		// Check is the chunk is pinned twice
		jsonhttptest.ResponseDirectWithJson(t, debugTestServer.Client, http.MethodGet, "/chunks-pin/"+hash.String(), nil, http.StatusOK, jsonhttp.StatusResponse{
			Message: `{"Address":"aabbcc","PinCounter":2}`,
			Code:    http.StatusOK,
		})
	})

	// unpin a chunk first time
	t.Run("unpin-chunk-1", func(t *testing.T) {
		jsonhttptest.ResponseDirect(t, debugTestServer.Client, http.MethodDelete, "/chunks-pin/"+hash.String(), nil, http.StatusOK, jsonhttp.StatusResponse{
			Message: http.StatusText(http.StatusOK),
			Code:    http.StatusOK,
		})

		// Check is the chunk is pinned once
		jsonhttptest.ResponseDirectWithJson(t, debugTestServer.Client, http.MethodGet, "/chunks-pin/"+hash.String(), nil, http.StatusOK, jsonhttp.StatusResponse{
			Message: `{"Address":"aabbcc","PinCounter":1}`,
			Code:    http.StatusOK,
		})
	})

	// unpin a chunk second time
	t.Run("unpin-chunk-2", func(t *testing.T) {
		jsonhttptest.ResponseDirect(t, debugTestServer.Client, http.MethodDelete, "/chunks-pin/"+hash.String(), nil, http.StatusOK, jsonhttp.StatusResponse{
			Message: http.StatusText(http.StatusOK),
			Code:    http.StatusOK,
		})

		// Check if the chunk is removed from the pinIndex
		jsonhttptest.ResponseDirect(t, debugTestServer.Client, http.MethodGet, "/chunks-pin/"+hash.String(), nil, http.StatusInternalServerError, jsonhttp.StatusResponse{
			Message: "pin chunks: leveldb: not found",
			Code:    http.StatusInternalServerError,
		})
	})

	// Add 2 chunks, pin it and check if they show up in the list
	t.Run("list-chunks", func(t *testing.T) {
		// Post a chunk
		jsonhttptest.ResponseDirect(t, hopTestServer, http.MethodPost, resource(hash), bytes.NewReader(data), http.StatusOK, jsonhttp.StatusResponse{
			Message: http.StatusText(http.StatusOK),
			Code:    http.StatusOK,
		})

		jsonhttptest.ResponseDirect(t, debugTestServer.Client, http.MethodPost, "/chunks-pin/"+hash.String(), nil, http.StatusOK, jsonhttp.StatusResponse{
			Message: http.StatusText(http.StatusOK),
			Code:    http.StatusOK,
		})

		// post another chunk
		hash2 := swarm.MustParseHexAddress("ddeeff")
		data2 := []byte("eagle")
		mockValidator.AddPair(hash2, data2)
		jsonhttptest.ResponseDirect(t, hopTestServer, http.MethodPost, resource(hash2), bytes.NewReader(data2), http.StatusOK, jsonhttp.StatusResponse{
			Message: http.StatusText(http.StatusOK),
			Code:    http.StatusOK,
		})
		jsonhttptest.ResponseDirect(t, debugTestServer.Client, http.MethodPost, "/chunks-pin/"+hash2.String(), nil, http.StatusOK, jsonhttp.StatusResponse{
			Message: http.StatusText(http.StatusOK),
			Code:    http.StatusOK,
		})

		jsonhttptest.ResponseDirectWithJson(t, debugTestServer.Client, http.MethodGet, "/chunks-pin", nil, http.StatusOK, jsonhttp.StatusResponse{
			Message: `{"Address":"aabbcc","PinCounter":1},{"Address":"ddeeff","PinCounter":1}`,
			Code:    http.StatusOK,
		})

	})
}
