package postage

import "math/big"

// ChainState contains data the batch service reads from the chain.
type ChainState struct {
	Block        uint64   // The block number of the last postage event.
	TotalAmount  *big.Int // Cumulative amount paid per stamp.
	CurrentPrice *big.Int // mop/chunk/block normalised price.
}
