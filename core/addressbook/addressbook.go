package addressbook

import (
	"github.com/redesblock/hop/core/swarm"

	ma "github.com/multiformats/go-multiaddr"
)

type GetPutter interface {
	Getter
	Putter
}

type Getter interface {
	Get(overlay swarm.Address) (addr ma.Multiaddr, exists bool)
}

type Putter interface {
	Put(overlay swarm.Address, addr ma.Multiaddr) (exists bool)
}
