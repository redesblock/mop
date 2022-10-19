package retrieval

import (
	"context"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/p2p"
)

func (s *Service) Handler(ctx context.Context, p p2p.Peer, stream p2p.Stream) error {
	return s.handler(ctx, p, stream)
}

func (s *Service) ClosestPeer(addr cluster.Address, skipPeers []cluster.Address, allowUpstream bool) (cluster.Address, error) {
	return s.closestPeer(addr, skipPeers, allowUpstream)
}
