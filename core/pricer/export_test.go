package pricer

import (
	"github.com/redesblock/hop/core/swarm"
)

func (s *Pricer) PeerPricePO(peer swarm.Address, po uint8) (uint64, error) {
	return s.peerPricePO(peer, po)
}
