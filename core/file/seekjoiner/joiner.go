// Package joiner provides implementations of the file.Joiner interface
package seekjoiner

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/redesblock/hop/core/encryption/store"
	"github.com/redesblock/hop/core/file"
	"github.com/redesblock/hop/core/file/seekjoiner/internal"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
)

// simpleJoiner wraps a non-optimized implementation of file.SeekJoiner.
type simpleJoiner struct {
	getter storage.Getter
}

// NewSimpleJoiner creates a new simpleJoiner.
func NewSimpleJoiner(getter storage.Getter) file.JoinSeeker {
	return &simpleJoiner{
		getter: store.New(getter),
	}
}

func (s *simpleJoiner) Size(ctx context.Context, address swarm.Address) (int64, error) {
	// retrieve the root chunk to read the total data length the be retrieved
	rootChunk, err := s.getter.Get(ctx, storage.ModeGetRequest, address)
	if err != nil {
		return 0, err
	}

	chunkData := rootChunk.Data()
	if l := len(chunkData); l < 8 {
		return 0, fmt.Errorf("invalid chunk content of %d bytes", l)
	}

	dataLength := binary.LittleEndian.Uint64(chunkData[:8])
	return int64(dataLength), nil
}

// Join implements the file.Joiner interface.
//
// It uses a non-optimized internal component that only retrieves a data chunk
// after the previous has been read.
func (s *simpleJoiner) Join(ctx context.Context, address swarm.Address) (dataOut io.ReadSeeker, dataSize int64, err error) {
	return internal.NewSimpleJoiner(ctx, s.getter, address)
}
