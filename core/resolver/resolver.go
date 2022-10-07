package resolver

import (
	"io"

	"github.com/redesblock/mop/core/flock"
)

// Address is the flock mop address.
type Address = flock.Address

// Interface can resolve an URL into an associated Ethereum address.
type Interface interface {
	Resolve(url string) (Address, error)
	io.Closer
}
