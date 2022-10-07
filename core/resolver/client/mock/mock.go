package mock

import (
	"github.com/redesblock/mop/core/flock"
	"github.com/redesblock/mop/core/resolver/client"
)

// Ensure mock Client implements the Client interface.
var _ client.Interface = (*Client)(nil)

// Client is the mock resolver client implementation.
type Client struct {
	isConnected    bool
	endpoint       string
	defaultAddress flock.Address
	resolveFn      func(string) (flock.Address, error)
}

// Option is a function that applies an option to a Client.
type Option func(*Client)

// NewClient construct a new mock Client.
func NewClient(opts ...Option) *Client {
	cl := &Client{}

	for _, o := range opts {
		o(cl)
	}

	cl.isConnected = true
	return cl
}

// WithEndpoint will set the endpoint.
func WithEndpoint(endpoint string) Option {
	return func(cl *Client) {
		cl.endpoint = endpoint
	}
}

// WitResolveAddress will set the address returned by Resolve.
func WitResolveAddress(addr flock.Address) Option {
	return func(cl *Client) {
		cl.defaultAddress = addr
	}
}

// WithResolveFunc will set the Resolve function implementation.
func WithResolveFunc(fn func(string) (flock.Address, error)) Option {
	return func(cl *Client) {
		cl.resolveFn = fn
	}
}

// IsConnected is the mock IsConnected implementation.
func (cl *Client) IsConnected() bool {
	return cl.isConnected
}

// Endpoint is the mock Endpoint implementation.
func (cl *Client) Endpoint() string {
	return cl.endpoint
}

// Resolve is the mock Resolve implementation
func (cl *Client) Resolve(name string) (flock.Address, error) {
	if cl.resolveFn == nil {
		return cl.defaultAddress, nil
	}
	return cl.resolveFn(name)
}

// Close is the mock Close implementation.
func (cl *Client) Close() error {
	cl.isConnected = false
	return nil
}
