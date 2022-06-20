// Package p2p provides the peer-to-peer abstractions used
// across different protocols in hop.
package p2p

import (
	"context"
	"github.com/redesblock/hop/core/hop"
	"io"
	"time"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/redesblock/hop/core/swarm"
)

// Service provides methods to handle p2p Peers and Protocols.
type Service interface {
	AddProtocol(ProtocolSpec) error
	// Connect to a peer but do not notify topology about the established connection.
	Connect(ctx context.Context, addr ma.Multiaddr) (address *hop.Address, err error)
	Disconnecter
	Peers() []Peer
	BlocklistedPeers() ([]Peer, error)
	Addresses() ([]ma.Multiaddr, error)
	SetNotifier(Notifier)
}

type Disconnecter interface {
	Disconnect(overlay swarm.Address) error
	// Blocklist will disconnect a peer and put it on a blocklist (blocking in & out connections) for provided duration
	// duration 0 is treated as an infinite duration
	Blocklist(overlay swarm.Address, duration time.Duration) error
}

type Notifier interface {
	Connected(context.Context, Peer) error
	Disconnected(Peer)
}

// DebugService extends the Service with method used for debugging.
type DebugService interface {
	Service
	SetWelcomeMessage(val string) error
	GetWelcomeMessage() string
}

// Streamer is able to create a new Stream.
type Streamer interface {
	NewStream(ctx context.Context, address swarm.Address, h Headers, protocol, version, stream string) (Stream, error)
}

type StreamerDisconnecter interface {
	Streamer
	Disconnecter
}

// Stream represent a bidirectional data Stream.
type Stream interface {
	io.ReadWriter
	io.Closer
	Headers() Headers
	FullClose() error
	Reset() error
}

// ProtocolSpec defines a collection of Stream specifications with handlers.
type ProtocolSpec struct {
	Name          string
	Version       string
	StreamSpecs   []StreamSpec
	ConnectIn     func(context.Context, Peer) error
	ConnectOut    func(context.Context, Peer) error
	DisconnectIn  func(Peer) error
	DisconnectOut func(Peer) error
}

// StreamSpec defines a Stream handling within the protocol.
type StreamSpec struct {
	Name    string
	Handler HandlerFunc
	Headler HeadlerFunc
}

// Peer holds information about a Peer.
type Peer struct {
	Address swarm.Address `json:"address"`
}

// HandlerFunc handles a received Stream from a Peer.
type HandlerFunc func(context.Context, Peer, Stream) error

// HandlerMiddleware decorates a HandlerFunc by returning a new one.
type HandlerMiddleware func(HandlerFunc) HandlerFunc

// HeadlerFunc is returning response headers based on the received request
// headers.
type HeadlerFunc func(Headers) Headers

// Headers represents a collection of p2p header key value pairs.
type Headers map[string][]byte

// Common header names.
const (
	HeaderNameTracingSpanContext = "tracing-span-context"
)

// NewSwarmStreamName constructs a libp2p compatible stream name out of
// protocol name and version and stream name.
func NewSwarmStreamName(protocol, version, stream string) string {
	return "/swarm/" + protocol + "/" + version + "/" + stream
}
