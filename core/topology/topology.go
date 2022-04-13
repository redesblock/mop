package topology

import (
	"context"
	"errors"

	"github.com/redesblock/hop/core/swarm"
)

var ErrNotFound = errors.New("no peer found")
var ErrWantSelf = errors.New("node wants self")

type Driver interface {
	PeerAdder
	ChunkPeerer
	SyncPeerer
}

type PeerAdder interface {
	AddPeer(ctx context.Context, addr swarm.Address) error
}

type ChunkPeerer interface {
	ChunkPeer(addr swarm.Address) (peerAddr swarm.Address, err error)
}

type SyncPeerer interface {
	SyncPeer(addr swarm.Address) (peerAddr swarm.Address, err error)
}
