package mock

import (
	"fmt"

	"github.com/redesblock/hop/core/resolver"
)

// Assure mock Resolver implements the Resolver interface.
var _ resolver.Interface = (*Resolver)(nil)

// Resolver is the mock Resolver implementation.
type Resolver struct {
	resolveFunc func(string) (resolver.Address, error)
}

// Option function sets the option on the mock Resolver.
type Option func(*Resolver)

// NewResolver will create a new mock Resolver.
func NewResolver(opts ...Option) resolver.Interface {
	r := &Resolver{}

	// Apply all options.
	for _, o := range opts {
		o(r)
	}

	return r
}

// WithResolveFunc will override the Resolve function implementation.
func WithResolveFunc(f func(string) (resolver.Address, error)) Option {
	return func(r *Resolver) {
		r.resolveFunc = f
	}
}

// Resolve implements the Resolver interface.
func (r *Resolver) Resolve(name string) (resolver.Address, error) {
	if r.resolveFunc != nil {
		return r.resolveFunc(name)
	}
	return resolver.Address{}, fmt.Errorf("not implemented")
}
