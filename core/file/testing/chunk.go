package testing

import (
	"encoding/binary"
	"math/rand"

	"github.com/redesblock/mop/core/cluster"
)

// GenerateTestRandomFileChunk generates one single chunk with arbitrary content and address
func GenerateTestRandomFileChunk(address cluster.Address, spanLength, dataSize int) cluster.Chunk {
	data := make([]byte, dataSize+8)
	binary.LittleEndian.PutUint64(data, uint64(spanLength))
	_, _ = rand.Read(data[8:])
	key := make([]byte, cluster.SectionSize)
	if address.IsZero() {
		_, _ = rand.Read(key)
	} else {
		copy(key, address.Bytes())
	}
	return cluster.NewChunk(cluster.NewAddress(key), data)
}
