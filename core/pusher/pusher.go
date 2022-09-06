// Package pusher provides protocol-orchestrating functionality
// over the pushsync protocol. It makes sure that chunks meant
// to be distributed over the network are sent used using the
// pushsync protocol.
package pusher

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/redesblock/hop/core/crypto"
	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/postage"
	"github.com/redesblock/hop/core/pushsync"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/tags"
	"github.com/redesblock/hop/core/topology"
	"github.com/redesblock/hop/core/tracing"

	"github.com/sirupsen/logrus"
)

type Service struct {
	networkID         uint64
	storer            storage.Storer
	pushSyncer        pushsync.PushSyncer
	validStamp        postage.ValidStampFn
	depther           topology.NeighborhoodDepther
	logger            logging.Logger
	tag               *tags.Tags
	metrics           metrics
	quit              chan struct{}
	chunksWorkerQuitC chan struct{}
	inflight          *inflight
	attempts          *attempts
	sem               chan struct{}
}

var (
	retryInterval    = 5 * time.Second  // time interval between retries
	traceDuration    = 30 * time.Second // duration for every root tracing span
	concurrentPushes = 10               // how many chunks to push simultaneously
	retryCount       = 6
)

var (
	ErrInvalidAddress = errors.New("invalid address")
	ErrShallowReceipt = errors.New("shallow recipt")
)

func New(networkID uint64, storer storage.Storer, depther topology.NeighborhoodDepther, pushSyncer pushsync.PushSyncer, validStamp postage.ValidStampFn, tagger *tags.Tags, logger logging.Logger, tracer *tracing.Tracer, warmupTime time.Duration) *Service {
	p := &Service{
		networkID:         networkID,
		storer:            storer,
		pushSyncer:        pushSyncer,
		validStamp:        validStamp,
		depther:           depther,
		tag:               tagger,
		logger:            logger,
		metrics:           newMetrics(),
		quit:              make(chan struct{}),
		chunksWorkerQuitC: make(chan struct{}),
		inflight:          newInflight(),
		attempts:          &attempts{attempts: make(map[string]int)},
		sem:               make(chan struct{}, concurrentPushes),
	}
	go p.chunksWorker(warmupTime, tracer)
	return p
}

// chunksWorker is a loop that keeps looking for chunks that are locally uploaded ( by monitoring pushIndex )
// and pushes them to the closest peer and get a receipt.
func (s *Service) chunksWorker(warmupTime time.Duration, tracer *tracing.Tracer) {
	defer close(s.chunksWorkerQuitC)
	select {
	case <-time.After(warmupTime):
		s.logger.Info("pusher: warmup period complete, worker starting.")
	case <-s.quit:
		return
	}

	var (
		cctx, cancel      = context.WithCancel(context.Background())
		mtx               sync.Mutex
		wg                sync.WaitGroup
		span, logger, ctx = tracer.StartSpanFromContext(cctx, "pusher-sync-batch", s.logger)
		timer             = time.NewTimer(traceDuration)
	)

	// inflight.set handles the backpressure for the maximum amount of inflight chunks
	// and duplicate handling.
	chunks, repeat, unsubscribe := s.storer.SubscribePush(ctx, s.inflight.set)
	go func() {
		<-s.quit
		unsubscribe()
		cancel()
		if !timer.Stop() {
			<-timer.C
		}
	}()

	ctxLogger := func() (context.Context, *logrus.Entry) {
		mtx.Lock()
		defer mtx.Unlock()
		return ctx, logger
	}

	go func() {
		for {
			select {
			case <-s.quit:
				return
			case <-timer.C:
				// reset the span
				mtx.Lock()
				span.Finish()
				span, logger, ctx = tracer.StartSpanFromContext(cctx, "pusher-sync-batch", s.logger)
				mtx.Unlock()
			}
		}
	}()

	for ch := range chunks {
		// If the stamp is invalid, the chunk is not synced with the network
		// since other nodes would reject the chunk, so the chunk is marked as
		// synced which makes it available to the node but not to the network
		if err := s.valid(ch); err != nil {
			logger.Warningf("pusher: stamp with batch ID %x is no longer valid, skipping syncing for chunk %s: %v", ch.Stamp().BatchID(), ch.Address().String(), err)
			if err = s.storer.Set(ctx, storage.ModeSetSync, ch.Address()); err != nil {
				s.logger.Errorf("pusher: set sync: %w", err)
			}
			continue
		}
		select {
		case s.sem <- struct{}{}:
		case <-s.quit:
			return
		}
		s.metrics.TotalToPush.Inc()
		ctx, logger := ctxLogger()
		startTime := time.Now()
		wg.Add(1)
		go func(ctx context.Context, ch swarm.Chunk) {
			defer func() {
				wg.Done()
				<-s.sem
			}()
			if err := s.pushChunk(ctx, ch, logger); err != nil {
				repeat()
				s.metrics.TotalErrors.Inc()
				s.metrics.ErrorTime.Observe(time.Since(startTime).Seconds())
				logger.Tracef("pusher: cannot push chunk %s: %v", ch.Address().String(), err)
				return
			}
			s.metrics.TotalSynced.Inc()
		}(ctx, ch)
	}

	wg.Wait()
}

func (s *Service) pushChunk(ctx context.Context, ch swarm.Chunk, logger *logrus.Entry) error {
	defer s.inflight.delete(ch)
	var wantSelf bool
	// Later when we process receipt, get the receipt and process it
	// for now ignoring the receipt and checking only for error
	t := time.Now()
	receipt, err := s.pushSyncer.PushChunkToClosest(ctx, ch)
	if err != nil {
		if !errors.Is(err, topology.ErrWantSelf) {
			return err
		}
		// we are the closest ones - this is fine
		// this is to make sure that the sent number does not diverge from the synced counter
		// the edge case is on the uploader node, in the case where the uploader node is
		// connected to other nodes, but is the closest one to the chunk.
		wantSelf = true
		logger.Tracef("pusher: chunk %s stays here, i'm the closest node", ch.Address().String())
	} else if err = s.checkReceipt(receipt); err != nil {
		return err
	}
	logger.Debugf("xxxxx tagID %v, chunk %v, peer %v, size %v, duration %v", ch.TagID(), ch.Address().String(), receipt.Address.String(), len(ch.Data())/1024, time.Now().Sub(t))
	if err = s.storer.Set(ctx, storage.ModeSetSync, ch.Address()); err != nil {
		return fmt.Errorf("pusher: set sync: %w", err)
	}
	if ch.TagID() > 0 {
		// for individual chunks uploaded using the
		// /chunks api endpoint the tag will be missing
		// by default, unless the api consumer specifies one
		t, err := s.tag.Get(ch.TagID())
		if err == nil && t != nil {
			err = t.Inc(tags.StateSynced)
			if err != nil {
				logger.Debugf("pusher: increment synced: %v", err)
				return nil // tag error is non-fatal
			}
			if wantSelf {
				err = t.Inc(tags.StateSent)
				if err != nil {
					logger.Debugf("pusher: increment sent: %v", err)
					return nil // tag error is non-fatal
				}
			}
		}
	}
	return nil
}

func (s *Service) checkReceipt(receipt *pushsync.Receipt) error {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, s.networkID)

	addr := receipt.Address
	publicKey, err := crypto.Recover(receipt.Signature, append(addr.Bytes(), b...))
	if err != nil {
		return fmt.Errorf("pusher: receipt recover: %w", err)
	}

	peer, err := crypto.NewOverlayAddress(*publicKey, s.networkID, receipt.BlockHash)
	if err != nil {
		return fmt.Errorf("pusher: receipt storer address: %w", err)
	}

	po := swarm.Proximity(addr.Bytes(), peer.Bytes())
	d := s.depther.NeighborhoodDepth()
	if po < d && s.attempts.try(addr) {
		s.metrics.ShallowReceiptDepth.WithLabelValues(strconv.Itoa(int(po))).Inc()
		return fmt.Errorf("pusher: shallow receipt depth %d, want at least %d", po, d)
	}
	s.logger.Tracef("pusher: pushed chunk %s to node %s, receipt depth %d", addr, peer, po)
	s.metrics.ReceiptDepth.WithLabelValues(strconv.Itoa(int(po))).Inc()
	return nil
}

// valid checks whether the stamp for a chunk is valid before sending
// it out on the network.
func (s *Service) valid(ch swarm.Chunk) error {
	stampBytes, err := ch.Stamp().MarshalBinary()
	if err != nil {
		return fmt.Errorf("pusher: valid stamp marshal: %w", err)
	}
	_, err = s.validStamp(ch, stampBytes)
	if err != nil {
		return fmt.Errorf("pusher: valid stamp: %w", err)
	}
	return nil
}

func (s *Service) Close() error {
	s.logger.Info("pusher shutting down")
	close(s.quit)

	// Wait for chunks worker to finish
	select {
	case <-s.chunksWorkerQuitC:
	case <-time.After(6 * time.Second):
	}
	return nil
}
