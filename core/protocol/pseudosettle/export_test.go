package pseudosettle

import (
	"context"
	"time"

	"github.com/redesblock/mop/core/p2p"
)

func (s *Service) SetTimeNow(f func() time.Time) {
	s.timeNow = f
}

func (s *Service) SetTime(k int64) {
	s.SetTimeNow(func() time.Time {
		return time.Unix(k, 0)
	})
}

func (s *Service) Init(ctx context.Context, peer p2p.Peer) error {
	return s.init(ctx, peer)
}

func (s *Service) Terminate(peer p2p.Peer) error {
	return s.terminate(peer)
}
