package inmemstore

import (
	"context"
	"sync"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/storer/storage"
)

// Store implements a simple Putter and Getter which can be used to temporarily cache
// chunks. Currently this is used in the bootstrapping process of new nodes where
// we chainsync the voucher events from the cluster network.
type Store struct {
	mtx   sync.Mutex
	store map[string]cluster.Chunk
}

func New() *Store {
	return &Store{
		store: make(map[string]cluster.Chunk),
	}
}

func (s *Store) Get(_ context.Context, _ storage.ModeGet, addr cluster.Address) (ch cluster.Chunk, err error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if ch, ok := s.store[addr.ByteString()]; ok {
		return ch, nil
	}

	return nil, storage.ErrNotFound
}

func (s *Store) Put(_ context.Context, _ storage.ModePut, chs ...cluster.Chunk) (exist []bool, err error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for _, ch := range chs {
		s.store[ch.Address().ByteString()] = ch
	}

	exist = make([]bool, len(chs))

	return exist, err
}

func (s *Store) GetMulti(_ context.Context, _ storage.ModeGet, _ ...cluster.Address) (ch []cluster.Chunk, err error) {
	panic("not implemented")
}

func (s *Store) Has(_ context.Context, _ cluster.Address) (yes bool, err error) {
	panic("not implemented")
}

func (s *Store) HasMulti(_ context.Context, _ ...cluster.Address) (yes []bool, err error) {
	panic("not implemented")
}

func (s *Store) Set(_ context.Context, _ storage.ModeSet, _ ...cluster.Address) (err error) {
	panic("not implemented")
}

func (s *Store) LastPullSubscriptionBinID(_ uint8) (id uint64, err error) {
	panic("not implemented")
}

func (s *Store) SubscribePull(_ context.Context, _ uint8, _ uint64, _ uint64) (c <-chan storage.Descriptor, closed <-chan struct{}, stop func()) {
	panic("not implemented")
}

func (s *Store) SubscribePush(_ context.Context, _ func([]byte) bool) (c <-chan cluster.Chunk, repeat func(), stop func()) {
	panic("not implemented")
}

func (s *Store) Close() error {
	return nil
}
