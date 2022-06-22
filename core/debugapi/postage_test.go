package debugapi_test

import (
	"math/big"
	"net/http"
	"testing"

	"github.com/redesblock/hop/core/bigint"
	"github.com/redesblock/hop/core/debugapi"
	"github.com/redesblock/hop/core/jsonhttp/jsonhttptest"
	"github.com/redesblock/hop/core/postage"
	"github.com/redesblock/hop/core/postage/batchstore/mock"
)

func TestReserveState(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		ts := newTestServer(t, testServerOptions{
			BatchStore: mock.New(mock.WithReserveState(&postage.ReserveState{
				Radius: 5,
				Outer:  big.NewInt(5),
				Inner:  big.NewInt(5),
			})),
		})
		jsonhttptest.Request(t, ts.Client, http.MethodGet, "/reservestate", http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(&debugapi.ReserveStateResponse{
				Radius: 5,
				Outer:  bigint.Wrap(big.NewInt(5)),
				Inner:  bigint.Wrap(big.NewInt(5)),
			}),
		)
	})
	t.Run("empty", func(t *testing.T) {
		ts := newTestServer(t, testServerOptions{
			BatchStore: mock.New(),
		})
		jsonhttptest.Request(t, ts.Client, http.MethodGet, "/reservestate", http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(&debugapi.ReserveStateResponse{}),
		)
	})
}

func TestChainState(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		cs := &postage.ChainState{
			Block:        123456,
			TotalAmount:  big.NewInt(50),
			CurrentPrice: big.NewInt(5),
		}
		ts := newTestServer(t, testServerOptions{
			BatchStore: mock.New(mock.WithChainState(cs)),
		})
		jsonhttptest.Request(t, ts.Client, http.MethodGet, "/chainstate", http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(&debugapi.ChainStateResponse{
				Block:        123456,
				TotalAmount:  bigint.Wrap(big.NewInt(50)),
				CurrentPrice: bigint.Wrap(big.NewInt(5)),
			}),
		)
	})

	t.Run("empty", func(t *testing.T) {
		ts := newTestServer(t, testServerOptions{
			BatchStore: mock.New(),
		})
		jsonhttptest.Request(t, ts.Client, http.MethodGet, "/chainstate", http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(&debugapi.ChainStateResponse{}),
		)
	})
}
