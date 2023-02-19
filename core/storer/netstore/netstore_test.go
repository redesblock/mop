package netstore_test

import (
	"bytes"
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/incentives/voucher"
	vouchertesting "github.com/redesblock/mop/core/incentives/voucher/testing"
	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/storer/netstore"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/redesblock/mop/core/storer/storage/mock"
	chunktesting "github.com/redesblock/mop/core/storer/storage/testing"
)

var testChunk = chunktesting.GenerateTestRandomChunk()
var chunkStamp = vouchertesting.MustNewStamp()

// TestNetstoreRetrieval verifies that a chunk is asked from the network whenever
// it is not found locally
func TestNetstoreRetrieval(t *testing.T) {
	retrieve, store, nstore := newRetrievingNetstore(t, noopValidStamp)
	addr := testChunk.Address()
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

	if !bytes.Equal(d.Data(), testChunk.Data()) {
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
	if !bytes.Equal(d.Data(), testChunk.Data()) {
		t.Fatal("chunk data not equal to expected data")
	}

}

// TestNetstoreNoRetrieval verifies that a chunk is not requested from the network
// whenever it is found locally.
func TestNetstoreNoRetrieval(t *testing.T) {
	retrieve, store, nstore := newRetrievingNetstore(t, noopValidStamp)
	addr := testChunk.Address()

	// store should have the chunk in advance
	_, err := store.Put(context.Background(), storage.ModePutUpload, testChunk)
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
	if !bytes.Equal(c.Data(), testChunk.Data()) {
		t.Fatal("chunk data mismatch")
	}
}

func TestInvalidChunkNetstoreRetrieval(t *testing.T) {
	retrieve, store, nstore := newRetrievingNetstore(t, noopValidStamp)

	invalidChunk := cluster.NewChunk(testChunk.Address(), []byte("deadbeef"))
	// store invalid chunk, i.e. hash doesnt match the data to simulate corruption
	_, err := store.Put(context.Background(), storage.ModePutUpload, invalidChunk)
	if err != nil {
		t.Fatal(err)
	}

	addr := testChunk.Address()
	_, err = nstore.Get(context.Background(), storage.ModeGetRequest, addr)
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

	if !bytes.Equal(d.Data(), testChunk.Data()) {
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
	if !bytes.Equal(d.Data(), testChunk.Data()) {
		t.Fatal("chunk data not equal to expected data")
	}
}

func TestInvalidVoucherStamp(t *testing.T) {
	f := func(c cluster.Chunk, _ []byte) (cluster.Chunk, error) {
		return nil, errors.New("invalid voucher stamp")
	}
	retrieve, store, nstore := newRetrievingNetstore(t, f)
	addr := testChunk.Address()
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

	if !bytes.Equal(d.Data(), testChunk.Data()) {
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
	if !bytes.Equal(d.Data(), testChunk.Data()) {
		t.Fatal("chunk data not equal to expected data")
	}
}

func waitAndGetChunk(t *testing.T, store storage.Storer, addr cluster.Address, mode storage.ModeGet) cluster.Chunk {
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
func newRetrievingNetstore(t *testing.T, validStamp voucher.ValidStampFn) (ret *retrievalMock, mockStore *mock.MockStorer, ns storage.Storer) {
	retrieve := &retrievalMock{}
	store := mock.NewStorer()
	logger := log.Noop
	ns = netstore.New(store, validStamp, retrieve, logger, false)
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
	addr      cluster.Address
}

func (r *retrievalMock) RetrieveChunk(ctx context.Context, addr, sourceAddr cluster.Address) (chunk cluster.Chunk, err error) {
	if r.failure {
		return nil, errors.New("chunk not found")
	}
	r.called = true
	atomic.AddInt32(&r.callCount, 1)
	r.addr = addr
	return testChunk.WithStamp(chunkStamp), nil
}

var noopValidStamp = func(c cluster.Chunk, _ []byte) (cluster.Chunk, error) {
	return c, nil
}
