package netstore_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redesblock/mop/core/flock"
	"github.com/redesblock/mop/core/logging"
	"github.com/redesblock/mop/core/netstore"
	"github.com/redesblock/mop/core/postage"
	postagetesting "github.com/redesblock/mop/core/postage/testing"
	"github.com/redesblock/mop/core/pss"
	"github.com/redesblock/mop/core/recovery"
	"github.com/redesblock/mop/core/sctx"
	"github.com/redesblock/mop/core/storage"
	"github.com/redesblock/mop/core/storage/mock"
)

var chunkData = []byte("mockdata")
var chunkVouch = postagetesting.MustNewVouch()

// TestNetstoreRetrieval verifies that a chunk is asked from the network whenever
// it is not found locally
func TestNetstoreRetrieval(t *testing.T) {
	retrieve, store, nstore := newRetrievingNetstore(t, nil, noopvalidVouch)
	addr := flock.MustParseHexAddress("000001")
	_, err := nstore.Get(context.Background(), storage.ModeGetRequest, addr)
	if err != nil {
		t.Fatal(err)
	}
	if !retrieve.called {
		t.Fatal("retrieve request not issued")
	}
	if retrieve.callCount != 1 {
		t.Fatalf("call count %d", retrieve.callCount)
	}
	if !retrieve.addr.Equal(addr) {
		t.Fatalf("addresses not equal. got %s want %s", retrieve.addr, addr)
	}

	// store should have the chunk once the background PUT is complete
	d := waitAndGetChunk(t, store, addr, storage.ModeGetRequest)

	if !bytes.Equal(d.Data(), chunkData) {
		t.Fatal("chunk data not equal to expected data")
	}

	// check that the second call does not result in another retrieve request
	d, err = nstore.Get(context.Background(), storage.ModeGetRequest, addr)
	if err != nil {
		t.Fatal(err)
	}

	if retrieve.callCount != 1 {
		t.Fatalf("call count %d", retrieve.callCount)
	}
	if !bytes.Equal(d.Data(), chunkData) {
		t.Fatal("chunk data not equal to expected data")
	}

}

// TestNetstoreNoRetrieval verifies that a chunk is not requested from the network
// whenever it is found locally.
func TestNetstoreNoRetrieval(t *testing.T) {
	retrieve, store, nstore := newRetrievingNetstore(t, nil, noopvalidVouch)
	addr := flock.MustParseHexAddress("000001")

	// store should have the chunk in advance
	_, err := store.Put(context.Background(), storage.ModePutUpload, flock.NewChunk(addr, chunkData))
	if err != nil {
		t.Fatal(err)
	}

	c, err := nstore.Get(context.Background(), storage.ModeGetRequest, addr)
	if err != nil {
		t.Fatal(err)
	}
	if retrieve.called {
		t.Fatal("retrieve request issued but shouldn't")
	}
	if retrieve.callCount != 0 {
		t.Fatalf("call count %d", retrieve.callCount)
	}
	if !bytes.Equal(c.Data(), chunkData) {
		t.Fatal("chunk data mismatch")
	}
}

func TestRecovery(t *testing.T) {
	callbackWasCalled := make(chan bool, 1)
	rec := &mockRecovery{
		callbackC: callbackWasCalled,
	}

	retrieve, _, nstore := newRetrievingNetstore(t, rec.recovery, noopvalidVouch)
	addr := flock.MustParseHexAddress("deadbeef")
	retrieve.failure = true
	ctx := context.Background()
	ctx = sctx.SetTargets(ctx, "be, cd")

	_, err := nstore.Get(ctx, storage.ModeGetRequest, addr)
	if err != nil && !errors.Is(err, netstore.ErrRecoveryAttempt) {
		t.Fatal(err)
	}

	select {
	case <-callbackWasCalled:
		break
	case <-time.After(100 * time.Millisecond):
		t.Fatal("recovery callback was not called")
	}
}

func TestInvalidRecoveryFunction(t *testing.T) {
	retrieve, _, nstore := newRetrievingNetstore(t, nil, noopvalidVouch)
	addr := flock.MustParseHexAddress("deadbeef")
	retrieve.failure = true
	ctx := context.Background()
	ctx = sctx.SetTargets(ctx, "be, cd")

	_, err := nstore.Get(ctx, storage.ModeGetRequest, addr)
	if err != nil && err.Error() != "chunk not found" {
		t.Fatal(err)
	}
}

func TestInvalidPostageVouch(t *testing.T) {
	f := func(c flock.Chunk, _ []byte) (flock.Chunk, error) {
		return nil, errors.New("invalid postage vouch")
	}
	retrieve, store, nstore := newRetrievingNetstore(t, nil, f)
	addr := flock.MustParseHexAddress("000001")
	_, err := nstore.Get(context.Background(), storage.ModeGetRequest, addr)
	if err != nil {
		t.Fatal(err)
	}
	if !retrieve.called {
		t.Fatal("retrieve request not issued")
	}
	if retrieve.callCount != 1 {
		t.Fatalf("call count %d", retrieve.callCount)
	}
	if !retrieve.addr.Equal(addr) {
		t.Fatalf("addresses not equal. got %s want %s", retrieve.addr, addr)
	}

	// store should have the chunk once the background PUT is complete
	d := waitAndGetChunk(t, store, addr, storage.ModeGetRequest)

	if !bytes.Equal(d.Data(), chunkData) {
		t.Fatal("chunk data not equal to expected data")
	}

	if mode := store.GetModePut(addr); mode != storage.ModePutRequestCache {
		t.Fatalf("wanted ModePutRequestCache but got %s", mode)
	}

	// check that the second call does not result in another retrieve request
	d, err = nstore.Get(context.Background(), storage.ModeGetRequest, addr)
	if err != nil {
		t.Fatal(err)
	}

	if retrieve.callCount != 1 {
		t.Fatalf("call count %d", retrieve.callCount)
	}
	if !bytes.Equal(d.Data(), chunkData) {
		t.Fatal("chunk data not equal to expected data")
	}
}

func waitAndGetChunk(t *testing.T, store storage.Storer, addr flock.Address, mode storage.ModeGet) flock.Chunk {
	t.Helper()

	start := time.Now()
	for {
		time.Sleep(time.Millisecond * 10)

		d, err := store.Get(context.Background(), mode, addr)
		if err != nil {
			if time.Since(start) > 3*time.Second {
				t.Fatal("waited 3 secs for background put operation", err)
			}
		} else {
			return d
		}
	}
}

// returns a mock retrieval protocol, a mock local storage and a netstore
func newRetrievingNetstore(t *testing.T, rec recovery.Callback, validVouch postage.ValidVouchFn) (ret *retrievalMock, mockStore *mock.MockStorer, ns storage.Storer) {
	retrieve := &retrievalMock{}
	store := mock.NewStorer()
	logger := logging.New(io.Discard, 0)
	ns = netstore.New(store, validVouch, rec, retrieve, logger)
	t.Cleanup(func() {
		err := ns.Close()
		if err != nil {
			t.Fatal("failed closing netstore", err)
		}
	})
	return retrieve, store, ns
}

type retrievalMock struct {
	called    bool
	callCount int32
	failure   bool
	addr      flock.Address
}

func (r *retrievalMock) RetrieveChunk(ctx context.Context, addr flock.Address, orig bool) (chunk flock.Chunk, err error) {
	if r.failure {
		return nil, fmt.Errorf("chunk not found")
	}
	r.called = true
	atomic.AddInt32(&r.callCount, 1)
	r.addr = addr
	return flock.NewChunk(addr, chunkData).WithVouch(chunkVouch), nil
}

type mockRecovery struct {
	callbackC chan bool
}

// Send mocks the pss Send function
func (mr *mockRecovery) recovery(chunkAddress flock.Address, targets pss.Targets) {
	mr.callbackC <- true
}

func (r *mockRecovery) RetrieveChunk(ctx context.Context, addr flock.Address, orig bool) (chunk flock.Chunk, err error) {
	return nil, fmt.Errorf("chunk not found")
}

var noopvalidVouch = func(c flock.Chunk, _ []byte) (flock.Chunk, error) {
	return c, nil
}
