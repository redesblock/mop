// Package netstore provides an abstraction layer over the
// Flock local storage layer that leverages connectivity
// with other peers in order to retrieve chunks from the network that cannot
// be found locally.
package netstore

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redesblock/mop/core/flock"
	"github.com/redesblock/mop/core/logging"
	"github.com/redesblock/mop/core/postage"
	"github.com/redesblock/mop/core/recovery"
	"github.com/redesblock/mop/core/retrieval"
	"github.com/redesblock/mop/core/sctx"
	"github.com/redesblock/mop/core/storage"
)

const (
	maxBgPutters int = 16
)

type store struct {
	storage.Storer
	retrieval        retrieval.Interface
	logger           logging.Logger
	validVouch       postage.ValidVouchFn
	recoveryCallback recovery.Callback // this is the callback to be executed when a chunk fails to be retrieved
	bgWorkers        chan struct{}
	sCtx             context.Context
	sCancel          context.CancelFunc
	wg               sync.WaitGroup
}

var (
	ErrRecoveryAttempt = errors.New("failed to retrieve chunk, recovery initiated")
)

// New returns a new NetStore that wraps a given Storer.
func New(s storage.Storer, validVouch postage.ValidVouchFn, rcb recovery.Callback, r retrieval.Interface, logger logging.Logger) storage.Storer {
	ns := &store{Storer: s, validVouch: validVouch, recoveryCallback: rcb, retrieval: r, logger: logger}
	ns.sCtx, ns.sCancel = context.WithCancel(context.Background())
	ns.bgWorkers = make(chan struct{}, maxBgPutters)
	return ns
}

// Get retrieves a given chunk address.
// It will request a chunk from the network whenever it cannot be found locally.
// If the network path is taken, the method also stores the found chunk into the
// local-store.
func (s *store) Get(ctx context.Context, mode storage.ModeGet, addr flock.Address) (ch flock.Chunk, err error) {
	ch, err = s.Storer.Get(ctx, mode, addr)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			// request from network
			ch, err = s.retrieval.RetrieveChunk(ctx, addr, true)
			if err != nil {
				targets := sctx.GetTargets(ctx)
				if targets == nil || s.recoveryCallback == nil {
					return nil, err
				}
				go s.recoveryCallback(addr, targets)
				return nil, ErrRecoveryAttempt
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
func (s *store) put(ch flock.Chunk, mode storage.ModeGet) {
	go func() {
		defer s.wg.Done()

		select {
		case <-s.sCtx.Done():
			s.logger.Debug("netstore: stopping netstore")
			return
		case s.bgWorkers <- struct{}{}:
		}
		defer func() {
			<-s.bgWorkers
		}()

		vouch, err := ch.Vouch().MarshalBinary()
		if err != nil {
			s.logger.Errorf("netstore: failed to marshal vouch from chunk %s err:%s", ch.Address(), err.Error())
			return
		}

		putMode := storage.ModePutRequest
		if mode == storage.ModeGetRequestPin {
			putMode = storage.ModePutRequestPin
		}

		cch, err := s.validVouch(ch, vouch)
		if err != nil {
			// if a chunk with an invalid postage vouch was received
			// we force it into the cache.
			putMode = storage.ModePutRequestCache
			cch = ch
		}

		_, err = s.Storer.Put(s.sCtx, putMode, cch)
		if err != nil {
			s.logger.Errorf("netstore: failed to put chunk %s err: %s", cch.Address(), err.Error())
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
