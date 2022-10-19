package mock

import (
	"context"
	"sync"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/protocol/pullsync/pullstorage"
	"github.com/redesblock/mop/core/storer/storage"
)

var _ pullstorage.Storer = (*PullStorage)(nil)

type chunksResponse struct {
	chunks  []cluster.Address
	topmost uint64
	err     error
}

// WithIntervalsResp mocks a desired response when calling IntervalChunks method.
// Different possible responses for subsequent responses in multi-call scenarios
// are possible (i.e. first call yields a,b,c, second call yields d,e,f).
// Mock maintains state of current call using chunksCalls counter.
func WithIntervalsResp(addrs []cluster.Address, top uint64, err error) Option {
	return optionFunc(func(p *PullStorage) {
		p.intervalChunksResponses = append(p.intervalChunksResponses, chunksResponse{chunks: addrs, topmost: top, err: err})
	})
}

// WithChunks mocks the set of chunks that the store is aware of (used in Get and Has calls).
func WithChunks(chs ...cluster.Chunk) Option {
	return optionFunc(func(p *PullStorage) {
		for _, c := range chs {
			c := c
			p.chunks[c.Address().String()] = c
		}
	})
}

// WithEvilChunk allows to inject a malicious chunk (request a certain address
// of a chunk, but get another), in order to mock unsolicited chunk delivery.
func WithEvilChunk(addr cluster.Address, ch cluster.Chunk) Option {
	return optionFunc(func(p *PullStorage) {
		p.evilAddr = addr
		p.evilChunk = ch
	})
}

func WithCursors(c []uint64) Option {
	return optionFunc(func(p *PullStorage) {
		p.cursors = c
	})
}

func WithCursorsErr(e error) Option {
	return optionFunc(func(p *PullStorage) {
		p.cursorsErr = e
	})
}

type PullStorage struct {
	mtx         sync.Mutex
	chunksCalls int
	putCalls    int
	setCalls    int

	chunks    map[string]cluster.Chunk
	evilAddr  cluster.Address
	evilChunk cluster.Chunk

	cursors    []uint64
	cursorsErr error

	intervalChunksResponses []chunksResponse
}

// NewPullStorage returns a new PullStorage mock.
func NewPullStorage(opts ...Option) *PullStorage {
	s := &PullStorage{
		chunks: make(map[string]cluster.Chunk),
	}
	for _, v := range opts {
		v.apply(s)
	}
	return s
}

// IntervalChunks returns a set of chunk in a requested interval.
func (s *PullStorage) IntervalChunks(_ context.Context, bin uint8, from, to uint64, limit int) (chunks []cluster.Address, topmost uint64, err error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	r := s.intervalChunksResponses[s.chunksCalls]
	s.chunksCalls++

	return r.chunks, r.topmost, r.err
}

func (s *PullStorage) Cursors(ctx context.Context) (curs []uint64, err error) {
	return s.cursors, s.cursorsErr
}

// PutCalls returns the amount of times Put was called.
func (s *PullStorage) PutCalls() int {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	return s.putCalls
}

// SetCalls returns the amount of times Set was called.
func (s *PullStorage) SetCalls() int {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.setCalls
}

// Get chunks.
func (s *PullStorage) Get(_ context.Context, _ storage.ModeGet, addrs ...cluster.Address) (chs []cluster.Chunk, err error) {
	for _, a := range addrs {
		if s.evilAddr.Equal(a) {
			//inject the malicious chunk instead
			chs = append(chs, s.evilChunk)
			continue
		}

		if v, ok := s.chunks[a.String()]; ok {
			chs = append(chs, v)
		} else if !ok {
			return nil, storage.ErrNotFound
		}
	}
	return chs, nil
}

// Put chunks.
func (s *PullStorage) Put(_ context.Context, _ storage.ModePut, chs ...cluster.Chunk) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	for _, c := range chs {
		c := c
		s.chunks[c.Address().String()] = c
	}
	s.putCalls++
	return nil
}

// Set chunks.
func (s *PullStorage) Set(ctx context.Context, mode storage.ModeSet, addrs ...cluster.Address) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.setCalls++
	return nil
}

// Has chunks.
func (s *PullStorage) Has(_ context.Context, addr cluster.Address) (bool, error) {
	if _, ok := s.chunks[addr.String()]; !ok {
		return false, nil
	}
	return true, nil
}

type Option interface {
	apply(*PullStorage)
}
type optionFunc func(*PullStorage)

func (f optionFunc) apply(r *PullStorage) { f(r) }
