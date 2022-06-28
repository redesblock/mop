// Package pushsync provides the pushsync protocol
// implementation.
package pushsync

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/redesblock/hop/core/accounting"
	"github.com/redesblock/hop/core/cac"
	"github.com/redesblock/hop/core/crypto"
	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/p2p"
	"github.com/redesblock/hop/core/p2p/protobuf"
	"github.com/redesblock/hop/core/postage"
	"github.com/redesblock/hop/core/pricer"
	"github.com/redesblock/hop/core/pushsync/pb"
	"github.com/redesblock/hop/core/soc"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/tags"
	"github.com/redesblock/hop/core/topology"
	"github.com/redesblock/hop/core/tracing"
)

const (
	protocolName    = "pushsync"
	protocolVersion = "1.1.0"
	streamName      = "pushsync"
)

const (
	defaultTTL     = 30 * time.Second // request time to live
	p90TTL         = 5 * time.Second  // P90 request time to live
	sanctionWait   = 5 * time.Minute
	replicationTTL = 5 * time.Second // time to live for neighborhood replication

)

const (
	nPeersToReplicate = 3 // number of peers to replicate to as receipt is sent upstream
	maxAttempts       = 8
	maxPeers          = 16
)

var (
	ErrNoPush            = errors.New("could not push chunk")
	ErrOutOfDepthStoring = errors.New("storing outside of the neighborhood")
	ErrWarmup            = errors.New("node warmup time not complete")
)

type PushSyncer interface {
	PushChunkToClosest(ctx context.Context, ch swarm.Chunk) (*Receipt, error)
}

type Receipt struct {
	Address   swarm.Address
	Signature []byte
	BlockHash []byte
}

type PushSync struct {
	networkID       uint64
	address         swarm.Address
	blockHash       []byte
	streamer        p2p.StreamerDisconnecter
	storer          storage.Putter
	topologyDriver  topology.Driver
	tagger          *tags.Tags
	unwrap          func(swarm.Chunk)
	logger          logging.Logger
	accounting      accounting.Interface
	pricer          pricer.Interface
	metrics         metrics
	tracer          *tracing.Tracer
	validStamp      postage.ValidStampFn
	signer          crypto.Signer
	isFullNode      bool
	warmupPeriod    time.Time
	skipList        *peerSkipList
	receiptEndPoint string
}

type receiptResult struct {
	pushTime  time.Time
	peer      swarm.Address
	receipt   *pb.Receipt
	attempted bool
	err       error
}

func New(networkID uint64, address swarm.Address, blockHash []byte, streamer p2p.StreamerDisconnecter, storer storage.Putter, topology topology.Driver, tagger *tags.Tags, isFullNode bool, unwrap func(swarm.Chunk), validStamp postage.ValidStampFn, logger logging.Logger, accounting accounting.Interface, pricer pricer.Interface, signer crypto.Signer, tracer *tracing.Tracer, warmupTime time.Duration, receiptEndPoint string) *PushSync {
	ps := &PushSync{
		networkID:       networkID,
		address:         address,
		blockHash:       blockHash,
		streamer:        streamer,
		storer:          storer,
		topologyDriver:  topology,
		tagger:          tagger,
		isFullNode:      isFullNode,
		unwrap:          unwrap,
		logger:          logger,
		accounting:      accounting,
		pricer:          pricer,
		metrics:         newMetrics(),
		tracer:          tracer,
		validStamp:      validStamp,
		signer:          signer,
		skipList:        newPeerSkipList(),
		warmupPeriod:    time.Now().Add(warmupTime),
		receiptEndPoint: receiptEndPoint,
	}
	return ps
}

func (s *PushSync) Protocol() p2p.ProtocolSpec {
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

// handler handles chunk delivery from other node and forwards to its destination node.
// If the current node is the destination, it stores in the local store and sends a receipt.
func (ps *PushSync) handler(ctx context.Context, p p2p.Peer, stream p2p.Stream) (err error) {
	now := time.Now()
	w, r := protobuf.NewWriterAndReader(stream)
	ctx, cancel := context.WithTimeout(ctx, defaultTTL)
	defer cancel()
	defer func() {
		if err != nil {
			ps.metrics.TotalHandlerTime.WithLabelValues("failure").Observe(time.Since(now).Seconds())
			ps.metrics.TotalHandlerErrors.Inc()
			_ = stream.Reset()
		} else {
			ps.metrics.TotalHandlerTime.WithLabelValues("success").Observe(time.Since(now).Seconds())
			_ = stream.FullClose()
		}
	}()
	var ch pb.Delivery
	if err = r.ReadMsgWithContext(ctx, &ch); err != nil {
		return fmt.Errorf("pushsync read delivery: %w", err)
	}
	ps.metrics.TotalReceived.Inc()

	chunk := swarm.NewChunk(swarm.NewAddress(ch.Address), ch.Data)
	chunkAddress := chunk.Address()

	span, _, ctx := ps.tracer.StartSpanFromContext(ctx, "pushsync-handler", ps.logger, opentracing.Tag{Key: "address", Value: chunkAddress.String()})
	defer span.Finish()

	stamp := new(postage.Stamp)
	// attaching the stamp is required becase pushToClosest expects a chunk with a stamp
	err = stamp.UnmarshalBinary(ch.Stamp)
	if err != nil {
		return fmt.Errorf("pushsync stamp unmarshall: %w", err)
	}
	chunk.WithStamp(stamp)

	if cac.Valid(chunk) {
		if ps.unwrap != nil {
			go ps.unwrap(chunk)
		}
	} else if !soc.Valid(chunk) {
		return swarm.ErrInvalidChunk
	}

	price := ps.pricer.Price(chunkAddress)

	// if the peer is closer to the chunk, AND it's a full node, we were selected for replication. Return early.
	if p.FullNode {
		if closer, _ := p.Address.Closer(chunkAddress, ps.address); closer {

			ps.metrics.HandlerReplication.Inc()

			ctxd, canceld := context.WithTimeout(context.Background(), replicationTTL)
			defer canceld()

			span, _, ctxd := ps.tracer.StartSpanFromContext(ctxd, "pushsync-replication-storage", ps.logger, opentracing.Tag{Key: "address", Value: chunkAddress.String()})
			defer span.Finish()

			realClosestPeer, err := ps.topologyDriver.ClosestPeer(chunk.Address(), false, topology.Filter{Reachable: true})
			if err == nil {
				if !realClosestPeer.Equal(p.Address) {
					ps.metrics.TotalReplicationFromDistantPeer.Inc()
				} else {
					ps.metrics.TotalReplicationFromClosestPeer.Inc()
				}
			}

			chunk, err = ps.validStamp(chunk, ch.Stamp)
			if err != nil {
				ps.metrics.InvalidStampErrors.Inc()
				ps.metrics.HandlerReplicationErrors.Inc()
				return fmt.Errorf("pushsync replication valid stamp: %w", err)
			}

			_, err = ps.storer.Put(ctxd, storage.ModePutSync, chunk)
			if err != nil {
				ps.metrics.HandlerReplicationErrors.Inc()
				return fmt.Errorf("chunk store: %w", err)
			}

			debit, err := ps.accounting.PrepareDebit(p.Address, price)
			if err != nil {
				ps.metrics.HandlerReplicationErrors.Inc()
				return fmt.Errorf("prepare debit to peer %s before writeback: %w", p.Address.String(), err)
			}
			defer debit.Cleanup()

			b := make([]byte, 8)
			binary.BigEndian.PutUint64(b, ps.networkID)

			// return back receipt
			signature, err := ps.signer.Sign(append(chunkAddress.Bytes(), b...))
			if err != nil {
				ps.metrics.HandlerReplicationErrors.Inc()
				return fmt.Errorf("receipt signature: %w", err)
			}

			receipt := pb.Receipt{Address: chunkAddress.Bytes(), Signature: signature, BlockHash: ps.blockHash}
			if err := w.WriteMsgWithContext(ctxd, &receipt); err != nil {
				ps.metrics.HandlerReplicationErrors.Inc()
				return fmt.Errorf("send receipt to peer %s: %w", p.Address.String(), err)
			}

			err = debit.Apply()
			if err != nil {
				ps.metrics.HandlerReplicationErrors.Inc()
			}
			return err
		}
	}

	// forwarding replication
	storedChunk := false
	defer func() {
		if !storedChunk {
			if ps.warmedUp() && ps.topologyDriver.IsWithinDepth(chunkAddress) {
				verifiedChunk, err := ps.validStamp(chunk, ch.Stamp)
				if err != nil {
					ps.metrics.InvalidStampErrors.Inc()
					ps.logger.Warningf("pushsync: forwarder, invalid stamp for chunk %s", chunkAddress.String())
				} else {
					chunk = verifiedChunk
					_, err = ps.storer.Put(ctx, storage.ModePutSync, chunk)
					if err != nil {
						ps.logger.Warningf("pushsync: within depth peer's attempt to store chunk failed: %v", err)
					}
				}
			}
		}
	}()

	receipt, err := ps.pushToClosest(ctx, chunk, false, p.Address)
	if err != nil {
		if errors.Is(err, topology.ErrWantSelf) {
			ps.metrics.Storer.Inc()
			chunk, err = ps.validStamp(chunk, ch.Stamp)
			if err != nil {
				ps.metrics.InvalidStampErrors.Inc()
				return fmt.Errorf("pushsync storer valid stamp: %w", err)
			}

			_, err = ps.storer.Put(ctx, storage.ModePutSync, chunk)
			if err != nil {
				return fmt.Errorf("chunk store: %w", err)
			}

			storedChunk = true

			b := make([]byte, 8)
			binary.BigEndian.PutUint64(b, ps.networkID)

			signature, err := ps.signer.Sign(append(ch.Address, b...))
			if err != nil {
				return fmt.Errorf("receipt signature: %w", err)
			}

			// return back receipt
			debit, err := ps.accounting.PrepareDebit(p.Address, price)
			if err != nil {
				return fmt.Errorf("prepare debit to peer %s before writeback: %w", p.Address.String(), err)
			}
			defer debit.Cleanup()

			receipt := pb.Receipt{Address: chunkAddress.Bytes(), Signature: signature, BlockHash: ps.blockHash}
			if err := w.WriteMsgWithContext(ctx, &receipt); err != nil {
				return fmt.Errorf("send receipt to peer %s: %w", p.Address.String(), err)
			}

			return debit.Apply()
		}

		ps.metrics.Forwarder.Inc()

		return fmt.Errorf("handler: push to closest: %w", err)
	}

	ps.metrics.Forwarder.Inc()

	debit, err := ps.accounting.PrepareDebit(p.Address, price)
	if err != nil {
		return fmt.Errorf("prepare debit to peer %s before writeback: %w", p.Address.String(), err)
	}
	defer debit.Cleanup()

	// pass back the receipt
	if err := w.WriteMsgWithContext(ctx, receipt); err != nil {
		return fmt.Errorf("send receipt to peer %s: %w", p.Address.String(), err)
	}

	return debit.Apply()
}

// PushChunkToClosest sends chunk to the closest peer by opening a stream. It then waits for
// a receipt from that peer and returns error or nil based on the receiving and
// the validity of the receipt.
func (ps *PushSync) PushChunkToClosest(ctx context.Context, ch swarm.Chunk) (*Receipt, error) {
	ps.metrics.TotalOutgoing.Inc()
	r, err := ps.pushToClosest(ctx, ch, true, swarm.ZeroAddress)
	if err != nil {
		ps.metrics.TotalOutgoingErrors.Inc()
		return nil, err
	}
	return &Receipt{
		Address:   swarm.NewAddress(r.Address),
		Signature: r.Signature,
		BlockHash: r.BlockHash}, nil
}

func (ps *PushSync) pushToClosest(ctx context.Context, ch swarm.Chunk, origin bool, originAddr swarm.Address) (*pb.Receipt, error) {
	span, logger, ctx := ps.tracer.StartSpanFromContext(ctx, "push-closest", ps.logger, opentracing.Tag{Key: "address", Value: ch.Address().String()})
	defer span.Finish()
	defer ps.skipList.PruneExpired()

	var (
		// limits "attempted" requests, see pushPeer when a request becomes attempted
		allowedAttempts = 1
		// limits total requests, irregardless of "attempted"
		allowedRetries = maxPeers
	)

	if origin {
		allowedAttempts = maxAttempts
	}

	var (
		includeSelf = ps.isFullNode
		skipPeers   []swarm.Address
	)

	resultChan := make(chan receiptResult, 1)
	doneChan := make(chan struct{})

	timer := time.NewTimer(0)
	defer timer.Stop()

	nextPeer := func() (swarm.Address, error) {

		fullSkipList := append(ps.skipList.ChunkSkipPeers(ch.Address()), skipPeers...)

		peer, err := ps.topologyDriver.ClosestPeer(ch.Address(), includeSelf, topology.Filter{Reachable: true}, fullSkipList...)
		if err != nil {
			// ClosestPeer can return ErrNotFound in case we are not connected to any peers
			// in which case we should return immediately.
			// if ErrWantSelf is returned, it means we are the closest peer.
			if errors.Is(err, topology.ErrWantSelf) {

				if !ps.warmedUp() {
					return swarm.ZeroAddress, ErrWarmup
				}

				if !ps.topologyDriver.IsWithinDepth(ch.Address()) {
					return swarm.ZeroAddress, ErrOutOfDepthStoring
				}

				ps.pushToNeighbourhood(ctx, fullSkipList, ch, origin, originAddr)
				return swarm.ZeroAddress, err
			}

			return swarm.ZeroAddress, fmt.Errorf("closest peer: %w", err)
		}

		return peer, nil
	}

	for {
		select {
		case <-timer.C:

			allowedRetries--
			// decrement here to limit inflight requests, if the request is not "attempted", we will increment below
			allowedAttempts--

			peer, err := nextPeer()
			if err != nil {
				return nil, err
			}

			ps.metrics.TotalSendAttempts.Inc()

			skipPeers = append(skipPeers, peer)

			ctxd, cancel := context.WithCancel(ctx)
			// cancel only after defaultTTL to allow pushPeer to fully complete for inflight requests
			time.AfterFunc(defaultTTL, cancel)

			go ps.pushPeer(ctxd, resultChan, doneChan, peer, ch, origin)

			// reached the limit, do not set timer to retry
			if allowedRetries <= 0 || allowedAttempts <= 0 {
				continue
			}

			// retry
			timer.Reset(p90TTL)

		case result := <-resultChan:

			ps.measurePushPeer(result.pushTime, result.err)

			if result.err == nil {
				close(doneChan)
				bts, _ := json.Marshal(result.receipt)
				resp, err := http.Post(ps.receiptEndPoint, "application/json", strings.NewReader(string(bts)))
				if err != nil {
					logger.Errorf("pushsync: push receipt error: %s", err)
				} else {
					defer resp.Body.Close()
				}

				logger.Debugf("pushsync: push to peer %s: %s", result.peer, swarm.NewAddress(result.receipt.Address))
				return result.receipt, nil
			}

			ps.metrics.TotalFailedSendAttempts.Inc()
			logger.Debugf("pushsync: could not push to peer %s: %v", result.peer, result.err)

			// pushPeer returned early, do not count as an attempt
			if !result.attempted {
				allowedAttempts++
			}

			if ps.warmedUp() && !errors.Is(result.err, accounting.ErrOverdraft) {
				ps.skipList.Add(ch.Address(), result.peer, sanctionWait)
				ps.metrics.TotalSkippedPeers.Inc()
				logger.Debugf("pushsync: adding to skiplist peer %s", result.peer.String())
			}

			if allowedRetries <= 0 || allowedAttempts <= 0 {
				return nil, ErrNoPush
			}

			// retry immediately
			timer.Reset(0)
		}
	}
}

func (ps *PushSync) measurePushPeer(t time.Time, err error) {
	var status string
	if err != nil {
		status = "failure"
	} else {
		status = "success"
	}
	ps.metrics.PushToPeerTime.WithLabelValues(status).Observe(time.Since(t).Seconds())
}

func (ps *PushSync) pushPeer(ctx context.Context, resultChan chan<- receiptResult, doneChan <-chan struct{}, peer swarm.Address, ch swarm.Chunk, origin bool) {

	var (
		err       error
		receipt   pb.Receipt
		attempted bool
		now       = time.Now()
	)

	defer func() {
		select {
		case resultChan <- receiptResult{pushTime: now, peer: peer, err: err, attempted: attempted, receipt: &receipt}:
		case <-doneChan:
			ps.metrics.DuplicateReceipt.Inc()
		}
	}()

	// compute the price we pay for this receipt and reserve it for the rest of this function
	receiptPrice := ps.pricer.PeerPrice(peer, ch.Address())

	// Reserve to see whether we can make the request
	creditAction, err := ps.accounting.PrepareCredit(peer, receiptPrice, origin)
	if err != nil {
		err = fmt.Errorf("reserve balance for peer %s: %w", peer, err)
		return
	}
	defer creditAction.Cleanup()

	stamp, err := ch.Stamp().MarshalBinary()
	if err != nil {
		return
	}

	streamer, err := ps.streamer.NewStream(ctx, peer, nil, protocolName, protocolVersion, streamName)
	if err != nil {
		err = fmt.Errorf("new stream for peer %s: %w", peer, err)
		return
	}
	defer streamer.Close()

	w, r := protobuf.NewWriterAndReader(streamer)
	err = w.WriteMsgWithContext(ctx, &pb.Delivery{
		Address: ch.Address().Bytes(),
		Data:    ch.Data(),
		Stamp:   stamp,
	})
	if err != nil {
		_ = streamer.Reset()
		err = fmt.Errorf("chunk %s deliver to peer %s: %w", ch.Address(), peer, err)
		return
	}

	ps.metrics.TotalSent.Inc()

	attempted = true

	// if you manage to get a tag, just increment the respective counter
	t, err := ps.tagger.Get(ch.TagID())
	if err == nil && t != nil {
		err = t.Inc(tags.StateSent)
		if err != nil {
			err = fmt.Errorf("tag %d increment: %w", ch.TagID(), err)
			return
		}
	}

	err = r.ReadMsgWithContext(ctx, &receipt)
	if err != nil {
		_ = streamer.Reset()
		err = fmt.Errorf("chunk %s receive receipt from peer %s: %w", ch.Address(), peer, err)
		return
	}

	if !ch.Address().Equal(swarm.NewAddress(receipt.Address)) {
		// if the receipt is invalid, try to push to the next peer
		err = fmt.Errorf("invalid receipt. chunk %s, peer %s", ch.Address(), peer)
		return
	}

	err = creditAction.Apply()
}

func (ps *PushSync) pushToNeighbourhood(ctx context.Context, skiplist []swarm.Address, ch swarm.Chunk, origin bool, originAddr swarm.Address) {
	count := 0
	// Push the chunk to some peers in the neighborhood in parallel for replication.
	// Any errors here should NOT impact the rest of the handler.
	_ = ps.topologyDriver.EachPeer(func(peer swarm.Address, po uint8) (bool, bool, error) {
		// skip forwarding peer
		if peer.Equal(originAddr) {
			return false, false, nil
		}

		// skip skiplisted peers
		for _, s := range skiplist {
			if peer.Equal(s) {
				return false, false, nil
			}
		}

		// here we skip the peer if the peer is closer to the chunk than us
		// we replicate with peers that are further away than us because we are the storer
		if closer, _ := peer.Closer(ch.Address(), ps.address); closer {
			return false, false, nil
		}

		if count == nPeersToReplicate {
			return true, false, nil
		}
		count++
		go ps.pushToNeighbour(ctx, peer, ch, origin)
		return false, false, nil
	}, topology.Filter{Reachable: true})
}

// pushToNeighbour handles in-neighborhood replication for a single peer.
func (ps *PushSync) pushToNeighbour(ctx context.Context, peer swarm.Address, ch swarm.Chunk, origin bool) {
	var err error
	ps.metrics.TotalReplicatedAttempts.Inc()
	defer func() {
		if err != nil {
			ps.logger.Tracef("pushsync replication: %v", err)
			ps.metrics.TotalReplicatedError.Inc()
		}
	}()

	// price for neighborhood replication
	receiptPrice := ps.pricer.PeerPrice(peer, ch.Address())

	// decouple the span data from the original context so it doesn't get
	// cancelled, then glue the stuff on the new context
	span := tracing.FromContext(ctx)

	ctx, cancel := context.WithTimeout(context.Background(), replicationTTL)
	defer cancel()

	// now bring in the span data to the new context
	ctx = tracing.WithContext(ctx, span)
	spanInner, _, ctx := ps.tracer.StartSpanFromContext(ctx, "pushsync-replication", ps.logger, opentracing.Tag{Key: "address", Value: ch.Address().String()})
	defer spanInner.Finish()

	creditAction, err := ps.accounting.PrepareCredit(peer, receiptPrice, origin)
	if err != nil {
		err = fmt.Errorf("reserve balance for peer %s: %w", peer.String(), err)
		return
	}
	defer creditAction.Cleanup()

	streamer, err := ps.streamer.NewStream(ctx, peer, nil, protocolName, protocolVersion, streamName)
	if err != nil {
		err = fmt.Errorf("new stream for peer %s: %w", peer.String(), err)
		return
	}

	defer func() {
		if err != nil {
			_ = streamer.Reset()
		} else {
			_ = streamer.FullClose()
		}
	}()

	w, r := protobuf.NewWriterAndReader(streamer)
	stamp, err := ch.Stamp().MarshalBinary()
	if err != nil {
		return
	}
	err = w.WriteMsgWithContext(ctx, &pb.Delivery{
		Address: ch.Address().Bytes(),
		Data:    ch.Data(),
		Stamp:   stamp,
	})
	if err != nil {
		return
	}

	var receipt pb.Receipt
	if err = r.ReadMsgWithContext(ctx, &receipt); err != nil {
		return
	}

	if !ch.Address().Equal(swarm.NewAddress(receipt.Address)) {
		// if the receipt is invalid, give up
		return
	}

	if err = creditAction.Apply(); err != nil {
		return
	}
}

func (ps *PushSync) warmedUp() bool {
	return time.Now().After(ps.warmupPeriod)
}

type peerSkipList struct {
	sync.Mutex

	// key is chunk address, value is map of peer address to expiration
	skip map[string]map[string]time.Time
}

func newPeerSkipList() *peerSkipList {
	return &peerSkipList{
		skip: make(map[string]map[string]time.Time),
	}
}

func (l *peerSkipList) Add(chunk, peer swarm.Address, expire time.Duration) {
	l.Lock()
	defer l.Unlock()

	if _, ok := l.skip[chunk.ByteString()]; !ok {
		l.skip[chunk.ByteString()] = make(map[string]time.Time)
	}
	l.skip[chunk.ByteString()][peer.ByteString()] = time.Now().Add(expire)
}

func (l *peerSkipList) ChunkSkipPeers(ch swarm.Address) (peers []swarm.Address) {
	l.Lock()
	defer l.Unlock()

	if p, ok := l.skip[ch.ByteString()]; ok {
		for peer, exp := range p {
			if time.Now().Before(exp) {
				peers = append(peers, swarm.NewAddress([]byte(peer)))
			}
		}
	}
	return peers
}

func (l *peerSkipList) PruneExpired() {
	l.Lock()
	defer l.Unlock()

	now := time.Now()

	for k, v := range l.skip {
		kc := len(v)
		for kk, vv := range v {
			if vv.Before(now) {
				delete(v, kk)
				kc--
			}
		}
		if kc == 0 {
			// prune the chunk too
			delete(l.skip, k)
		}
	}
}
