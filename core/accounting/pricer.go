package accounting

import (
	"github.com/redesblock/hop/core/swarm"
)

type Pricer interface {
	// PeerPrice is the price the peer charges for a given chunk hash
	PeerPrice(peer, chunk swarm.Address) uint64
	// Price is the price we charge for a given chunk hash
	Price(chunk swarm.Address) uint64
}

type FixedPricer struct {
	overlay swarm.Address
	poPrice uint64
}

func NewFixedPricer(overlay swarm.Address, poPrice uint64) *FixedPricer {
	return &FixedPricer{
		overlay: overlay,
		poPrice: poPrice,
	}
}

func (pricer *FixedPricer) PeerPrice(peer, chunk swarm.Address) uint64 {
	return uint64(swarm.MaxPO-swarm.Proximity(peer.Bytes(), chunk.Bytes())) * pricer.poPrice
}

func (pricer *FixedPricer) Price(chunk swarm.Address) uint64 {
	return pricer.PeerPrice(pricer.overlay, chunk)
}
