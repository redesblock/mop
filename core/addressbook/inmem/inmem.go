package inmem

import (
	"sync"

	"github.com/redesblock/hop/core/addressbook"
	"github.com/redesblock/hop/core/swarm"

	ma "github.com/multiformats/go-multiaddr"
)

type inmem struct {
	mtx     sync.Mutex
	entries map[string]peerEntry // key: overlay in string value, value: peerEntry
}

type peerEntry struct {
	overlay   swarm.Address
	multiaddr ma.Multiaddr
}

func New() addressbook.GetPutter {
	return &inmem{
		entries: make(map[string]peerEntry),
	}
}

func (i *inmem) Get(overlay swarm.Address) (addr ma.Multiaddr, exists bool) {
	i.mtx.Lock()
	defer i.mtx.Unlock()

	val, exists := i.entries[overlay.String()]
	return val.multiaddr, exists
}

func (i *inmem) Put(overlay swarm.Address, addr ma.Multiaddr) (exists bool) {
	i.mtx.Lock()
	defer i.mtx.Unlock()
	_, e := i.entries[overlay.String()]
	i.entries[overlay.String()] = peerEntry{overlay: overlay, multiaddr: addr}
	return e
}

func (i *inmem) Overlays() []swarm.Address {
	keys := make([]swarm.Address, 0, len(i.entries))
	for k := range i.entries {
		keys = append(keys, swarm.MustParseHexAddress(k))
	}

	return keys
}

func (i *inmem) Multiaddresses() []ma.Multiaddr {
	values := make([]ma.Multiaddr, 0, len(i.entries))
	for _, v := range i.entries {
		values = append(values, v.multiaddr)
	}

	return values
}
