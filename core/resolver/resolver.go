package resolver

import (
	"errors"
	"io"

	"github.com/redesblock/mop/core/cluster"
)

// Address is the cluster mop address.
type Address = cluster.Address

var (
	// ErrParse denotes failure to parse given value
	ErrParse = errors.New("could not parse")
	// ErrNotFound denotes that given name was not found
	ErrNotFound = errors.New("not found")
	// ErrServiceNotAvailable denotes that remote ENS service is not available
	ErrServiceNotAvailable = errors.New("not available")
	// ErrInvalidContentHash denotes that the value of the response contenthash record is not valid.
	ErrInvalidContentHash = errors.New("invalid cluster content hash")
)

// Interface can resolve an URL into an associated BNB Smart Chain.
type Interface interface {
	Resolve(url string) (Address, error)
	io.Closer
}
