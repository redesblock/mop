package testing

import (
	"math/rand"
	"time"

	"github.com/redesblock/hop/core/swarm"
)

func init() {
	// needed for GenerateTestRandomChunk
	rand.Seed(time.Now().UnixNano())
}

// GenerateTestRandomChunk generates a Chunk that is not
// valid, but it contains a random key and a random value.
// This function is faster then storage.GenerateRandomChunk
// which generates a valid chunk.
// Some tests in do not need valid chunks, just
// random data, and their execution time can be decreased
// using this function.
func GenerateTestRandomChunk() swarm.Chunk {
	data := make([]byte, swarm.ChunkSize)
	rand.Read(data)
	key := make([]byte, 32)
	rand.Read(key)
	return swarm.NewChunk(swarm.NewAddress(key), data)
}

// GenerateTestRandomChunks generates a slice of random
// Chunks by using GenerateTestRandomChunk function.
func GenerateTestRandomChunks(count int) []swarm.Chunk {
	chunks := make([]swarm.Chunk, count)
	for i := 0; i < count; i++ {
		chunks[i] = GenerateTestRandomChunk()
	}
	return chunks
}
