package batchstore_test

import (
	"errors"
	"math/rand"
	"testing"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/incentives/voucher"
	"github.com/redesblock/mop/core/incentives/voucher/batchstore"
	vouchertest "github.com/redesblock/mop/core/incentives/voucher/testing"
	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/storer/statestore/leveldb"
	"github.com/redesblock/mop/core/storer/statestore/mock"
	"github.com/redesblock/mop/core/storer/storage"
)

var noopEvictFn = func([]byte) error { return nil }

func TestBatchStore_Get(t *testing.T) {
	testBatch := vouchertest.MustNewBatch()

	stateStore := mock.NewStateStore()
	batchStore, _ := batchstore.New(stateStore, nil, log.Noop)

	err := batchStore.Save(testBatch)
	if err != nil {
		t.Fatal(err)
	}

	got := batchStoreGetBatch(t, batchStore, testBatch.ID)
	vouchertest.CompareBatches(t, testBatch, got)
}

func TestBatchStore_Iterate(t *testing.T) {
	testBatch := vouchertest.MustNewBatch()
	key := batchstore.BatchKey(testBatch.ID)

	stateStore := mock.NewStateStore()
	batchStore, _ := batchstore.New(stateStore, nil, log.Noop)

	stateStorePut(t, stateStore, key, testBatch)

	var got *voucher.Batch
	err := batchStore.Iterate(func(b *voucher.Batch) (bool, error) {
		got = b
		return false, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	vouchertest.CompareBatches(t, testBatch, got)
}

func TestBatchStore_IterateStopsEarly(t *testing.T) {
	testBatch1 := vouchertest.MustNewBatch()
	key1 := batchstore.BatchKey(testBatch1.ID)

	testBatch2 := vouchertest.MustNewBatch()
	key2 := batchstore.BatchKey(testBatch2.ID)

	stateStore := mock.NewStateStore()
	batchStore, _ := batchstore.New(stateStore, nil, log.Noop)

	stateStorePut(t, stateStore, key1, testBatch1)
	stateStorePut(t, stateStore, key2, testBatch2)

	var iterations = 0
	err := batchStore.Iterate(func(b *voucher.Batch) (bool, error) {
		iterations++
		return false, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if iterations != 2 {
		t.Fatalf("wanted 2 iteration, got %d", iterations)
	}

	iterations = 0
	err = batchStore.Iterate(func(b *voucher.Batch) (bool, error) {
		iterations++
		return true, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if iterations > 2 {
		t.Fatalf("wanted 1 iteration, got %d", iterations)
	}

	iterations = 0
	err = batchStore.Iterate(func(b *voucher.Batch) (bool, error) {
		iterations++
		return false, errors.New("test error")
	})
	if err == nil {
		t.Fatalf("wanted error")
	}
	if iterations > 2 {
		t.Fatalf("wanted 1 iteration, got %d", iterations)
	}
}

func TestBatchStore_SaveAndUpdate(t *testing.T) {

	testBatch := vouchertest.MustNewBatch()
	key := batchstore.BatchKey(testBatch.ID)

	stateStore := mock.NewStateStore()
	batchStore, _ := batchstore.New(stateStore, nil, log.Noop)

	if err := batchStore.Save(testBatch); err != nil {
		t.Fatalf("storer.Save(...): unexpected error: %v", err)
	}

	// call Unreserve once to increase storage radius of the test batch
	if err := batchStore.Unreserve(func(id []byte, radius uint8) (bool, error) { return false, nil }); err != nil {
		t.Fatalf("storer.Unreserve(...): unexpected error: %v", err)
	}

	//get test batch after save call
	stateStoreGet(t, stateStore, key, testBatch)

	var have voucher.Batch
	stateStoreGet(t, stateStore, key, &have)
	vouchertest.CompareBatches(t, testBatch, &have)

	// Check for idempotency.
	if err := batchStore.Save(testBatch); err == nil {
		t.Fatalf("storer.Save(...): expected error")
	}

	cnt := 0
	if err := stateStore.Iterate(batchstore.ValueKey(testBatch.Value, testBatch.ID), func(k, v []byte) (stop bool, err error) {
		cnt++
		return false, nil
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cnt > 1 {
		t.Fatal("storer.Save(...): method is not idempotent")
	}

	// Check update.
	newValue := vouchertest.NewBigInt()
	newDepth := uint8(rand.Intn(int(cluster.MaxPO)))
	if err := batchStore.Update(testBatch, newValue, newDepth); err != nil {
		t.Fatalf("storer.Update(...): unexpected error: %v", err)
	}
	stateStoreGet(t, stateStore, key, &have)
	vouchertest.CompareBatches(t, testBatch, &have)
}

func TestBatchStore_GetChainState(t *testing.T) {
	testChainState := vouchertest.NewChainState()

	stateStore := mock.NewStateStore()
	batchStore, _ := batchstore.New(stateStore, nil, log.Noop)

	err := batchStore.PutChainState(testChainState)
	if err != nil {
		t.Fatal(err)
	}
	got := batchStore.GetChainState()
	vouchertest.CompareChainState(t, testChainState, got)
}

func TestBatchStore_PutChainState(t *testing.T) {
	testChainState := vouchertest.NewChainState()

	stateStore := mock.NewStateStore()
	batchStore, _ := batchstore.New(stateStore, nil, log.Noop)

	batchStorePutChainState(t, batchStore, testChainState)
	var got voucher.ChainState
	stateStoreGet(t, stateStore, batchstore.StateKey, &got)
	vouchertest.CompareChainState(t, testChainState, &got)
}

func TestBatchStore_SetStorageRadius(t *testing.T) {

	var (
		radius           uint8 = 5
		oldStorageRadius uint8 = 5
		newStorageRadius uint8 = 3
	)

	stateStore := mock.NewStateStore()
	_ = stateStore.Put(batchstore.ReserveStateKey, &voucher.ReserveState{Radius: radius})
	batchStore, _ := batchstore.New(stateStore, nil, log.Noop)

	_ = batchStore.SetStorageRadius(func(uint8) uint8 {
		return oldStorageRadius
	})

	_ = batchStore.SetStorageRadius(func(radius uint8) uint8 {
		if radius != oldStorageRadius {
			t.Fatalf("got old radius %d, want %d", radius, oldStorageRadius)
		}
		return newStorageRadius
	})

	got := batchStore.GetReserveState().StorageRadius
	if got != newStorageRadius {
		t.Fatalf("got old radius %d, want %d", got, newStorageRadius)
	}
}

func TestBatchStore_Reset(t *testing.T) {
	testChainState := vouchertest.NewChainState()
	testBatch := vouchertest.MustNewBatch(
		vouchertest.WithValue(15),
		vouchertest.WithDepth(8),
	)

	path := t.TempDir()
	logger := log.Noop

	// we use the real statestore since the mock uses a mutex,
	// therefore deleting while iterating (in Reset() implementation)
	// leads to a deadlock.
	stateStore, err := leveldb.NewStateStore(path, logger)
	if err != nil {
		t.Fatal(err)
	}
	defer stateStore.Close()

	batchStore, _ := batchstore.New(stateStore, noopEvictFn, log.Noop)
	err = batchStore.Save(testBatch)
	if err != nil {
		t.Fatal(err)
	}
	err = batchStore.PutChainState(testChainState)
	if err != nil {
		t.Fatal(err)
	}
	err = batchStore.Reset()
	if err != nil {
		t.Fatal(err)
	}
	c := 0
	_ = stateStore.Iterate("", func(k, _ []byte) (bool, error) {
		c++
		return false, nil
	})

	// we expect one key in the statestore since the schema name
	// will always be there.
	if c != 1 {
		t.Fatalf("expected only one key in statestore, got %d", c)
	}
}

func stateStoreGet(t *testing.T, st storage.StateStorer, k string, v interface{}) {
	t.Helper()

	if err := st.Get(k, v); err != nil {
		t.Fatalf("store get batch: %v", err)
	}
}

func stateStorePut(t *testing.T, st storage.StateStorer, k string, v interface{}) {
	t.Helper()

	if err := st.Put(k, v); err != nil {
		t.Fatalf("store put batch: %v", err)
	}
}

func batchStoreGetBatch(t *testing.T, st voucher.Storer, id []byte) *voucher.Batch {
	t.Helper()

	b, err := st.Get(id)
	if err != nil {
		t.Fatalf("voucher storer get: %v", err)
	}
	return b
}

func batchStorePutChainState(t *testing.T, st voucher.Storer, cs *voucher.ChainState) {
	t.Helper()

	if err := st.PutChainState(cs); err != nil {
		t.Fatalf("voucher storer put chain state: %v", err)
	}
}
