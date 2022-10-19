package pricer

import (
	"github.com/redesblock/mop/core/cluster"
)

// Pricer returns pricing information for chunk hashes.
type Interface interface {
	// PeerPrice is the price the peer charges for a given chunk hash.
	PeerPrice(peer, chunk cluster.Address) uint64
	// Price is the price we charge for a given chunk hash.
	Price(chunk cluster.Address) uint64
}

// FixedPricer is a Pricer that has a fixed price for chunks.
type FixedPricer struct {
	overlay cluster.Address
	poPrice uint64
}

// NewFixedPricer returns a new FixedPricer with a given price.
func NewFixedPricer(overlay cluster.Address, poPrice uint64) *FixedPricer {
	return &FixedPricer{
		overlay: overlay,
		poPrice: poPrice,
	}
}

// PeerPrice implements Pricer.
func (pricer *FixedPricer) PeerPrice(peer, chunk cluster.Address) uint64 {
	return uint64(cluster.MaxPO-cluster.Proximity(peer.Bytes(), chunk.Bytes())+1) * pricer.poPrice
}

// Price implements Pricer.
func (pricer *FixedPricer) Price(chunk cluster.Address) uint64 {
	return pricer.PeerPrice(pricer.overlay, chunk)
}
