// Package bmtpool provides easy access to binary
// merkle tree hashers managed in as a resource pool.
package bmt

import (
	"github.com/redesblock/mop/core/cluster"
)

const Capacity = 32

var instance *Pool

func init() {
	instance = NewPool(NewConf(cluster.NewHasher, cluster.BmtBranches, Capacity))
}

// Get a bmt Hasher instance.
// Instances are reset before being returned to the caller.
func Get() *Hasher {
	return instance.Get()
}

// Put a bmt Hasher back into the pool
func Put(h *Hasher) {
	instance.Put(h)
}
