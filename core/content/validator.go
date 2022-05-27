package content

import (
	"github.com/redesblock/hop/core/swarm"
)

var _ swarm.ChunkValidator = (*ContentAddressValidator)(nil)

// ContentAddressValidator validates that the address of a given chunk
// is the content address of its contents.
type ContentAddressValidator struct {
}

// NewContentAddressValidator constructs a new ContentAddressValidator
func NewContentAddressValidator() swarm.ChunkValidator {
	return &ContentAddressValidator{}
}

// Validate performs the validation check.
func (v *ContentAddressValidator) Validate(ch swarm.Chunk) (valid bool) {
	chunkData := ch.Data()
	rch, err := contentChunkFromBytes(chunkData)
	if err != nil {
		return false
	}

	address := ch.Address()
	return address.Equal(rch.Address())
}
