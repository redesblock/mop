package pins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/redesblock/mop/core/chunk/encryption"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/redesblock/mop/core/traverser"
)

// ErrTraversal signals that errors occurred during nodes traverser.
var ErrTraversal = errors.New("traverser iteration failed")

// Interface defines pins operations.
type Interface interface {
	// CreatePin creates a new pins for the given reference.
	// The boolean arguments specifies whether all nodes
	// in the tree should also be traversed and pinned.
	// Repeating calls of this method are idempotent.
	CreatePin(context.Context, cluster.Address, bool) error
	// DeletePin deletes given reference. All the existing
	// nodes in the tree will also be traversed and un-pinned.
	// Repeating calls of this method are idempotent.
	DeletePin(context.Context, cluster.Address) error
	// HasPin returns true if the given reference has root pins.
	HasPin(cluster.Address) (bool, error)
	// Pins return all pinned references.
	Pins() ([]cluster.Address, error)
}

const storePrefix = "root-pins"

func rootPinKey(ref cluster.Address) string {
	return fmt.Sprintf("%s-%s", storePrefix, ref)
}

// NewService is a convenient constructor for Service.
func NewService(
	pinStorage storage.Storer,
	rhStorage storage.StateStorer,
	traverser traverser.Traverser,
) *Service {
	return &Service{
		pinStorage: pinStorage,
		rhStorage:  rhStorage,
		traverser:  traverser,
	}
}

// Service is implementation of the pins.Interface.
type Service struct {
	pinStorage storage.Storer
	rhStorage  storage.StateStorer
	traverser  traverser.Traverser
}

// CreatePin implements Interface.CreatePin method.
func (s *Service) CreatePin(ctx context.Context, ref cluster.Address, traverse bool) error {
	// iterFn is a pins iterator function over the leaves of the root.
	iterFn := func(leaf cluster.Address) error {
		switch err := s.pinStorage.Set(ctx, storage.ModeSetPin, leaf); {
		case errors.Is(err, storage.ErrNotFound):
			ch, err := s.pinStorage.Get(ctx, storage.ModeGetRequestPin, leaf)
			if err != nil {
				return fmt.Errorf("unable to get pins for leaf %q of root %q: %w", leaf, ref, err)
			}
			_, err = s.pinStorage.Put(ctx, storage.ModePutRequestPin, ch)
			if err != nil {
				return fmt.Errorf("unable to put pins for leaf %q of root %q: %w", leaf, ref, err)
			}
		case err != nil:
			return fmt.Errorf("unable to set pins for leaf %q of root %q: %w", leaf, ref, err)
		}
		return nil
	}

	if traverse {
		if err := s.traverser.Traverse(ctx, ref, iterFn); err != nil {
			return fmt.Errorf("traverser of %q failed: %w", ref, err)
		}
	}

	key := rootPinKey(ref)
	switch err := s.rhStorage.Get(key, new(cluster.Address)); {
	case errors.Is(err, storage.ErrNotFound):
		return s.rhStorage.Put(key, ref)
	case err != nil:
		return fmt.Errorf("unable to pins %q: %w", ref, err)
	}
	return nil
}

// DeletePin implements Interface.DeletePin method.
func (s *Service) DeletePin(ctx context.Context, ref cluster.Address) error {
	var iterErr error
	// iterFn is a unpinning iterator function over the leaves of the root.
	iterFn := func(leaf cluster.Address) error {
		if len(leaf.Bytes()) == encryption.ReferenceSize {
			// the traverser service might report back encrypted reference.
			// this is not so trivial to mitigate inside the traverser service
			// since it might introduce complexity with determining which entries
			// should be treated with which address length, since the decryption keys
			// on encrypted references are still needed for correct traverser.
			// we therefore just make sure that localstore gets the correct reference size
			// for unpinning.
			leaf = cluster.NewAddress(leaf.Bytes()[:cluster.HashSize])
		}
		err := s.pinStorage.Set(ctx, storage.ModeSetUnpin, leaf)
		if err != nil {
			iterErr = multierror.Append(err, fmt.Errorf("unable to unpin the chunk for leaf %q of root %q: %w", leaf, ref, err))
			// Continue un-pins all chunks.
		}
		return nil
	}

	if err := s.traverser.Traverse(ctx, ref, iterFn); err != nil {
		return fmt.Errorf("traverser of %q failed: %w", ref, multierror.Append(err, iterErr))
	}
	if iterErr != nil {
		return multierror.Append(ErrTraversal, iterErr)
	}

	key := rootPinKey(ref)
	if err := s.rhStorage.Delete(key); err != nil {
		return fmt.Errorf("unable to delete pins for key %q: %w", key, err)
	}
	return nil
}

// HasPin implements Interface.HasPin method.
func (s *Service) HasPin(ref cluster.Address) (bool, error) {
	key, val := rootPinKey(ref), cluster.NewAddress(nil)
	switch err := s.rhStorage.Get(key, &val); {
	case errors.Is(err, storage.ErrNotFound):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("unable to get pins for key %q: %w", key, err)
	}
	return val.Equal(ref), nil
}

// Pins implements Interface.Pins method.
func (s *Service) Pins() ([]cluster.Address, error) {
	var refs = make([]cluster.Address, 0)
	err := s.rhStorage.Iterate(storePrefix, func(key, val []byte) (stop bool, err error) {
		var ref cluster.Address
		if err := json.Unmarshal(val, &ref); err != nil {
			return true, fmt.Errorf("invalid reference value %q: %w", string(val), err)
		}
		refs = append(refs, ref)
		return false, nil
	})
	if err != nil {
		return nil, fmt.Errorf("iteration failed: %w", err)
	}
	return refs, nil
}
