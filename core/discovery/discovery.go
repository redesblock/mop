package discovery

import (
	"context"

	"github.com/redesblock/hop/core/swarm"
)

type Driver interface {
	BroadcastPeers(ctx context.Context, addressee swarm.Address, peers ...swarm.Address) error
}
