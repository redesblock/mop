package hive

import (
	"context"
	"fmt"
	"time"

	"github.com/redesblock/hop/core/addressbook"
	"github.com/redesblock/hop/core/hive/pb"
	"github.com/redesblock/hop/core/hop"
	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/p2p"
	"github.com/redesblock/hop/core/p2p/protobuf"
	"github.com/redesblock/hop/core/swarm"
)

const (
	protocolName    = "hive"
	protocolVersion = "1.0.0"
	peersStreamName = "peers"
	messageTimeout  = 1 * time.Minute // maximum allowed time for a message to be read or written.
	maxBatchSize    = 30
)

type Service struct {
	streamer        p2p.Streamer
	addressBook     addressbook.GetPutter
	addPeersHandler func(context.Context, ...swarm.Address) error
	networkID       uint64
	logger          logging.Logger
}

func New(streamer p2p.Streamer, addressbook addressbook.GetPutter, networkID uint64, logger logging.Logger) *Service {
	return &Service{
		streamer:    streamer,
		logger:      logger,
		addressBook: addressbook,
		networkID:   networkID,
	}
}

func (s *Service) Protocol() p2p.ProtocolSpec {
	return p2p.ProtocolSpec{
		Name:    protocolName,
		Version: protocolVersion,
		StreamSpecs: []p2p.StreamSpec{
			{
				Name:    peersStreamName,
				Handler: s.peersHandler,
			},
		},
	}
}

func (s *Service) BroadcastPeers(ctx context.Context, addressee swarm.Address, peers ...swarm.Address) error {
	max := maxBatchSize
	for len(peers) > 0 {
		if max > len(peers) {
			max = len(peers)
		}
		if err := s.sendPeers(ctx, addressee, peers[:max]); err != nil {
			return err
		}

		peers = peers[max:]
	}

	return nil
}

func (s *Service) SetAddPeersHandler(h func(ctx context.Context, addr ...swarm.Address) error) {
	s.addPeersHandler = h
}

func (s *Service) sendPeers(ctx context.Context, peer swarm.Address, peers []swarm.Address) (err error) {
	stream, err := s.streamer.NewStream(ctx, peer, nil, protocolName, protocolVersion, peersStreamName)
	if err != nil {
		return fmt.Errorf("new stream: %w", err)
	}
	defer func() {
		if err != nil {
			_ = stream.Reset()
		} else {
			_ = stream.FullClose()
		}
	}()
	w, _ := protobuf.NewWriterAndReader(stream)
	var peersRequest pb.Peers
	for _, p := range peers {
		addr, err := s.addressBook.Get(p)
		if err != nil {
			if err == addressbook.ErrNotFound {
				s.logger.Debugf("hive broadcast peers: peer not found in the addressbook. Skipping peer %s", p)
				continue
			}
			return err
		}

		peersRequest.Peers = append(peersRequest.Peers, &pb.HopAddress{
			Overlay:   addr.Overlay.Bytes(),
			Underlay:  addr.Underlay.Bytes(),
			Signature: addr.Signature,
		})
	}

	if err := w.WriteMsg(&peersRequest); err != nil {
		return fmt.Errorf("write Peers message: %w", err)
	}

	return nil
}

func (s *Service) peersHandler(ctx context.Context, peer p2p.Peer, stream p2p.Stream) error {
	_, r := protobuf.NewWriterAndReader(stream)
	var peersReq pb.Peers
	if err := r.ReadMsgWithTimeout(messageTimeout, &peersReq); err != nil {
		_ = stream.Reset()
		return fmt.Errorf("read requestPeers message: %w", err)
	}

	// close the stream before processing in order to unblock the sending side
	// fullclose is called async because there is no need to wait for conformation,
	// but we still want to handle not closed stream from the other side to avoid zombie stream
	go stream.FullClose()

	var peers []swarm.Address
	for _, newPeer := range peersReq.Peers {
		hopAddress, err := hop.ParseAddress(newPeer.Underlay, newPeer.Overlay, newPeer.Signature, s.networkID)
		if err != nil {
			s.logger.Warningf("skipping peer in response %s: %v", newPeer.String(), err)
			continue
		}

		err = s.addressBook.Put(hopAddress.Overlay, *hopAddress)
		if err != nil {
			s.logger.Warningf("skipping peer in response %s: %v", newPeer.String(), err)
			continue
		}

		peers = append(peers, hopAddress.Overlay)
	}

	if s.addPeersHandler != nil {
		if err := s.addPeersHandler(ctx, peers...); err != nil {
			return err
		}
	}

	return nil
}
