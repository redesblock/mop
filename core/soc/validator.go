package soc

import (
	"github.com/redesblock/hop/core/swarm"
)

var _ swarm.Validator = (*Validator)(nil)

// SocVaildator validates that the address of a given chunk
// is a single-owner chunk.
type Validator struct {
}

// NewValidator creates a new Validator.
func NewValidator() swarm.Validator {
	return &Validator{}
}

// Validate performs the validation check.
func (v *Validator) Validate(ch swarm.Chunk) (valid bool) {
	s, err := FromChunk(ch)
	if err != nil {
		return false
	}

	address, err := s.Address()
	if err != nil {
		return false
	}
	return ch.Address().Equal(address)
}
