package api_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/redesblock/mop/core/api"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/api/jsonhttp/jsonhttptest"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/p2p"
	pingpongmock "github.com/redesblock/mop/core/protocol/pingpong/mock"
)

func TestPingpong(t *testing.T) {
	rtt := time.Minute
	peerID := cluster.MustParseHexAddress("ca1e9f3938cc1425c6061b96ad9eb93e134dfe8734ad490164ef20af9d1cf59c")
	unknownPeerID := cluster.MustParseHexAddress("ca1e9f3938cc1425c6061b96ad9eb93e134dfe8734ad490164ef20af9d1cf59e")
	errorPeerID := cluster.MustParseHexAddress("ca1e9f3938cc1425c6061b96ad9eb93e134dfe8734ad490164ef20af9d1cf59a")
	testErr := errors.New("test error")

	pingpongService := pingpongmock.New(func(ctx context.Context, address cluster.Address, msgs ...string) (time.Duration, error) {
		if address.Equal(errorPeerID) {
			return 0, testErr
		}
		if !address.Equal(peerID) {
			return 0, p2p.ErrPeerNotFound
		}
		return rtt, nil
	})

	ts, _, _, _ := newTestServer(t, testServerOptions{
		DebugAPI: true,
		Pingpong: pingpongService,
	})

	t.Run("ok", func(t *testing.T) {
		jsonhttptest.Request(t, ts, http.MethodPost, "/pingpong/"+peerID.String(), http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(api.PingpongResponse{
				RTT: rtt.String(),
			}),
		)
	})

	t.Run("peer not found", func(t *testing.T) {
		jsonhttptest.Request(t, ts, http.MethodPost, "/pingpong/"+unknownPeerID.String(), http.StatusNotFound,
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Code:    http.StatusNotFound,
				Message: "peer not found",
			}),
		)
	})

	t.Run("invalid peer address", func(t *testing.T) {
		jsonhttptest.Request(t, ts, http.MethodPost, "/pingpong/invalid-address", http.StatusBadRequest,
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Code:    http.StatusBadRequest,
				Message: "invalid peer address",
			}),
		)
	})

	t.Run("error", func(t *testing.T) {
		jsonhttptest.Request(t, ts, http.MethodPost, "/pingpong/"+errorPeerID.String(), http.StatusInternalServerError,
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Code:    http.StatusInternalServerError,
				Message: "pingpong: ping failed",
			}),
		)
	})

	t.Run("get method not allowed", func(t *testing.T) {
		jsonhttptest.Request(t, ts, http.MethodGet, "/pingpong/"+peerID.String(), http.StatusMethodNotAllowed,
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Code:    http.StatusMethodNotAllowed,
				Message: http.StatusText(http.StatusMethodNotAllowed),
			}),
		)
	})
}
