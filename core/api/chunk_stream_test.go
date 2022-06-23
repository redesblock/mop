package api_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redesblock/hop/core/api"
	"github.com/redesblock/hop/core/logging"
	pinning "github.com/redesblock/hop/core/pinning/mock"
	mockpost "github.com/redesblock/hop/core/postage/mock"
	statestore "github.com/redesblock/hop/core/statestore/mock"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/storage/mock"
	testingc "github.com/redesblock/hop/core/storage/testing"
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/tags"
)

func TestChunkUploadStream(t *testing.T) {

	wsHeaders := http.Header{}
	wsHeaders.Set("Content-Type", "application/octet-stream")
	wsHeaders.Set("Swarm-Postage-Batch-Id", batchOkStr)

	var (
		statestoreMock = statestore.NewStateStore()
		logger         = logging.New(ioutil.Discard, 0)
		tag            = tags.NewTags(statestoreMock, logger)
		storerMock     = mock.NewStorer()
		pinningMock    = pinning.NewServiceMock()
		_, wsConn, _   = newTestServer(t, testServerOptions{
			Storer:    storerMock,
			Pinning:   pinningMock,
			Tags:      tag,
			Post:      mockpost.New(mockpost.WithAcceptAll()),
			WsPath:    "/chunks/stream",
			WsHeaders: wsHeaders,
		})
	)

	t.Run("upload and verify", func(t *testing.T) {
		chsToGet := []swarm.Chunk{}
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
