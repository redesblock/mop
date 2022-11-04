package testing

import (
	"math/rand"
	"testing"

	"github.com/redesblock/mop/core/incentives/voucher"
)

// NewChainState will create a new ChainState with random values.
func NewChainState() *voucher.ChainState {
	return &voucher.ChainState{
		Block:        rand.Uint64(),
		CurrentPrice: NewBigInt(),
		TotalAmount:  NewBigInt(),
	}
}

// CompareChainState is a test helper that compares two ChainStates and fails
// the test if they are not exactly equal.
// Fails on first difference and returns a descriptive comparison.
func CompareChainState(t *testing.T, want, got *voucher.ChainState) {
	t.Helper()

	if want.Block != got.Block {
		t.Fatalf("block: want %v, got %v", want.Block, got.Block)
	}
	if want.CurrentPrice.Cmp(got.CurrentPrice) != 0 {
		t.Fatalf("price: want %v, got %v", want.CurrentPrice, got.CurrentPrice)
	}
	if want.TotalAmount.Cmp(got.TotalAmount) != 0 {
		t.Fatalf("total: want %v, got %v", want.TotalAmount, got.TotalAmount)
	}
}
