package mock

import (
	"context"

	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
)

type mockStorer struct {
	store     map[string][]byte
	validator storage.ChunkValidatorFunc
}

func NewStorer() storage.Storer {
	return &mockStorer{
		store: make(map[string][]byte),
	}
}

func NewValidatingStorer(f storage.ChunkValidatorFunc) storage.Storer {
	return &mockStorer{
		store:     make(map[string][]byte),
		validator: f,
	}
}

func (m *mockStorer) Get(ctx context.Context, addr swarm.Address) (data []byte, err error) {
	v, has := m.store[addr.String()]
	if !has {
		return nil, storage.ErrNotFound
	}
	return v, nil
}

func (m *mockStorer) Put(ctx context.Context, addr swarm.Address, data []byte) error {
	if m.validator != nil {
		if !m.validator(addr, data) {
			return storage.ErrInvalidChunk
		}
	}
	m.store[addr.String()] = data
	return nil
}
