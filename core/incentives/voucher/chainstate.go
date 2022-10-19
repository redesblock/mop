package voucher

import "math/big"

// ChainState contains data the batch service reads from the chain.
type ChainState struct {
	Block        uint64   // The block number of the last voucher event.
	TotalAmount  *big.Int // Cumulative amount paid per stamp.
	CurrentPrice *big.Int // Mop/chunk/block normalised price.
}
