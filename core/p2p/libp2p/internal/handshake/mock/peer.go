package mock

import "github.com/redesblock/mop/core/flock"

// todo: implement peer registry mocks, export appropriate interface and move those in libp2p so it can be used in handshake
type PeerFinder struct {
	found bool
}

func (p *PeerFinder) SetFound(found bool) {
	p.found = found
}

func (p *PeerFinder) Exists(overlay flock.Address) (found bool) {
	return p.found
}
