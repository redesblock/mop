// Package traverser provides abstraction and implementation
// needed to traverse all chunks below a given root hash.
// It tries to parse all manifests and collections in its
// attempt to log all chunk addresses on the way.
package traverser

import (
	"context"
	"errors"
	"fmt"

	"github.com/redesblock/mop/core/chunk/soc"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/file/joiner"
	"github.com/redesblock/mop/core/file/loadsave"
	"github.com/redesblock/mop/core/manifest"
	"github.com/redesblock/mop/core/manifest/mantaray"
	"github.com/redesblock/mop/core/storer/storage"
)

// Traverser represents service which traverse through address dependent chunks.
type Traverser interface {
	// Traverse iterates through each address related to the supplied one, if possible.
	Traverse(context.Context, cluster.Address, cluster.AddressIterFunc) error
}

type PutGetter interface {
	storage.Putter
	storage.Getter
}

// New constructs for a new Traverser.
func New(store PutGetter) Traverser {
	return &service{store: store}
}

// service is implementation of Traverser using storage.Storer as its storage.
type service struct {
	store PutGetter
}

// Traverse implements Traverser.Traverse method.
func (s *service) Traverse(ctx context.Context, addr cluster.Address, iterFn cluster.AddressIterFunc) error {
	processBytes := func(ref cluster.Address) error {
		j, _, err := joiner.New(ctx, s.store, ref)
		if err != nil {
			return fmt.Errorf("traverser: joiner error on %q: %w", ref, err)
		}
		err = j.IterateChunkAddresses(iterFn)
		if err != nil {
			return fmt.Errorf("traverser: iterate chunk address error for %q: %w", ref, err)
		}
		return nil
	}

	ch, err := s.store.Get(ctx, storage.ModeGetRequest, addr)
	if err != nil {
		return fmt.Errorf("traverser: failed to get root chunk %s: %w", addr.String(), err)
	}
	if soc.Valid(ch) {
		// if this is a SOC, the traverser will be just be the single chunk
		return iterFn(addr)
	}

	ls := loadsave.NewReadonly(s.store)
	switch mf, err := manifest.NewDefaultManifestReference(addr, ls); {
	case errors.Is(err, manifest.ErrInvalidManifestType):
		break
	case err != nil:
		return fmt.Errorf("traverser: unable to create manifest reference for %q: %w", addr, err)
	default:
		err := mf.IterateAddresses(ctx, processBytes)
		if errors.Is(err, mantaray.ErrTooShort) || errors.Is(err, mantaray.ErrInvalidVersionHash) {
			// Based on the returned errors we conclude that it might
			// not be a manifest, so we try non-manifest processing.
			break
		}
		if err != nil {
			return fmt.Errorf("traverser: unable to process bytes for %q: %w", addr, err)
		}
		return nil
	}

	// Non-manifest processing.
	if err := processBytes(addr); err != nil {
		return fmt.Errorf("traverser: unable to process bytes for %q: %w", addr, err)
	}
	return nil
}
