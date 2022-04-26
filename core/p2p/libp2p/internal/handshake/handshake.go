package handshake

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redesblock/hop/core/crypto"
	"github.com/redesblock/hop/core/hop"
	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/p2p"
	"github.com/redesblock/hop/core/p2p/libp2p/internal/handshake/pb"
	"github.com/redesblock/hop/core/p2p/protobuf"
	"github.com/redesblock/hop/core/swarm"

	"github.com/libp2p/go-libp2p-core/network"
	libp2ppeer "github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	ProtocolName    = "handshake"
	ProtocolVersion = "1.0.0"
	StreamName      = "handshake"
	messageTimeout  = 5 * time.Second // maximum allowed time for a message to be read or written.
)

var (
	// ErrNetworkIDIncompatible is returned if response from the other peer does not have valid networkID.
	ErrNetworkIDIncompatible = errors.New("incompatible network ID")

	// ErrHandshakeDuplicate is returned  if the handshake response has been received by an already processed peer.
	ErrHandshakeDuplicate = errors.New("duplicate handshake")

	// ErrInvalidHopAddress is returned if peer info was received with invalid hop address
	ErrInvalidHopAddress = errors.New("invalid hop address")

	// ErrInvalidAck is returned if ack does not match the syn provided
	ErrInvalidAck = errors.New("invalid ack")
)

// PeerFinder has the information if the peer already exists in swarm.
type PeerFinder interface {
	Exists(overlay swarm.Address) (found bool)
}

type Service struct {
	HopAddress           hop.Address
	networkID            uint64
	receivedHandshakes   map[libp2ppeer.ID]struct{}
	receivedHandshakesMu sync.Mutex
	logger               logging.Logger

	network.Notifiee // handshake service can be the receiver for network.Notify
}

func New(overlay swarm.Address, underlay ma.Multiaddr, signer crypto.Signer, networkID uint64, logger logging.Logger) (*Service, error) {
	HopAddress, err := hop.NewAddress(signer, underlay, overlay, networkID)
	if err != nil {
		return nil, err
	}

	return &Service{
		HopAddress:         *HopAddress,
		networkID:          networkID,
		receivedHandshakes: make(map[libp2ppeer.ID]struct{}),
		logger:             logger,
		Notifiee:           new(network.NoopNotifiee),
	}, nil
}

func (s *Service) Handshake(stream p2p.Stream) (i *Info, err error) {
	w, r := protobuf.NewWriterAndReader(stream)
	if err := w.WriteMsgWithTimeout(messageTimeout, &pb.Syn{
		HopAddress: &pb.HopAddress{
			Underlay:  s.HopAddress.Underlay.Bytes(),
			Signature: s.HopAddress.Signature,
			Overlay:   s.HopAddress.Overlay.Bytes(),
		},
		NetworkID: s.networkID,
	}); err != nil {
		return nil, fmt.Errorf("write syn message: %w", err)
	}

	var resp pb.SynAck
	if err := r.ReadMsgWithTimeout(messageTimeout, &resp); err != nil {
		return nil, fmt.Errorf("read synack message: %w", err)
	}

	if err := s.checkAck(resp.Ack); err != nil {
		return nil, err
	}

	if resp.Syn.NetworkID != s.networkID {
		return nil, ErrNetworkIDIncompatible
	}

	HopAddress, err := hop.ParseAddress(resp.Syn.HopAddress.Underlay, resp.Syn.HopAddress.Overlay, resp.Syn.HopAddress.Signature, resp.Syn.NetworkID)
	if err != nil {
		return nil, ErrInvalidHopAddress
	}

	if err := w.WriteMsgWithTimeout(messageTimeout, &pb.Ack{
		HopAddress: resp.Syn.HopAddress,
	}); err != nil {
		return nil, fmt.Errorf("write ack message: %w", err)
	}

	s.logger.Tracef("handshake finished for peer %s", swarm.NewAddress(resp.Syn.HopAddress.Overlay).String())
	return &Info{
		HopAddress: HopAddress,
		Light:      resp.Syn.Light,
	}, nil
}

func (s *Service) Handle(stream p2p.Stream, peerID libp2ppeer.ID) (i *Info, err error) {
	s.receivedHandshakesMu.Lock()
	if _, exists := s.receivedHandshakes[peerID]; exists {
		s.receivedHandshakesMu.Unlock()
		return nil, ErrHandshakeDuplicate
	}

	s.receivedHandshakes[peerID] = struct{}{}
	s.receivedHandshakesMu.Unlock()
	w, r := protobuf.NewWriterAndReader(stream)

	var req pb.Syn
	if err := r.ReadMsgWithTimeout(messageTimeout, &req); err != nil {
		return nil, fmt.Errorf("read syn message: %w", err)
	}

	if req.NetworkID != s.networkID {
		return nil, ErrNetworkIDIncompatible
	}

	HopAddress, err := hop.ParseAddress(req.HopAddress.Underlay, req.HopAddress.Overlay, req.HopAddress.Signature, req.NetworkID)
	if err != nil {
		return nil, ErrInvalidHopAddress
	}

	if err := w.WriteMsgWithTimeout(messageTimeout, &pb.SynAck{
		Syn: &pb.Syn{
			HopAddress: &pb.HopAddress{
				Underlay:  s.HopAddress.Underlay.Bytes(),
				Signature: s.HopAddress.Signature,
				Overlay:   s.HopAddress.Overlay.Bytes(),
			},
			NetworkID: s.networkID,
		},
		Ack: &pb.Ack{HopAddress: req.HopAddress},
	}); err != nil {
		return nil, fmt.Errorf("write synack message: %w", err)
	}

	var ack pb.Ack
	if err := r.ReadMsgWithTimeout(messageTimeout, &ack); err != nil {
		return nil, fmt.Errorf("read ack message: %w", err)
	}

	if err := s.checkAck(&ack); err != nil {
		return nil, err
	}

	s.logger.Tracef("handshake finished for peer %s", swarm.NewAddress(req.HopAddress.Overlay).String())
	return &Info{
		HopAddress: HopAddress,
		Light:      req.Light,
	}, nil
}

func (s *Service) Disconnected(_ network.Network, c network.Conn) {
	s.receivedHandshakesMu.Lock()
	defer s.receivedHandshakesMu.Unlock()
	delete(s.receivedHandshakes, c.RemotePeer())
}

func (s *Service) checkAck(ack *pb.Ack) error {
	if !bytes.Equal(ack.HopAddress.Overlay, s.HopAddress.Overlay.Bytes()) ||
		!bytes.Equal(ack.HopAddress.Underlay, s.HopAddress.Underlay.Bytes()) ||
		!bytes.Equal(ack.HopAddress.Signature, s.HopAddress.Signature) {
		return ErrInvalidAck
	}

	return nil
}

type Info struct {
	HopAddress *hop.Address
	Light      bool
}
