package mock

import (
	"github.com/redesblock/hop/core/swarm"
)

var _ swarm.Validator = (*Validator)(nil)

type Validator struct {
	rv bool
}

// NewValidator constructs a new Validator
func NewValidator(rv bool) swarm.Validator {
	return &Validator{rv: rv}
}

// Validate returns rv from mock struct
func (v *Validator) Validate(ch swarm.Chunk) (valid bool) {
	return v.rv
}
