package testing

import (
	"math/rand"
	"testing"

	"github.com/redesblock/hop/core/postage"
)

// NewChainState will create a new ChainState with random values.
func NewChainState() *postage.ChainState {
	return &postage.ChainState{
		Block: rand.Uint64(), // skipcq: GSC-G404
		Price: NewBigInt(),
		Total: NewBigInt(),
	}
}

// CompareChainState is a test helper that compares two ChainStates and fails
// the test if they are not exactly equal.
// Fails on first difference and returns a descriptive comparison.
func CompareChainState(t *testing.T, want, got *postage.ChainState) {
	t.Helper()

	if want.Block != got.Block {
		t.Fatalf("block: want %v, got %v", want.Block, got.Block)
	}
	if want.Price.Cmp(got.Price) != 0 {
		t.Fatalf("price: want %v, got %v", want.Price, got.Price)
	}
	if want.Total.Cmp(got.Total) != 0 {
		t.Fatalf("total: want %v, got %v", want.Total, got.Total)
	}
}
