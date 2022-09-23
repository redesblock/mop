package resolver

import (
	"io"

	"github.com/redesblock/mop/core/swarm"
)

// Address is the swarm mop address.
type Address = swarm.Address

// Interface can resolve an URL into an associated Ethereum address.
type Interface interface {
	Resolve(url string) (Address, error)
	io.Closer
}
