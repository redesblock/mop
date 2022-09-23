// Package discovery exposes the discovery driver interface
// which is implemented by discovery protocols.
package discovery

import (
	"context"

	"github.com/redesblock/mop/core/swarm"
)

type Driver interface {
	BroadcastPeers(ctx context.Context, addressee swarm.Address, peers ...swarm.Address) error
}
