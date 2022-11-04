// Package discovery exposes the discovery driver interface
// which is implemented by discovery protocols.
package discovery

import (
	"context"

	"github.com/redesblock/mop/core/cluster"
)

type Driver interface {
	BroadcastPeers(ctx context.Context, addressee cluster.Address, peers ...cluster.Address) error
}
