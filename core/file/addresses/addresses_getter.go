package addresses

import (
	"context"

	"github.com/redesblock/mop/core/flock"
	"github.com/redesblock/mop/core/storage"
)

type addressesGetterStore struct {
	getter storage.Getter
	fn     flock.AddressIterFunc
}

// NewGetter creates a new proxy storage.Getter which calls provided function
// for each chunk address processed.
func NewGetter(getter storage.Getter, fn flock.AddressIterFunc) storage.Getter {
	return &addressesGetterStore{getter, fn}
}

func (s *addressesGetterStore) Get(ctx context.Context, mode storage.ModeGet, addr flock.Address) (flock.Chunk, error) {
	ch, err := s.getter.Get(ctx, mode, addr)
	if err != nil {
		return nil, err
	}

	return ch, s.fn(ch.Address())
}
