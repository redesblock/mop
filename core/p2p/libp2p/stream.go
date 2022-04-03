package libp2p

import (
	"github.com/libp2p/go-libp2p-core/helpers"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/redesblock/hop/core/p2p"
)

type stream struct {
	network.Stream
	headers map[string][]byte
}

func newStream(s network.Stream) *stream {
	return &stream{Stream: s}
}

func (s *stream) Headers() p2p.Headers {
	return s.headers
}

func (s *stream) FullClose() error {
	return helpers.FullClose(s)
}
