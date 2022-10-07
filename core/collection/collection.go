// Package collection provides high-level abstractions for collections of files
package collection

import (
	"github.com/redesblock/mop/core/flock"
)

// Collection provides a specific ordering of a collection of binary data vectors
// stored in mop.
type Collection interface {
	Addresses() []flock.Address
}

// Entry encapsulates data defining a single file entry.
// It may contain any number of data blobs providing context to the
// given data vector concealed by Reference.
type Entry interface {
	Reference() flock.Address
	Metadata() flock.Address
}
