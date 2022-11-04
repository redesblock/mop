package api_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redesblock/mop/core/api"
	"github.com/redesblock/mop/core/cluster"
	mockpost "github.com/redesblock/mop/core/incentives/voucher/mock"
	"github.com/redesblock/mop/core/log"
	pinning "github.com/redesblock/mop/core/pins/mock"
	statestore "github.com/redesblock/mop/core/storer/statestore/mock"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/redesblock/mop/core/storer/storage/mock"
	testingc "github.com/redesblock/mop/core/storer/storage/testing"
	"github.com/redesblock/mop/core/tags"
)

func TestChunkUploadStream(t *testing.T) {
	wsHeaders := http.Header{}
	wsHeaders.Set(api.ClusterDeferredUploadHeader, "true")
	wsHeaders.Set(api.ContentTypeHeader, "application/octet-stream")
	wsHeaders.Set(api.ClusterVoucherBatchIdHeader, batchOkStr)

	var (
		statestoreMock  = statestore.NewStateStore()
		logger          = log.Noop
		tag             = tags.NewTags(statestoreMock, logger)
		storerMock      = mock.NewStorer()
		pinningMock     = pinning.NewServiceMock()
		_, wsConn, _, _ = newTestServer(t, testServerOptions{
			Storer:    storerMock,
			Pinning:   pinningMock,
			Tags:      tag,
			Post:      mockpost.New(mockpost.WithAcceptAll()),
			WsPath:    "/chunks/stream",
			WsHeaders: wsHeaders,
		})
	)

	t.Run("upload and verify", func(t *testing.T) {
		chsToGet := []cluster.Chunk{}
		for i := 0; i < 5; i++ {
			ch := testingc.GenerateTestRandomChunk()

			err := wsConn.SetWriteDeadline(time.Now().Add(time.Second))
			if err != nil {
				t.Fatal(err)
			}

			err = wsConn.WriteMessage(websocket.BinaryMessage, ch.Data())
			if err != nil {
				t.Fatal(err)
			}

			err = wsConn.SetReadDeadline(time.Now().Add(time.Second))
			if err != nil {
				t.Fatal(err)
			}

			mt, msg, err := wsConn.ReadMessage()
			if err != nil {
				t.Fatal(err)
			}

			if mt != websocket.BinaryMessage || !bytes.Equal(msg, api.SuccessWsMsg) {
				t.Fatal("invalid response", mt, string(msg))
			}

			chsToGet = append(chsToGet, ch)
		}

		for _, c := range chsToGet {
			ch, err := storerMock.Get(context.Background(), storage.ModeGetRequest, c.Address())
			if err != nil {
				t.Fatal("failed to get chunk after upload", err)
			}
			if !ch.Equal(c) {
				t.Fatal("invalid chunk read")
			}
		}
	})

	t.Run("close on incorrect msg", func(t *testing.T) {
		err := wsConn.SetWriteDeadline(time.Now().Add(time.Second))
		if err != nil {
			t.Fatal(err)
		}

		err = wsConn.WriteMessage(websocket.TextMessage, []byte("incorrect msg"))
		if err != nil {
			t.Fatal(err)
		}

		err = wsConn.SetReadDeadline(time.Now().Add(time.Second))
		if err != nil {
			t.Fatal(err)
		}

		_, _, err = wsConn.ReadMessage()
		if err == nil {
			t.Fatal("expected failure on read")
		}
		if cerr, ok := err.(*websocket.CloseError); !ok {
			t.Fatal("invalid error on read")
		} else if cerr.Text != "invalid message" {
			t.Fatalf("incorrect response on error, exp: (invalid message) got (%s)", cerr.Text)
		}
	})
}
