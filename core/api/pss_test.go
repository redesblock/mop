package api_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/redesblock/mop/core/chunk/trojan"
	"github.com/redesblock/mop/core/psser"
	"math/big"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/gorilla/websocket"
	"github.com/redesblock/mop/core/api"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/api/jsonhttp/jsonhttptest"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/incentives/voucher"
	mockpost "github.com/redesblock/mop/core/incentives/voucher/mock"
	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/protocol/pushsync"
	"github.com/redesblock/mop/core/storer/storage/mock"
)

var (
	target  = trojan.Target([]byte{1})
	targets = trojan.Targets([]trojan.Target{target})
	payload = []byte("testdata")
	topic   = trojan.NewTopic("testtopic")
	// mTimeout is used to wait for checking the message contents, whereas rTimeout
	// is used to wait for reading the message. For the negative cases, i.e. ensuring
	// no message is received, the rTimeout might trigger before the mTimeout
	// (Issue #1388) causing test to fail. Hence the rTimeout should be slightly more
	// than the mTimeout
	mTimeout    = 10 * time.Second
	rTimeout    = 15 * time.Second
	longTimeout = 30 * time.Second
)

// creates a single websocket handler for an arbitrary topic, and receives a message
func TestPssWebsocketSingleHandler(t *testing.T) {
	var (
		p, publicKey, cl, _ = newPssTest(t, opts{})

		msgContent = make([]byte, len(payload))
		tc         cluster.Chunk
		mtx        sync.Mutex
		done       = make(chan struct{})
	)

	// the long timeout is needed so that we dont time out while still mining the message with Wrap()
	// otherwise the test (and other tests below) flakes
	err := cl.SetReadDeadline(time.Now().Add(longTimeout))
	if err != nil {
		t.Fatal(err)
	}
	cl.SetReadLimit(cluster.ChunkSize)

	defer close(done)
	go waitReadMessage(t, &mtx, cl, msgContent, done)

	tc, err = trojan.Wrap(context.Background(), topic, payload, publicKey, targets)
	if err != nil {
		t.Fatal(err)
	}

	p.TryUnwrap(tc)

	waitMessage(t, msgContent, payload, &mtx)
}

func TestPssWebsocketSingleHandlerDeregister(t *testing.T) {
	// create a new psser instance, register a handle through ws, call
	// trojan.TryUnwrap with a chunk designated for this handler and expect
	// the handler to be notified
	var (
		p, publicKey, cl, _ = newPssTest(t, opts{})

		msgContent = make([]byte, len(payload))
		tc         cluster.Chunk
		mtx        sync.Mutex
		done       = make(chan struct{})
	)

	err := cl.SetReadDeadline(time.Now().Add(longTimeout))

	if err != nil {
		t.Fatal(err)
	}
	cl.SetReadLimit(cluster.ChunkSize)
	defer close(done)
	go waitReadMessage(t, &mtx, cl, msgContent, done)

	tc, err = trojan.Wrap(context.Background(), topic, payload, publicKey, targets)
	if err != nil {
		t.Fatal(err)
	}

	// close the websocket before calling psser with the message
	err = cl.WriteMessage(websocket.CloseMessage, []byte{})
	if err != nil {
		t.Fatal(err)
	}

	p.TryUnwrap(tc)

	waitMessage(t, msgContent, nil, &mtx)
}

func TestPssWebsocketMultiHandler(t *testing.T) {
	var (
		p, publicKey, cl, listener = newPssTest(t, opts{})

		u           = url.URL{Scheme: "ws", Host: listener, Path: "/psser/subscribe/testtopic"}
		cl2, _, err = websocket.DefaultDialer.Dial(u.String(), nil)

		msgContent  = make([]byte, len(payload))
		msgContent2 = make([]byte, len(payload))
		tc          cluster.Chunk
		mtx         sync.Mutex
		done        = make(chan struct{})
	)
	if err != nil {
		t.Fatalf("dial: %v. url %v", err, u.String())
	}

	err = cl.SetReadDeadline(time.Now().Add(longTimeout))
	if err != nil {
		t.Fatal(err)
	}
	cl.SetReadLimit(cluster.ChunkSize)

	defer close(done)
	go waitReadMessage(t, &mtx, cl, msgContent, done)
	go waitReadMessage(t, &mtx, cl2, msgContent2, done)

	tc, err = trojan.Wrap(context.Background(), topic, payload, publicKey, targets)
	if err != nil {
		t.Fatal(err)
	}

	// close the websocket before calling psser with the message
	err = cl.WriteMessage(websocket.CloseMessage, []byte{})
	if err != nil {
		t.Fatal(err)
	}

	p.TryUnwrap(tc)

	waitMessage(t, msgContent, nil, &mtx)
	waitMessage(t, msgContent2, nil, &mtx)
}

// TestPssSend tests that the psser message sending over http works correctly.
func TestPssSend(t *testing.T) {
	var (
		logger = log.Noop

		mtx             sync.Mutex
		receivedTopic   trojan.Topic
		receivedBytes   []byte
		receivedTargets trojan.Targets
		done            bool

		privk, _       = crypto.GenerateSecp256k1Key()
		publicKeyBytes = (*btcec.PublicKey)(&privk.PublicKey).SerializeCompressed()

		sendFn = func(ctx context.Context, targets trojan.Targets, chunk cluster.Chunk) error {
			mtx.Lock()
			topic, msg, err := trojan.Unwrap(ctx, privk, chunk, []trojan.Topic{topic})
			receivedTopic = topic
			receivedBytes = msg
			receivedTargets = targets
			done = true
			mtx.Unlock()
			return err
		}
		mp              = mockpost.New(mockpost.WithIssuer(voucher.NewStampIssuer("", "", batchOk, big.NewInt(3), 11, 10, 1000, true)))
		p               = newMockPss(sendFn)
		client, _, _, _ = newTestServer(t, testServerOptions{
			Pss:    p,
			Storer: mock.NewStorer(),
			Logger: logger,
			Post:   mp,
		})

		recipient = hex.EncodeToString(publicKeyBytes)
		targets   = fmt.Sprintf("[[%d]]", 0x12)
		topic     = "testtopic"
		hasher    = cluster.NewHasher()
		_, err    = hasher.Write([]byte(topic))
		topicHash = hasher.Sum(nil)
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("err - bad targets", func(t *testing.T) {
		jsonhttptest.Request(t, client, http.MethodPost, "/psser/send/to/badtarget?recipient="+recipient, http.StatusBadRequest,
			jsonhttptest.WithRequestBody(bytes.NewReader(payload)),
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: "target is not valid hex string",
				Code:    http.StatusBadRequest,
			}),
		)

		// If this test needs to be modified (most probably because the max target length changed)
		// the please verify that Common.yaml -> components -> PssTarget also reflects the correct value
		jsonhttptest.Request(t, client, http.MethodPost, "/psser/send/to/123456789abcdf?recipient="+recipient, http.StatusBadRequest,
			jsonhttptest.WithRequestBody(bytes.NewReader(payload)),
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: "hex string target exceeds max length of 6",
				Code:    http.StatusBadRequest,
			}),
		)
	})

	t.Run("err - bad batch", func(t *testing.T) {
		hexbatch := hex.EncodeToString(batchInvalid)
		jsonhttptest.Request(t, client, http.MethodPost, "/psser/send/to/12", http.StatusBadRequest,
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, hexbatch),
			jsonhttptest.WithRequestBody(bytes.NewReader(payload)),
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: "invalid voucher batch id",
				Code:    http.StatusBadRequest,
			}),
		)
	})

	t.Run("ok batch", func(t *testing.T) {
		hexbatch := hex.EncodeToString(batchOk)
		jsonhttptest.Request(t, client, http.MethodPost, "/psser/send/to/12", http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, hexbatch),
			jsonhttptest.WithRequestBody(bytes.NewReader(payload)),
		)
	})
	t.Run("bad request - batch empty", func(t *testing.T) {
		hexbatch := hex.EncodeToString(batchEmpty)
		jsonhttptest.Request(t, client, http.MethodPost, "/psser/send/to/12", http.StatusBadRequest,
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, hexbatch),
			jsonhttptest.WithRequestBody(bytes.NewReader(payload)),
		)
	})

	t.Run("ok", func(t *testing.T) {
		jsonhttptest.Request(t, client, http.MethodPost, "/psser/send/testtopic/12?recipient="+recipient, http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(bytes.NewReader(payload)),
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: "Created",
				Code:    http.StatusCreated,
			}),
		)
		waitDone(t, &mtx, &done)
		if !bytes.Equal(receivedBytes, payload) {
			t.Fatalf("payload mismatch. want %v got %v", payload, receivedBytes)
		}
		if targets != fmt.Sprint(receivedTargets) {
			t.Fatalf("targets mismatch. want %v got %v", targets, receivedTargets)
		}
		if string(topicHash) != string(receivedTopic[:]) {
			t.Fatalf("topic mismatch. want %v got %v", topic, string(receivedTopic[:]))
		}
	})

	t.Run("without recipient", func(t *testing.T) {
		jsonhttptest.Request(t, client, http.MethodPost, "/psser/send/testtopic/12", http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(bytes.NewReader(payload)),
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: "Created",
				Code:    http.StatusCreated,
			}),
		)
		waitDone(t, &mtx, &done)
		if !bytes.Equal(receivedBytes, payload) {
			t.Fatalf("payload mismatch. want %v got %v", payload, receivedBytes)
		}
		if targets != fmt.Sprint(receivedTargets) {
			t.Fatalf("targets mismatch. want %v got %v", targets, receivedTargets)
		}
		if string(topicHash) != string(receivedTopic[:]) {
			t.Fatalf("topic mismatch. want %v got %v", topic, string(receivedTopic[:]))
		}
	})
}

// TestPssPingPong tests that the websocket api adheres to the websocket standard
// and sends ping-pong messages to keep the connection alive.
// The test opens a websocket, keeps it alive for 500ms, then receives a psser message.
func TestPssPingPong(t *testing.T) {
	var (
		p, publicKey, cl, _ = newPssTest(t, opts{pingPeriod: 90 * time.Millisecond})

		msgContent = make([]byte, len(payload))
		tc         cluster.Chunk
		mtx        sync.Mutex
		pongWait   = 1 * time.Millisecond
		done       = make(chan struct{})
	)

	cl.SetReadLimit(cluster.ChunkSize)
	err := cl.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		t.Fatal(err)
	}
	defer close(done)
	go waitReadMessage(t, &mtx, cl, msgContent, done)

	tc, err = trojan.Wrap(context.Background(), topic, payload, publicKey, targets)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(500 * time.Millisecond) // wait to see that the websocket is kept alive

	p.TryUnwrap(tc)

	waitMessage(t, msgContent, nil, &mtx)
}

func waitReadMessage(t *testing.T, mtx *sync.Mutex, cl *websocket.Conn, targetContent []byte, done <-chan struct{}) {
	t.Helper()
	timeout := time.After(rTimeout)
	for {
		select {
		case <-done:
			return
		case <-timeout:
			t.Error("timed out waiting for message")
			return
		default:
		}

		msgType, message, err := cl.ReadMessage()
		if err != nil {
			return
		}
		if msgType == websocket.PongMessage {
			// ignore pings
			continue
		}

		if message != nil {
			mtx.Lock()
			copy(targetContent, message)
			mtx.Unlock()
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func waitDone(t *testing.T, mtx *sync.Mutex, done *bool) {
	for i := 0; i < 10; i++ {
		mtx.Lock()
		if *done {
			mtx.Unlock()
			return
		}
		mtx.Unlock()
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("timed out waiting for send")
}

func waitMessage(t *testing.T, data, expData []byte, mtx *sync.Mutex) {
	t.Helper()

	ttl := time.After(mTimeout)
	for {
		select {
		case <-ttl:
			if expData == nil {
				return
			}
			t.Fatal("timed out waiting for psser message")
		default:
		}
		mtx.Lock()
		if bytes.Equal(data, expData) {
			mtx.Unlock()
			return
		}
		mtx.Unlock()
		time.Sleep(100 * time.Millisecond)
	}
}

type opts struct {
	pingPeriod time.Duration
}

func newPssTest(t *testing.T, o opts) (psser.Interface, *ecdsa.PublicKey, *websocket.Conn, string) {
	t.Helper()

	privkey, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatal(err)
	}
	var (
		logger = log.Noop
		pss    = psser.New(privkey, logger)
	)
	if o.pingPeriod == 0 {
		o.pingPeriod = 10 * time.Second
	}
	_, cl, listener, _ := newTestServer(t, testServerOptions{
		Pss:          pss,
		WsPath:       "/psser/subscribe/testtopic",
		Storer:       mock.NewStorer(),
		Logger:       logger,
		WsPingPeriod: o.pingPeriod,
	})
	return pss, &privkey.PublicKey, cl, listener
}

type pssSendFn func(context.Context, trojan.Targets, cluster.Chunk) error
type mpss struct {
	f pssSendFn
}

func newMockPss(f pssSendFn) *mpss {
	return &mpss{f}
}

// Send arbitrary byte slice with the given topic to Targets.
func (m *mpss) Send(ctx context.Context, topic trojan.Topic, payload []byte, _ voucher.Stamper, recipient *ecdsa.PublicKey, targets trojan.Targets) error {
	chunk, err := trojan.Wrap(ctx, topic, payload, recipient, targets)
	if err != nil {
		return err
	}
	return m.f(ctx, targets, chunk)
}

// Register a Handler for a given Topic.
func (m *mpss) Register(_ trojan.Topic, _ psser.Handler) func() {
	panic("not implemented") // TODO: Implement
}

// TryUnwrap tries to unwrap a wrapped trojan message.
func (m *mpss) TryUnwrap(_ cluster.Chunk) {
	panic("not implemented") // TODO: Implement
}

func (m *mpss) SetPushSyncer(pushSyncer pushsync.PushSyncer) {
	panic("not implemented") // TODO: Implement
}

func (m *mpss) Close() error {
	panic("not implemented") // TODO: Implement
}
