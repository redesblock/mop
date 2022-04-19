package mock

import (
	"context"

	"github.com/redesblock/hop/core/pushsync"
	"github.com/redesblock/hop/core/swarm"
)

type PushSync struct {
	sendChunk func(ctx context.Context, chunk swarm.Chunk) (*pushsync.Receipt, error)
}

func New(sendChunk func(ctx context.Context, chunk swarm.Chunk) (*pushsync.Receipt, error)) *PushSync {
	return &PushSync{sendChunk: sendChunk}
}

func (s *PushSync) PushChunkToClosest(ctx context.Context, chunk swarm.Chunk) (*pushsync.Receipt, error) {
	return s.sendChunk(ctx, chunk)
}
