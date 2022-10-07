// Package discovery exposes the discovery driver interface
// which is implemented by discovery protocols.
package discovery

import (
	"context"

	"github.com/redesblock/mop/core/flock"
)

type Driver interface {
	BroadcastPeers(ctx context.Context, addressee flock.Address, peers ...flock.Address) error
}
