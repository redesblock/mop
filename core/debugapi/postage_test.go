package debugapi_test

import (
	"net/http"
	"testing"

	"github.com/redesblock/hop/core/jsonhttp/jsonhttptest"
	"github.com/redesblock/hop/core/postage"
	"github.com/redesblock/hop/core/postage/batchstore/mock"
)

func TestReservestate(t *testing.T) {

	ts := newTestServer(t, testServerOptions{
		BatchStore: mock.New(mock.WithReserveState(&postage.Reservestate{
			Radius: 5,
		})),
	})

	t.Run("ok", func(t *testing.T) {
		jsonhttptest.Request(t, ts.Client, http.MethodGet, "/reservestate", http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(&postage.Reservestate{
				Radius: 5,
			}),
		)
	})
}
