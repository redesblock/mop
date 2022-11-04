// Package psser exposes functionalities needed to communicate
// with other peers on the network. Pss uses pushsync and
// pullsync for message delivery and mailboxing. All messages are disguised as content-addressed chunks. Sending and
// receiving of messages is exposed over the HTTP API, with
// websocket subscriptions for incoming messages.
package psser

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/redesblock/mop/core/chunk/trojan"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/incentives/voucher"
	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/protocol/pushsync"
)

// loggerName is the tree path name of the logger for this package.
const loggerName = "psser"

var (
	_            Interface = (*pss)(nil)
	ErrNoHandler           = errors.New("no handler found")
)

type Sender interface {
	// Send arbitrary byte slice with the given topic to Targets.
	Send(context.Context, trojan.Topic, []byte, voucher.Stamper, *ecdsa.PublicKey, trojan.Targets) error
}

type Interface interface {
	Sender
	// Register a Handler for a given Topic.
	Register(trojan.Topic, Handler) func()
	// TryUnwrap tries to unwrap a wrapped trojan message.
	TryUnwrap(cluster.Chunk)

	SetPushSyncer(pushSyncer pushsync.PushSyncer)
	io.Closer
}

type pss struct {
	key        *ecdsa.PrivateKey
	pusher     pushsync.PushSyncer
	handlers   map[trojan.Topic][]*Handler
	handlersMu sync.Mutex
	metrics    metrics
	logger     log.Logger
	quit       chan struct{}
}

// New returns a new psser service.
func New(key *ecdsa.PrivateKey, logger log.Logger) Interface {
	return &pss{
		key:      key,
		logger:   logger.WithName(loggerName).Register(),
		handlers: make(map[trojan.Topic][]*Handler),
		metrics:  newMetrics(),
		quit:     make(chan struct{}),
	}
}

func (ps *pss) Close() error {
	close(ps.quit)
	ps.handlersMu.Lock()
	defer ps.handlersMu.Unlock()

	ps.handlers = make(map[trojan.Topic][]*Handler) //unset handlers on shutdown

	return nil
}

func (ps *pss) SetPushSyncer(pushSyncer pushsync.PushSyncer) {
	ps.pusher = pushSyncer
}

// Handler defines code to be executed upon reception of a trojan message.
type Handler func(context.Context, []byte)

// Send constructs a padded message with topic and payload,
// wraps it in a trojan chunk such that one of the targets is a prefix of the chunk address.
// Uses push-chainsync to deliver message.
func (p *pss) Send(ctx context.Context, topic trojan.Topic, payload []byte, stamper voucher.Stamper, recipient *ecdsa.PublicKey, targets trojan.Targets) error {
	p.metrics.TotalMessagesSentCounter.Inc()

	tStart := time.Now()

	tc, err := trojan.Wrap(ctx, topic, payload, recipient, targets)
	if err != nil {
		return err
	}

	stamp, err := stamper.Stamp(tc.Address())
	if err != nil {
		return err
	}
	tc = tc.WithStamp(stamp)

	p.metrics.MessageMiningDuration.Set(time.Since(tStart).Seconds())

	// push the chunk using push chainsync so that it reaches it destination in network
	if _, err = p.pusher.PushChunkToClosest(ctx, tc); err != nil {
		return err
	}

	return nil
}

// Register allows the definition of a Handler func for a specific topic on the psser struct.
func (p *pss) Register(topic trojan.Topic, handler Handler) (cleanup func()) {
	p.handlersMu.Lock()
	defer p.handlersMu.Unlock()

	p.handlers[topic] = append(p.handlers[topic], &handler)

	return func() {
		p.handlersMu.Lock()
		defer p.handlersMu.Unlock()

		h := p.handlers[topic]
		for i := 0; i < len(h); i++ {
			if h[i] == &handler {
				p.handlers[topic] = append(h[:i], h[i+1:]...)
				return
			}
		}
	}
}

func (p *pss) topics() []trojan.Topic {
	p.handlersMu.Lock()
	defer p.handlersMu.Unlock()

	ts := make([]trojan.Topic, 0, len(p.handlers))
	for t := range p.handlers {
		ts = append(ts, t)
	}

	return ts
}

// TryUnwrap allows unwrapping a chunk as a trojan message and calling its handlers based on the topic.
func (p *pss) TryUnwrap(c cluster.Chunk) {
	if len(c.Data()) < cluster.ChunkWithSpanSize {
		return // chunk not full
	}
	ctx := context.Background()
	topic, msg, err := trojan.Unwrap(ctx, p.key, c, p.topics())
	if err != nil {
		return // cannot unwrap
	}
	h := p.getHandlers(topic)
	if h == nil {
		return // no handler
	}

	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	var wg sync.WaitGroup
	go func() {
		defer cancel()
		select {
		case <-p.quit:
		case <-done:
		}
	}()
	for _, hh := range h {
		wg.Add(1)
		go func(hh Handler) {
			defer wg.Done()
			hh(ctx, msg)
		}(*hh)
	}
	go func() {
		wg.Wait()
		close(done)
	}()
}

func (p *pss) getHandlers(topic trojan.Topic) []*Handler {
	p.handlersMu.Lock()
	defer p.handlersMu.Unlock()

	return p.handlers[topic]
}
