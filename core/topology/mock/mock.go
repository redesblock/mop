package mock

import (
	"context"
	"sync"

	"github.com/redesblock/hop/core/swarm"
)

type TopologyDriver struct {
	peers      []swarm.Address
	addPeerErr error
	mtx        sync.Mutex
}

func NewTopologyDriver() *TopologyDriver {
	return &TopologyDriver{}
}

func (d *TopologyDriver) SetAddPeerErr(err error) {
	d.addPeerErr = err
}

func (d *TopologyDriver) AddPeer(_ context.Context, addr swarm.Address) error {
	if d.addPeerErr != nil {
		return d.addPeerErr
	}

	d.mtx.Lock()
	d.peers = append(d.peers, addr)
	d.mtx.Unlock()
	return nil
}

func (d *TopologyDriver) Peers() []swarm.Address {
	return d.peers
}
