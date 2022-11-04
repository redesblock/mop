package mock_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/redesblock/mop/core/incentives/voucher/batchstore/mock"
	vouchertesting "github.com/redesblock/mop/core/incentives/voucher/testing"
)

func TestBatchStore(t *testing.T) {
	const testCnt = 3

	testBatch := vouchertesting.MustNewBatch(
		vouchertesting.WithValue(0),
		vouchertesting.WithDepth(0),
	)
	batchStore := mock.New(
		mock.WithGetErr(errors.New("fails"), testCnt),
		mock.WithUpdateErr(errors.New("fails"), testCnt),
	)

	if err := batchStore.Save(testBatch); err != nil {
		t.Fatal("unexpected error")
	}

	// Update should return error after a number of tries:
	for i := 0; i < testCnt; i++ {
		if err := batchStore.Update(testBatch, big.NewInt(0), 0); err != nil {
			t.Fatal(err)
		}
	}
	if err := batchStore.Update(testBatch, big.NewInt(0), 0); err == nil {
		t.Fatal("expected error")
	}

	// Get should fail on wrong id, and after a number of tries:
	if _, err := batchStore.Get(vouchertesting.MustNewID()); err == nil {
		t.Fatal("expected error")
	}
	for i := 0; i < testCnt-1; i++ {
		if _, err := batchStore.Get(testBatch.ID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := batchStore.Get(vouchertesting.MustNewID()); err == nil {
		t.Fatal("expected error")
	}
}

func TestBatchStorePutChainState(t *testing.T) {
	const testCnt = 3

	testChainState := vouchertesting.NewChainState()
	batchStore := mock.New(
		mock.WithChainState(testChainState),
		mock.WithUpdateErr(errors.New("fails"), testCnt),
	)

	// PutChainState should return an error after a number of tries:
	for i := 0; i < testCnt; i++ {
		if err := batchStore.PutChainState(testChainState); err != nil {
			t.Fatal(err)
		}
	}
	if err := batchStore.PutChainState(testChainState); err == nil {
		t.Fatal("expected error")
	}
}

func TestBatchStoreWithBatch(t *testing.T) {
	testBatch := vouchertesting.MustNewBatch()
	batchStore := mock.New(
		mock.WithBatch(testBatch),
	)

	b, err := batchStore.Get(testBatch.ID)
	if err != nil {
		t.Fatal(err)
	}

	vouchertesting.CompareBatches(t, testBatch, b)
}
