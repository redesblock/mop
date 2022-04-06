package storage

import (
	"context"
	"errors"

	"github.com/redesblock/hop/core/swarm"
)

var (
	ErrNotFound     = errors.New("storage: not found")
	ErrInvalidChunk = errors.New("storage: invalid chunk")
)

// ChunkValidatorFunc validates Swarm chunk address and chunk data
type ChunkValidatorFunc func(swarm.Address, []byte) bool

type Storer interface {
	Get(ctx context.Context, addr swarm.Address) (data []byte, err error)
	Put(ctx context.Context, addr swarm.Address, data []byte) error
}
