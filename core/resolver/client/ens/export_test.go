package ens

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
)

const SwarmContentHashPrefix = swarmContentHashPrefix

var ErrNotImplemented = errNotImplemented

// WithDialFunc will set the Dial function implementaton.
func WithDialFunc(fn func(ep string) (*ethclient.Client, error)) Option {
	return func(c *Client) {
		c.dialFn = fn
	}
}

// WithResolveFunc will set the Resolve function implementation.
func WithResolveFunc(fn func(backend bind.ContractBackend, input string) (string, error)) Option {
	return func(c *Client) {
		c.resolveFn = fn
	}
}
