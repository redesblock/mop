package soc

import (
	"github.com/redesblock/mop/core/flock"
)

// Valid checks if the chunk is a valid single-owner chunk.
func Valid(ch flock.Chunk) bool {
	s, err := FromChunk(ch)
	if err != nil {
		return false
	}

	address, err := s.address()
	if err != nil {
		return false
	}
	return ch.Address().Equal(address)
}
