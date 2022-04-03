package mock

import "github.com/redesblock/hop/core/swarm"

// todo: implement peer registry mocks, export appropriate interface and move those in libp2p so it can be used in handshake
type PeerFinder struct {
	found bool
}

func (p *PeerFinder) SetFound(found bool) {
	p.found = found
}

func (p *PeerFinder) Exists(overlay swarm.Address) (found bool) {
	return p.found
}
