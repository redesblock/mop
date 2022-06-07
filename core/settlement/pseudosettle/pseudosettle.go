package pseudosettle

import (
	"context"
	"fmt"
	"time"

	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/p2p"
	"github.com/redesblock/hop/core/p2p/protobuf"
	"github.com/redesblock/hop/core/settlement"
	pb "github.com/redesblock/hop/core/settlement/pseudosettle/pb"
	"github.com/redesblock/hop/core/swarm"
)

const (
	protocolName    = "pseudosettle"
	protocolVersion = "1.0.0"
	streamName      = "pseudosettle"
)

type Service struct {
	streamer p2p.Streamer
	logger   logging.Logger
	observer settlement.PaymentObserver
}

type Options struct {
	Streamer p2p.Streamer
	Logger   logging.Logger
}

func New(o Options) *Service {
	return &Service{
		streamer: o.Streamer,
		logger:   o.Logger,
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

func (s *Service) handler(ctx context.Context, p p2p.Peer, stream p2p.Stream) (err error) {
	r := protobuf.NewReader(stream)
	defer func() {
		if err != nil {
			_ = stream.Reset()
		} else {
			_ = stream.FullClose()
		}
	}()
	var req pb.Payment
	if err := r.ReadMsg(&req); err != nil {
		return fmt.Errorf("read request from peer %v: %w", p.Address, err)
	}

	s.logger.Tracef("received payment message from peer %v of %d", p.Address, req.Amount)
	return s.observer.NotifyPayment(p.Address, req.Amount)
}

// Pay initiates a payment to the given peer
func (s *Service) Pay(ctx context.Context, peer swarm.Address, amount uint64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	stream, err := s.streamer.NewStream(ctx, peer, nil, protocolName, protocolVersion, streamName)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = stream.Reset()
		} else {
			go stream.FullClose()
		}
	}()

	s.logger.Tracef("sending payment message to peer %v of %d", peer, amount)
	w := protobuf.NewWriter(stream)
	err = w.WriteMsgWithContext(ctx, &pb.Payment{
		Amount: amount,
	})
	return err
}

// SetPaymentObserver sets the payment observer which will be notified of incoming payments
func (s *Service) SetPaymentObserver(observer settlement.PaymentObserver) {
	s.observer = observer
}
