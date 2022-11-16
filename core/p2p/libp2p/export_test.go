package libp2p

import (
	"context"

	libp2pm "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	libp2ppeer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/redesblock/mop/core/p2p/libp2p/internal/handshake"
)

func (s *Service) HandshakeService() *handshake.Service {
	return s.handshakeService
}

func (s *Service) NewStreamForPeerID(peerID libp2ppeer.ID, protocolName, protocolVersion, streamName string) (network.Stream, error) {
	return s.newStreamForPeerID(context.Background(), peerID, protocolName, protocolVersion, streamName)
}

func (s *Service) Host() host.Host {
	return s.host
}

type StaticAddressResolver = staticAddressResolver

var (
	NewStaticAddressResolver = newStaticAddressResolver
	UserAgent                = userAgent
)

func WithHostFactory(factory func(...libp2pm.Option) (host.Host, error)) Options {
	return Options{
		hostFactory: factory,
	}
}
