// Package wardeness provides convenience methods
// for reseeding content on Cluster.
package warden

import (
	"context"
	"errors"
	"fmt"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/p2p/topology"
	"github.com/redesblock/mop/core/protocol/pushsync"
	"github.com/redesblock/mop/core/protocol/retrieval"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/redesblock/mop/core/traverser"
	"golang.org/x/sync/errgroup"
)

// how many parallel push operations
const parallelPush = 5

type Interface interface {
	// Reupload root hash and all of its underlying
	// associated chunks to the network.
	Reupload(context.Context, cluster.Address) error

	// IsRetrievable checks whether the content
	// on the given address is retrievable.
	IsRetrievable(context.Context, cluster.Address) (bool, error)
}

type warden struct {
	getter       storage.Getter
	push         pushsync.PushSyncer
	traverser    traverser.Traverser
	netTraverser traverser.Traverser
}

func New(getter storage.Getter, t traverser.Traverser, r retrieval.Interface, p pushsync.PushSyncer) Interface {
	return &warden{
		getter:       getter,
		push:         p,
		traverser:    t,
		netTraverser: traverser.New(&netGetter{r}),
	}
}

// Reupload content with the given root hash to the network.
// The service will automatically dereference and traverse all
// addresses and push every chunk individually to the network.
// It assumes all chunks are available locally. It is therefore
// advisable to pins the content locally before trying to reupload it.
func (s *warden) Reupload(ctx context.Context, root cluster.Address) error {
	sem := make(chan struct{}, parallelPush)
	eg, _ := errgroup.WithContext(ctx)
	fn := func(addr cluster.Address) error {
		c, err := s.getter.Get(ctx, storage.ModeGetSync, addr)
		if err != nil {
			return err
		}

		sem <- struct{}{}
		eg.Go(func() error {
			defer func() { <-sem }()
			_, err := s.push.PushChunkToClosest(ctx, c)
			if err != nil {
				if !errors.Is(err, topology.ErrWantSelf) {
					return err
				}
				// swallow the error in case we are the closest node
			}
			return nil
		})
		return nil
	}

	if err := s.traverser.Traverse(ctx, root, fn); err != nil {
		return fmt.Errorf("traverser of %s failed: %w", root.String(), err)
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("push error during reupload: %w", err)
	}
	return nil
}

// IsRetrievable implements Interface.IsRetrievable method.
func (s *warden) IsRetrievable(ctx context.Context, root cluster.Address) (bool, error) {
	noop := func(leaf cluster.Address) error { return nil }
	switch err := s.netTraverser.Traverse(ctx, root, noop); {
	case errors.Is(err, storage.ErrNotFound):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("traverser of %q failed: %w", root, err)
	default:
		return true, nil
	}
}

// netGetter implements the storage Getter.Get method in a way
// that it will try to retrieve the chunk only from the network.
type netGetter struct {
	retrieval retrieval.Interface
}

// Get implements the storage Getter.Get interface.
func (ng *netGetter) Get(ctx context.Context, _ storage.ModeGet, addr cluster.Address) (cluster.Chunk, error) {
	return ng.retrieval.RetrieveChunk(ctx, addr, cluster.ZeroAddress)
}

// Put implements the storage Putter.Put interface.
func (ng *netGetter) Put(_ context.Context, _ storage.ModePut, _ ...cluster.Chunk) ([]bool, error) {
	return nil, errors.New("operation is not supported")
}
