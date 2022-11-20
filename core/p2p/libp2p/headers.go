package libp2p

import (
	"context"
	"fmt"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/p2p"
	"github.com/redesblock/mop/core/p2p/libp2p/internal/headers/pb"
	"github.com/redesblock/mop/core/p2p/protobuf"
)

func sendHeaders(ctx context.Context, headers p2p.Headers, stream *stream) error {
	w, r := protobuf.NewWriterAndReader(stream)

	if err := w.WriteMsgWithContext(ctx, headersP2PToPB(headers)); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	h := new(pb.Headers)
	if err := r.ReadMsgWithContext(ctx, h); err != nil {
		return fmt.Errorf("read message: %w", err)
	}

	stream.headers = headersPBToP2P(h)

	return nil
}

func handleHeaders(ctx context.Context, headler p2p.HeadlerFunc, stream *stream, peerAddress cluster.Address) error {
	w, r := protobuf.NewWriterAndReader(stream)

	headers := new(pb.Headers)
	if err := r.ReadMsgWithContext(ctx, headers); err != nil {
		return fmt.Errorf("read message: %w", err)
	}

	stream.headers = headersPBToP2P(headers)

	var h p2p.Headers
	if headler != nil {
		h = headler(stream.headers, peerAddress)
	}

	stream.responseHeaders = h

	if err := w.WriteMsgWithContext(ctx, headersP2PToPB(h)); err != nil {
		return fmt.Errorf("write message: %w", err)
	}
	return nil
}

func headersPBToP2P(h *pb.Headers) p2p.Headers {
	p2ph := make(p2p.Headers)
	for _, rh := range h.Headers {
		p2ph[rh.Key] = rh.Value
	}
	return p2ph
}

func headersP2PToPB(h p2p.Headers) *pb.Headers {
	pbh := new(pb.Headers)
	pbh.Headers = make([]*pb.Header, 0)
	for key, value := range h {
		pbh.Headers = append(pbh.Headers, &pb.Header{
			Key:   key,
			Value: value,
		})
	}
	return pbh
}
