// Package bmtpool provides easy access to binary
// merkle tree hashers managed in as a resource pool.
package bmtpool

import (
	"github.com/redesblock/mop/core/bmt"
	"github.com/redesblock/mop/core/flock"
)

const Capacity = 32

var instance *bmt.Pool

func init() {
	instance = bmt.NewPool(bmt.NewConf(flock.NewHasher, flock.BmtBranches, Capacity))
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
