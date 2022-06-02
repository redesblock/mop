package content

import (
	"github.com/redesblock/hop/core/swarm"
)

var _ swarm.Validator = (*Validator)(nil)

// ContentAddressValidator validates that the address of a given chunk
// is the content address of its contents.
type Validator struct {
}

// NewContentAddressValidator constructs a new ContentAddressValidator
func NewValidator() swarm.Validator {
	return &Validator{}
}

// Validate performs the validation check.
func (v *Validator) Validate(ch swarm.Chunk) (valid bool) {
	chunkData := ch.Data()
	rch, err := contentChunkFromBytes(chunkData)
	if err != nil {
		return false
	}

	address := ch.Address()
	return address.Equal(rch.Address())
}
