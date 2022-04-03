package topology

import (
	"context"
	"errors"

	"github.com/redesblock/hop/core/swarm"
)

var ErrNotFound = errors.New("no peer found")

type Driver interface {
	PeerAdder
	ChunkPeerer
}

type PeerAdder interface {
	AddPeer(ctx context.Context, addr swarm.Address) error
}

type ChunkPeerer interface {
	ChunkPeer(addr swarm.Address) (peerAddr swarm.Address, err error)
}
