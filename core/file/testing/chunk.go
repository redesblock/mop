package testing

import (
	"encoding/binary"
	"math/rand"

	"github.com/redesblock/mop/core/flock"
)

// GenerateTestRandomFileChunk generates one single chunk with arbitrary content and address
func GenerateTestRandomFileChunk(address flock.Address, spanLength, dataSize int) flock.Chunk {
	data := make([]byte, dataSize+8)
	binary.LittleEndian.PutUint64(data, uint64(spanLength))
	_, _ = rand.Read(data[8:])
	key := make([]byte, flock.SectionSize)
	if address.IsZero() {
		_, _ = rand.Read(key)
	} else {
		copy(key, address.Bytes())
	}
	return flock.NewChunk(flock.NewAddress(key), data)
}
