package handshake

import (
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

	// ErrInvalidAck is returned if data in received in ack is not valid (invalid signature for example).
	ErrInvalidAck = errors.New("invalid ack")

	// ErrInvalidSyn is returned if observable address in ack is not a valid..
	ErrInvalidSyn = errors.New("invalid syn")
)

type AdvertisableAddressResolver interface {
	Resolve(observedAdddress ma.Multiaddr) (ma.Multiaddr, error)
}

type Service struct {
	signer                crypto.Signer
	advertisableAddresser AdvertisableAddressResolver
	overlay               swarm.Address
	lightNode             bool
	networkID             uint64
	receivedHandshakes    map[libp2ppeer.ID]struct{}
	receivedHandshakesMu  sync.Mutex
	logger                logging.Logger

	network.Notifiee // handshake service can be the receiver for network.Notify
}

func New(signer crypto.Signer, advertisableAddresser AdvertisableAddressResolver, overlay swarm.Address, networkID uint64, lighNode bool, logger logging.Logger) (*Service, error) {
	return &Service{
		signer:                signer,
		advertisableAddresser: advertisableAddresser,
		overlay:               overlay,
		networkID:             networkID,
		lightNode:             lighNode,
		receivedHandshakes:    make(map[libp2ppeer.ID]struct{}),
		logger:                logger,
		Notifiee:              new(network.NoopNotifiee),
	}, nil
}

func (s *Service) Handshake(stream p2p.Stream, peerMultiaddr ma.Multiaddr, peerID libp2ppeer.ID) (i *Info, err error) {
	w, r := protobuf.NewWriterAndReader(stream)
	fullRemoteMA, err := buildFullMA(peerMultiaddr, peerID)
	if err != nil {
		return nil, err
	}

	fullRemoteMABytes, err := fullRemoteMA.MarshalBinary()
	if err != nil {
		return nil, err
	}

	if err := w.WriteMsgWithTimeout(messageTimeout, &pb.Syn{
		ObservedUnderlay: fullRemoteMABytes,
	}); err != nil {
		return nil, fmt.Errorf("write syn message: %w", err)
	}

	var resp pb.SynAck
	if err := r.ReadMsgWithTimeout(messageTimeout, &resp); err != nil {
		return nil, fmt.Errorf("read synack message: %w", err)
	}

	remoteHopAddress, err := s.parseCheckAck(resp.Ack)
	if err != nil {
		return nil, err
	}

	observedUnderlay, err := ma.NewMultiaddrBytes(resp.Syn.ObservedUnderlay)
	if err != nil {
		return nil, ErrInvalidSyn
	}

	advertisableUnderlay, err := s.advertisableAddresser.Resolve(observedUnderlay)
	if err != nil {
		return nil, err
	}

	hopAddress, err := hop.NewAddress(s.signer, advertisableUnderlay, s.overlay, s.networkID)
	if err != nil {
		return nil, err
	}

	advertisableUnderlayBytes, err := hopAddress.Underlay.MarshalBinary()
	if err != nil {
		return nil, err
	}

	if err := w.WriteMsgWithTimeout(messageTimeout, &pb.Ack{
		Address: &pb.HopAddress{
			Underlay:  advertisableUnderlayBytes,
			Overlay:   hopAddress.Overlay.Bytes(),
			Signature: hopAddress.Signature,
		},
		NetworkID: s.networkID,
		Light:     s.lightNode,
	}); err != nil {
		return nil, fmt.Errorf("write ack message: %w", err)
	}

	s.logger.Tracef("handshake finished for peer (outbound) %s", remoteHopAddress.Overlay.String())
	return &Info{
		HopAddress: remoteHopAddress,
		Light:      resp.Ack.Light,
	}, nil
}

func (s *Service) Handle(stream p2p.Stream, remoteMultiaddr ma.Multiaddr, remotePeerID libp2ppeer.ID) (i *Info, err error) {
	s.receivedHandshakesMu.Lock()
	if _, exists := s.receivedHandshakes[remotePeerID]; exists {
		s.receivedHandshakesMu.Unlock()
		return nil, ErrHandshakeDuplicate
	}

	s.receivedHandshakes[remotePeerID] = struct{}{}
	s.receivedHandshakesMu.Unlock()
	w, r := protobuf.NewWriterAndReader(stream)
	fullRemoteMA, err := buildFullMA(remoteMultiaddr, remotePeerID)
	if err != nil {
		return nil, err
	}

	fullRemoteMABytes, err := fullRemoteMA.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var syn pb.Syn
	if err := r.ReadMsgWithTimeout(messageTimeout, &syn); err != nil {
		return nil, fmt.Errorf("read syn message: %w", err)
	}

	observedUnderlay, err := ma.NewMultiaddrBytes(syn.ObservedUnderlay)
	if err != nil {
		return nil, ErrInvalidSyn
	}

	advertisableUnderlay, err := s.advertisableAddresser.Resolve(observedUnderlay)
	if err != nil {
		return nil, err
	}

	hopAddress, err := hop.NewAddress(s.signer, advertisableUnderlay, s.overlay, s.networkID)
	if err != nil {
		return nil, err
	}

	advertisableUnderlayBytes, err := hopAddress.Underlay.MarshalBinary()
	if err != nil {
		return nil, err
	}

	if err := w.WriteMsgWithTimeout(messageTimeout, &pb.SynAck{
		Syn: &pb.Syn{
			ObservedUnderlay: fullRemoteMABytes,
		},
		Ack: &pb.Ack{
			Address: &pb.HopAddress{
				Underlay:  advertisableUnderlayBytes,
				Overlay:   hopAddress.Overlay.Bytes(),
				Signature: hopAddress.Signature,
			},
			NetworkID: s.networkID,
			Light:     s.lightNode,
		},
	}); err != nil {
		return nil, fmt.Errorf("write synack message: %w", err)
	}

	var ack pb.Ack
	if err := r.ReadMsgWithTimeout(messageTimeout, &ack); err != nil {
		return nil, fmt.Errorf("read ack message: %w", err)
	}

	remoteHopAddress, err := s.parseCheckAck(&ack)
	if err != nil {
		return nil, err
	}

	s.logger.Tracef("handshake finished for peer (inbound) %s", remoteHopAddress.Overlay.String())
	return &Info{
		HopAddress: remoteHopAddress,
		Light:      ack.Light,
	}, nil
}

func (s *Service) Disconnected(_ network.Network, c network.Conn) {
	s.receivedHandshakesMu.Lock()
	defer s.receivedHandshakesMu.Unlock()
	delete(s.receivedHandshakes, c.RemotePeer())
}

func buildFullMA(addr ma.Multiaddr, peerID libp2ppeer.ID) (ma.Multiaddr, error) {
	return ma.NewMultiaddr(fmt.Sprintf("%s/p2p/%s", addr.String(), peerID.Pretty()))
}

func (s *Service) parseCheckAck(ack *pb.Ack) (*hop.Address, error) {
	if ack.NetworkID != s.networkID {
		return nil, ErrNetworkIDIncompatible
	}

	hopAddress, err := hop.ParseAddress(ack.Address.Underlay, ack.Address.Overlay, ack.Address.Signature, s.networkID)
	if err != nil {
		return nil, ErrInvalidAck
	}

	return hopAddress, nil
}

type Info struct {
	HopAddress *hop.Address
	Light      bool
}
