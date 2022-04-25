package handshake

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redesblock/hop/core/crypto"
	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/p2p"
	"github.com/redesblock/hop/core/p2p/libp2p/internal/handshake/pb"
	"github.com/redesblock/hop/core/p2p/protobuf"
	"github.com/redesblock/hop/core/swarm"

	"github.com/libp2p/go-libp2p-core/network"
	libp2ppeer "github.com/libp2p/go-libp2p-core/peer"
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

	// ErrInvalidSignature is returned if peer info was received with invalid signature
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrInvalidAck is returned if ack does not match the syn provided
	ErrInvalidAck = errors.New("invalid ack")
)

// PeerFinder has the information if the peer already exists in swarm.
type PeerFinder interface {
	Exists(overlay swarm.Address) (found bool)
}

type Service struct {
	overlay              swarm.Address
	underlay             []byte
	signature            []byte
	signer               crypto.Signer
	networkID            uint64
	networkIDBytes       []byte
	receivedHandshakes   map[libp2ppeer.ID]struct{}
	receivedHandshakesMu sync.Mutex
	logger               logging.Logger

	network.Notifiee // handhsake service can be the receiver for network.Notify
}

func New(overlay swarm.Address, peerID libp2ppeer.ID, signer crypto.Signer, networkID uint64, logger logging.Logger) (*Service, error) {
	underlay, err := peerID.MarshalBinary()
	if err != nil {
		return nil, err
	}

	networkIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(networkIDBytes, networkID)
	signature, err := signer.Sign(append(underlay, networkIDBytes...))
	if err != nil {
		return nil, err
	}

	return &Service{
		overlay:            overlay,
		underlay:           underlay,
		signature:          signature,
		signer:             signer,
		networkID:          networkID,
		networkIDBytes:     networkIDBytes,
		receivedHandshakes: make(map[libp2ppeer.ID]struct{}),
		logger:             logger,
		Notifiee:           new(network.NoopNotifiee),
	}, nil
}

func (s *Service) Handshake(stream p2p.Stream) (i *Info, err error) {
	w, r := protobuf.NewWriterAndReader(stream)
	if err := w.WriteMsgWithTimeout(messageTimeout, &pb.Syn{
		HopAddress: &pb.HopAddress{
			Underlay:  s.underlay,
			Signature: s.signature,
			Overlay:   s.overlay.Bytes(),
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

	if err := s.checkSyn(resp.Syn); err != nil {
		return nil, err
	}

	if err := w.WriteMsgWithTimeout(messageTimeout, &pb.Ack{
		HopAddress: resp.Syn.HopAddress,
	}); err != nil {
		return nil, fmt.Errorf("write ack message: %w", err)
	}

	s.logger.Tracef("handshake finished for peer %s", swarm.NewAddress(resp.Syn.HopAddress.Overlay).String())
	return &Info{
		Overlay:   swarm.NewAddress(resp.Syn.HopAddress.Overlay),
		Underlay:  resp.Syn.HopAddress.Underlay,
		NetworkID: resp.Syn.NetworkID,
		Light:     resp.Syn.Light,
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

	if err := s.checkSyn(&req); err != nil {
		return nil, err
	}

	if err := w.WriteMsgWithTimeout(messageTimeout, &pb.SynAck{
		Syn: &pb.Syn{
			HopAddress: &pb.HopAddress{
				Underlay:  s.underlay,
				Signature: s.signature,
				Overlay:   s.overlay.Bytes(),
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

	s.logger.Tracef("handshake finished for peer %s", req.HopAddress.Overlay)
	return &Info{
		Overlay:   swarm.NewAddress(req.HopAddress.Overlay),
		Underlay:  req.HopAddress.Underlay,
		NetworkID: req.NetworkID,
		Light:     req.Light,
	}, nil
}

func (s *Service) Disconnected(_ network.Network, c network.Conn) {
	s.receivedHandshakesMu.Lock()
	defer s.receivedHandshakesMu.Unlock()
	delete(s.receivedHandshakes, c.RemotePeer())
}

func (s *Service) checkSyn(syn *pb.Syn) error {
	if syn.NetworkID != s.networkID {
		return ErrNetworkIDIncompatible
	}

	recoveredPK, err := crypto.Recover(syn.HopAddress.Signature, append(syn.HopAddress.Underlay, s.networkIDBytes...))
	if err != nil {
		return ErrInvalidSignature
	}

	recoveredOverlay := crypto.NewOverlayAddress(*recoveredPK, syn.NetworkID)
	if !bytes.Equal(recoveredOverlay.Bytes(), syn.HopAddress.Overlay) {
		return ErrInvalidSignature
	}

	return nil
}

func (s *Service) checkAck(ack *pb.Ack) error {
	if !bytes.Equal(ack.HopAddress.Overlay, s.overlay.Bytes()) ||
		!bytes.Equal(ack.HopAddress.Underlay, s.underlay) ||
		!bytes.Equal(ack.HopAddress.Signature, s.signature) {
		return ErrInvalidAck
	}

	return nil
}

type Info struct {
	Overlay   swarm.Address
	Underlay  []byte
	NetworkID uint64
	Light     bool
}
