package storage

import (
	"context"
	"errors"

	"github.com/redesblock/hop/core/swarm"
)

var ErrNotFound = errors.New("storage: not found")

type Storer interface {
	Get(ctx context.Context, addr swarm.Address) (data []byte, err error)
	Put(ctx context.Context, addr swarm.Address, data []byte) error
}
