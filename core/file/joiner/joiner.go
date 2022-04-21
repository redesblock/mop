// Package joiner provides implementations of the file.Joiner interface
package joiner

import (
	"context"
	"encoding/binary"
	"io"

	"github.com/redesblock/hop/core/file"
	"github.com/redesblock/hop/core/file/joiner/internal"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
)

// simpleJoiner wraps a non-optimized implementation of file.Joiner.
type simpleJoiner struct {
	store storage.Storer
}

// NewSimpleJoiner creates a new simpleJoiner.
func NewSimpleJoiner(store storage.Storer) file.Joiner {
	return &simpleJoiner{
		store: store,
	}
}

// Join implements the file.Joiner interface.
//
// It uses a non-optimized internal component that only retrieves a data chunk
// after the previous has been read.
func (s *simpleJoiner) Join(ctx context.Context, address swarm.Address) (dataOut io.ReadCloser, dataSize int64, err error) {

	// retrieve the root chunk to read the total data length the be retrieved
	rootChunk, err := s.store.Get(ctx, storage.ModeGetRequest, address)
	if err != nil {
		return nil, 0, err
	}

	// if this is a single chunk, short circuit to returning just that chunk
	spanLength := binary.LittleEndian.Uint64(rootChunk.Data())
	if spanLength <= swarm.ChunkSize {
		data := rootChunk.Data()[8:]
		return file.NewSimpleReadCloser(data), int64(spanLength), nil
	}

	r := internal.NewSimpleJoinerJob(ctx, s.store, rootChunk)
	return r, int64(spanLength), nil
}
