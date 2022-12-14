package mock

import (
	"context"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/protocol/pushsync"
)

type mock struct {
	sendChunk func(ctx context.Context, chunk cluster.Chunk) (*pushsync.Receipt, error)
}

func New(sendChunk func(ctx context.Context, chunk cluster.Chunk) (*pushsync.Receipt, error)) pushsync.PushSyncer {
	return &mock{sendChunk: sendChunk}
}

func (s *mock) PushChunkToClosest(ctx context.Context, chunk cluster.Chunk) (*pushsync.Receipt, error) {
	return s.sendChunk(ctx, chunk)
}

func (s *mock) Close() error {
	return nil
}
