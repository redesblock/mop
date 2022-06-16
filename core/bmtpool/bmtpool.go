package bmtpool

import (
	bmtlegacy "github.com/ethersphere/bmt/legacy"
	"github.com/ethersphere/bmt/pool"
	"github.com/redesblock/hop/core/swarm"
)

var instance pool.Pooler

func init() {
	instance = pool.New(8, swarm.BmtBranches)
}

// Get a bmt Hasher instance.
// Instances are reset before being returned to the caller.
func Get() *bmtlegacy.Hasher {
	return instance.Get()
}

// Put a bmt Hasher back into the pool
func Put(h *bmtlegacy.Hasher) {
	instance.Put(h)
}
