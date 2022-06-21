// Package bmtpool provides easy access to binary
// merkle tree hashers managed in as a resource pool.
package bmtpool

import (
	"github.com/redesblock/hop/core/bmt"
	"github.com/redesblock/hop/core/swarm"
)

const Capacity = 32

var instance *bmt.Pool

func init() {
	instance = bmt.NewPool(bmt.NewConf(swarm.NewHasher, swarm.BmtBranches, Capacity))
}

// Get a bmt Hasher instance.
// Instances are reset before being returned to the caller.
func Get() *bmt.Hasher {
	return instance.Get()
}

// Put a bmt Hasher back into the pool
func Put(h *bmt.Hasher) {
	instance.Put(h)
}
