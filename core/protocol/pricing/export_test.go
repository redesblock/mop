package pricing

import (
	"context"

	"github.com/redesblock/mop/core/p2p"
)

func (s *Service) Init(ctx context.Context, p p2p.Peer) error {
	return s.init(ctx, p)
}
