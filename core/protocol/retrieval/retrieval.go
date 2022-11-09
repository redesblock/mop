// Package retrieval provides the retrieval protocol implementation.
// The protocol is used to retrieve chunks over the network using forwarding-kademlia routing.
package retrieval

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/redesblock/mop/core/chunk/cac"
	"github.com/redesblock/mop/core/chunk/soc"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/incentives/bookkeeper"
	"github.com/redesblock/mop/core/incentives/voucher"
	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/p2p"
	"github.com/redesblock/mop/core/p2p/protobuf"
	"github.com/redesblock/mop/core/p2p/topology"
	"github.com/redesblock/mop/core/pricer"
	pb "github.com/redesblock/mop/core/protocol/retrieval/pb"
	"github.com/redesblock/mop/core/skiper"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/redesblock/mop/core/tracer"
	"resenje.org/singleflight"
)

// loggerName is the tree path name of the logger for this package.
const loggerName = "retrieval"

const (
	protocolName    = "retrieval"
	protocolVersion = "1.2.0"
	streamName      = "retrieval"
)

var _ Interface = (*Service)(nil)

type Interface interface {
	// RetrieveChunk retrieves a chunk from the network using the retrieval protocol.
	// it takes as parameters a context, a chunk address to retrieve (content-addressed or single-owner) and
	// a source peer address, for the case that we are requesting the chunk for another peer. In case the request
	// originates at the current node (i.e. no forwarding involved), the caller should use cluster.ZeroAddress
	// as the value for sourcePeerAddress.
	RetrieveChunk(ctx context.Context, address, sourcePeerAddr cluster.Address) (chunk cluster.Chunk, err error)
}

type retrievalResult struct {
	chunk     cluster.Chunk
	peer      cluster.Address
	err       error
	retrieved bool
}

type Service struct {
	addr          cluster.Address
	streamer      p2p.Streamer
	peerSuggester topology.ClosestPeerer
	storer        storage.Storer
	singleflight  singleflight.Group
	logger        log.Logger
	accounting    bookkeeper.Interface
	metrics       metrics
	pricer        pricer.Interface
	tracer        *tracer.Tracer
	caching       bool
	validStamp    voucher.ValidStampFn
}

func New(addr cluster.Address, storer storage.Storer, streamer p2p.Streamer, chunkPeerer topology.ClosestPeerer, logger log.Logger, accounting bookkeeper.Interface, pricer pricer.Interface, tracer *tracer.Tracer, forwarderCaching bool, validStamp voucher.ValidStampFn) *Service {
	return &Service{
		addr:          addr,
		streamer:      streamer,
		peerSuggester: chunkPeerer,
		storer:        storer,
		logger:        logger.WithName(loggerName).Register(),
		accounting:    accounting,
		pricer:        pricer,
		metrics:       newMetrics(),
		tracer:        tracer,
		caching:       forwarderCaching,
		validStamp:    validStamp,
	}
}

func (s *Service) Protocol() p2p.ProtocolSpec {
	return p2p.ProtocolSpec{
		Name:    protocolName,
		Version: protocolVersion,
		StreamSpecs: []p2p.StreamSpec{
			{
				Name:    streamName,
				Handler: s.handler,
			},
		},
	}
}

const (
	retrieveChunkTimeout          = 10 * time.Second
	retrieveRetryIntervalDuration = time.Second
	maxRequestRounds              = 1024
	maxSelects                    = 32
	originSuffix                  = "_origin"
)

func (s *Service) RetrieveChunk(ctx context.Context, addr, sourcePeerAddr cluster.Address) (cluster.Chunk, error) {
	loggerV1 := s.logger.V(1).Register()

	s.metrics.RequestCounter.Inc()

	origin := sourcePeerAddr.IsZero()

	flightRoute := addr.String()
	if origin {
		flightRoute = addr.String() + originSuffix
	}

	// topCtx is passing the tracer span to the first singleflight call
	topCtx := ctx
	v, _, err := s.singleflight.Do(ctx, flightRoute, func(ctx context.Context) (interface{}, error) {
		maxPeers := 1
		if origin {
			maxPeers = maxSelects
		}

		sp := new(skiper.List)

		if !sourcePeerAddr.IsZero() {
			sp.Add(sourcePeerAddr)
		}

		ticker := time.NewTicker(retrieveRetryIntervalDuration)
		defer ticker.Stop()

		var (
			peerAttempt  int
			peersResults int
			resultC      = make(chan retrievalResult)
		)

		requestAttempt := 0

		lastTime := time.Now().Unix()

		for requestAttempt < maxRequestRounds {

			if peerAttempt < maxSelects {

				// create a new context without cancelation but
				// set the tracer span to the new context from the context of the first caller
				ctx := tracer.WithContext(context.Background(), tracer.FromContext(topCtx))

				// get the tracer span
				span, _, ctx := s.tracer.StartSpanFromContext(ctx, "retrieve-chunk", s.logger, opentracing.Tag{Key: "address", Value: addr.String()})
				defer span.Finish()

				peerAttempt++
				s.metrics.PeerRequestCounter.Inc()
				go func() {

					// cancel the goroutine just with the timeout
					ctx, cancel := context.WithTimeout(ctx, retrieveChunkTimeout)
					defer cancel()
					chunk, peer, requested, err := s.retrieveChunk(ctx, addr, sp, origin)
					select {
					case resultC <- retrievalResult{
						chunk:     chunk,
						peer:      peer,
						err:       err,
						retrieved: requested,
					}:
					case <-ctx.Done():
					}

				}()
			} else {
				select {
				case resultC <- retrievalResult{}:
				case <-ctx.Done():
				default:
				}
			}

			select {
			case <-ticker.C:
				// break
			case res := <-resultC:
				if errors.Is(res.err, topology.ErrNotFound) {
					if sp.OverdraftListEmpty() {
						// if no peer is available, and none skipped temporarily
						loggerV1.Debug("failed to get chunk", "chunk_address", addr)
						return nil, storage.ErrNotFound
					} else {
						// skip to next request round if any peers are only skipped temporarily
						peerAttempt = maxSelects
					}
				}
				if res.retrieved {
					if res.err != nil {
						if !res.peer.IsZero() {
							s.logger.Debug("failed to get chunk from peer", "chunk_address", addr, "peer_address", res.peer, "error", res.err)
						}
						peersResults++
					} else {
						return res.chunk, nil
					}
				}
			case <-ctx.Done():
				loggerV1.Debug("failed to get chunk", "chunk_address", addr, "error", ctx.Err())
				return nil, fmt.Errorf("retrieval: %w", ctx.Err())
			}

			// all results received, only successfully attempted requests are counted
			if peersResults >= maxPeers {
				loggerV1.Debug("failed to get chunk", "chunk_address", addr)
				return nil, storage.ErrNotFound
			}

			// if we have not counted enough successful attempts but out of selection amount, reset
			if peerAttempt >= maxSelects {

				// forwarder gives up after first round of request attempts,
				// rather than holding up the forwarding chain by waiting to be able to reserve downstream peer
				if !origin {
					return nil, storage.ErrNotFound
				}

				requestAttempt++
				timeNow := time.Now().Unix()
				if timeNow > lastTime {
					lastTime = timeNow
					peerAttempt = 0
					sp.ResetOverdraft()
				} else {
					select {
					case <-time.After(600 * time.Millisecond):
					case <-ctx.Done():
						loggerV1.Debug("failed to get chunk", "chunk_address", addr, "error", ctx.Err())
						return nil, fmt.Errorf("retrieval: %w", ctx.Err())
					}
				}
			}

		}

		// if we have not managed to get results after 5 (maxRequestRounds) rounds of peer selections, give up
		return nil, storage.ErrNotFound

	})
	if err != nil {
		return nil, err
	}

	return v.(cluster.Chunk), nil
}

func (s *Service) retrieveChunk(ctx context.Context, addr cluster.Address, sp *skiper.List, isOrigin bool) (chunk cluster.Chunk, peer cluster.Address, requested bool, err error) {
	loggerV1 := s.logger.V(1).Build()

	startTimer := time.Now()
	// allow upstream requests if this node is the source of the request
	// i.e. the request was not forwarded, to improve retrieval
	// if this node is the closest to he chunk but still does not contain it
	allowUpstream := isOrigin

	ctx, cancel := context.WithTimeout(ctx, retrieveChunkTimeout)
	defer cancel()
	peer, err = s.closestPeer(addr, sp.All(), allowUpstream)
	if err != nil {
		return nil, peer, false, fmt.Errorf("get closest for address %s, allow upstream %v: %w", addr.String(), allowUpstream, err)
	}

	// compute the peer's price for this chunk for price header
	chunkPrice := s.pricer.PeerPrice(peer, addr)

	creditCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Reserve to see whether we can request the chunk
	creditAction, err := s.accounting.PrepareCredit(creditCtx, peer, chunkPrice, isOrigin)
	if err != nil {
		sp.AddOverdraft(peer)
		return nil, peer, false, err
	}
	defer creditAction.Cleanup()

	sp.Add(peer)

	loggerV1.Debug("requesting chunk from peer", "chunk_address", addr, "peer_address", peer)

	stream, err := s.streamer.NewStream(ctx, peer, nil, protocolName, protocolVersion, streamName)
	if err != nil {
		s.metrics.TotalErrors.Inc()
		return nil, peer, false, fmt.Errorf("new stream: %w", err)
	}

	defer func() {
		if err != nil {
			_ = stream.Reset()
		} else {
			_ = stream.FullClose()
		}
	}()

	w, r := protobuf.NewWriterAndReader(stream)
	if err := w.WriteMsgWithContext(ctx, &pb.Request{
		Addr: addr.Bytes(),
	}); err != nil {
		s.metrics.TotalErrors.Inc()
		return nil, peer, false, fmt.Errorf("write request: %w peer %s", err, peer.String())
	}
	var d pb.Delivery
	if err := r.ReadMsgWithContext(ctx, &d); err != nil {
		s.metrics.TotalErrors.Inc()
		return nil, peer, true, fmt.Errorf("read delivery: %w peer %s", err, peer.String())
	}
	s.metrics.ChunkRetrieveTime.Observe(time.Since(startTimer).Seconds())
	s.metrics.TotalRetrieved.Inc()

	stamp := new(voucher.Stamp)
	err = stamp.UnmarshalBinary(d.Stamp)
	if err != nil {
		return nil, peer, true, fmt.Errorf("stamp unmarshal: %w", err)
	}
	chunk = cluster.NewChunk(addr, d.Data).WithStamp(stamp)
	if !cac.Valid(chunk) {
		if !soc.Valid(chunk) {
			s.metrics.InvalidChunkRetrieved.Inc()
			s.metrics.TotalErrors.Inc()
			return nil, peer, true, cluster.ErrInvalidChunk
		}
	}

	// credit the peer after successful delivery
	err = creditAction.Apply()
	if err != nil {
		return nil, peer, true, err
	}
	s.metrics.ChunkPrice.Observe(float64(chunkPrice))
	return chunk, peer, true, err
}

// closestPeer returns address of the peer that is closest to the chunk with
// provided address addr. This function will ignore peers with addresses
// provided in skipPeers and if allowUpstream is true, peers that are further of
// the chunk than this node is, could also be returned, allowing the upstream
// retrieve request.
func (s *Service) closestPeer(addr cluster.Address, skipPeers []cluster.Address, allowUpstream bool) (cluster.Address, error) {

	closest, err := s.peerSuggester.ClosestPeer(addr, false, topology.Filter{Reachable: true}, skipPeers...)
	if err != nil {
		return cluster.Address{}, err
	}

	if allowUpstream {
		return closest, nil
	}

	closer, err := closest.Closer(addr, s.addr)
	if err != nil {
		return cluster.Address{}, fmt.Errorf("distance compare addr %s closest %s base address %s: %w", addr.String(), closest.String(), s.addr.String(), err)
	}
	if !closer {
		return cluster.Address{}, topology.ErrNotFound
	}

	return closest, nil
}

func (s *Service) handler(ctx context.Context, p p2p.Peer, stream p2p.Stream) (err error) {
	loggerV1 := s.logger.V(1).Register()

	ctx, cancel := context.WithTimeout(ctx, retrieveChunkTimeout)
	defer cancel()

	w, r := protobuf.NewWriterAndReader(stream)
	defer func() {
		if err != nil {
			_ = stream.Reset()
		} else {
			_ = stream.FullClose()
		}
	}()
	var req pb.Request
	if err := r.ReadMsgWithContext(ctx, &req); err != nil {
		return fmt.Errorf("read request: %w peer %s", err, p.Address.String())
	}

	span, _, ctx := s.tracer.StartSpanFromContext(ctx, "handle-retrieve-chunk", s.logger, opentracing.Tag{Key: "address", Value: cluster.NewAddress(req.Addr).String()})
	defer span.Finish()

	addr := cluster.NewAddress(req.Addr)

	forwarded := false
	chunk, err := s.storer.Get(ctx, storage.ModeGetRequest, addr)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			// forward the request
			chunk, err = s.RetrieveChunk(ctx, addr, p.Address)
			if err != nil {
				return fmt.Errorf("retrieve chunk: %w", err)
			}
			forwarded = true
		} else {
			return fmt.Errorf("get from store: %w", err)
		}
	}
	stamp, err := chunk.Stamp().MarshalBinary()
	if err != nil {
		return fmt.Errorf("stamp marshal: %w", err)
	}

	chunkPrice := s.pricer.Price(chunk.Address())
	debit, err := s.accounting.PrepareDebit(ctx, p.Address, chunkPrice)
	if err != nil {
		return fmt.Errorf("prepare debit to peer %s before writeback: %w", p.Address.String(), err)
	}
	defer debit.Cleanup()

	if err := w.WriteMsgWithContext(ctx, &pb.Delivery{
		Data:  chunk.Data(),
		Stamp: stamp,
	}); err != nil {
		return fmt.Errorf("write delivery: %w peer %s", err, p.Address.String())
	}

	loggerV1.Debug("retrieval protocol debiting peer", "peer_address", p.Address)

	// debit price from p's balance
	if err := debit.Apply(); err != nil {
		return fmt.Errorf("apply debit: %w", err)
	}

	// cache the request last, so that putting to the localstore does not slow down the request flow
	if s.caching && forwarded {
		putMode := storage.ModePutRequest

		cch, err := s.validStamp(chunk, stamp)
		if err != nil {
			// if a chunk with an invalid voucher stamp was received
			// we force it into the cache.
			putMode = storage.ModePutRequestCache
			cch = chunk
		}

		_, err = s.storer.Put(ctx, putMode, cch)
		if err != nil {
			return fmt.Errorf("retrieve cache put: %w", err)
		}
	}
	return nil
}
