package resolver

import (
	"io"

	"github.com/redesblock/hop/core/swarm"
)

// Address is the swarm hop address.
type Address = swarm.Address

// Interface can resolve an URL into an associated Ethereum address.
type Interface interface {
	Resolve(url string) (Address, error)
	io.Closer
}
