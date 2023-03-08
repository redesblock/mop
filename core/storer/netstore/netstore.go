// Package netstore provides an abstraction layer over the
// Cluster local storage layer that leverages connectivity
// with other peers in order to retrieve chunks from the network that cannot
// be found locally.
package netstore

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/redesblock/mop/core/chunk/cac"
	"github.com/redesblock/mop/core/chunk/soc"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/incentives/voucher"
	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/protocol/retrieval"
	"github.com/redesblock/mop/core/storer/storage"
)

// loggerName is the tree path name of the logger for this package.
const loggerName = "netstore"

const (
	maxBgPutters int = 16 * 8
)

type store struct {
	storage.Storer
	retrieval  retrieval.Interface
	logger     log.Logger
	validStamp voucher.ValidStampFn
	bgWorkers  chan struct{}
	sCtx       context.Context
	sCancel    context.CancelFunc
	wg         sync.WaitGroup
	metrics    metrics
	lru        *lru.Cache
	trust      bool
}

var (
	errInvalidLocalChunk = errors.New("invalid chunk found locally")
)

// New returns a new NetStore that wraps a given Storer.
func New(s storage.Storer, validStamp voucher.ValidStampFn, r retrieval.Interface, logger log.Logger, memCapacity uint64, trust bool) storage.Storer {
	ns := &store{
		Storer:     s,
		validStamp: validStamp,
		retrieval:  r,
		logger:     logger.WithName(loggerName).Register(),
		bgWorkers:  make(chan struct{}, maxBgPutters),
		metrics:    newMetrics(),
		trust:      trust,
	}
	if memCapacity > 0 {
		lruCache, err := lru.New(int(memCapacity))
		if err == nil {
			ns.lru = lruCache
		}
	}
	ns.sCtx, ns.sCancel = context.WithCancel(context.Background())
	return ns
}

// Get retrieves a given chunk address.
// It will request a chunk from the network whenever it cannot be found locally.
// If the network path is taken, the method also stores the found chunk into the
// local-store.
func (s *store) Get(ctx context.Context, mode storage.ModeGet, addr cluster.Address) (ch cluster.Chunk, err error) {
	ch, err = s.Storer.Get(ctx, mode, addr)
	if err == nil {
		s.metrics.LocalChunksCounter.Inc()
		if !s.trust {
			// ensure the chunk we get locally is valid. If not, retrieve the chunk
			// from network. If there is any corruption of data in the local storage,
			// this would ensure it is retrieved again from network and added back with
			// the correct data
			if !cac.Valid(ch) && !soc.Valid(ch) {
				err = errInvalidLocalChunk
				ch = nil
				s.logger.Warning("netstore: got invalid chunk from localstore, falling back to retrieval")
				s.metrics.InvalidLocalChunksCounter.Inc()
			}
		}
	}
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) || errors.Is(err, errInvalidLocalChunk) {
			found := false
			if s.lru != nil {
				if val, ok := s.lru.Get(addr.String()); ok {
					ch = val.(cluster.Chunk)
					found = true
					s.metrics.RetrievedMemChunksCounter.Inc()
				}
			}
			if !found {
				// request from network
				ch, err = s.retrieval.RetrieveChunk(ctx, addr, cluster.ZeroAddress)
				if err != nil {
					return nil, err
				}
				s.metrics.RetrievedChunksCounter.Inc()
			}

			s.wg.Add(1)
			s.put(ch, mode)
			return ch, nil
		}
		return nil, fmt.Errorf("netstore get: %w", err)
	}
	return ch, nil
}

// put will store the chunk into storage asynchronously
func (s *store) put(ch cluster.Chunk, mode storage.ModeGet) {
	go func() {
		defer s.wg.Done()

		select {
		case <-s.sCtx.Done():
			s.logger.Debug("netstore: stopping netstore")
			return
		case s.bgWorkers <- struct{}{}:
			if s.lru != nil {
				s.lru.Remove(ch.Address().String())
			}
		default:
			if s.lru != nil {
				s.lru.Add(ch.Address().String(), ch)
			}
			return
		}
		defer func() {
			<-s.bgWorkers
		}()

		stamp, err := ch.Stamp().MarshalBinary()
		if err != nil {
			s.logger.Error(err, "failed to marshal stamp from chunk", "chunk_address", ch.Address())
			return
		}

		putMode := storage.ModePutRequest
		if mode == storage.ModeGetRequestPin {
			putMode = storage.ModePutRequestPin
		}

		cch, err := s.validStamp(ch, stamp)
		if err != nil {
			// if a chunk with an invalid voucher stamp was received
			// we force it into the cache.
			putMode = storage.ModePutRequestCache
			cch = ch
		}

		_, err = s.Storer.Put(s.sCtx, putMode, cch)
		if err != nil {
			s.logger.Error(err, "failed to put chunk", "chunk_address", cch.Address())
		}
	}()
}

// The underlying store is not the netstore's responsibility to close
func (s *store) Close() error {
	s.sCancel()

	stopped := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(stopped)
	}()

	select {
	case <-stopped:
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("netstore: waited 5 seconds to close active goroutines")
	}
}
