package lightnode

import (
	"context"
	"crypto/rand"
	"math/big"
	"sync"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/p2p"
	"github.com/redesblock/mop/core/p2p/topology"
	"github.com/redesblock/mop/core/p2p/topology/pslice"
)

type Container struct {
	base              cluster.Address
	peerMu            sync.Mutex // peerMu guards connectedPeers and disconnectedPeers.
	connectedPeers    *pslice.PSlice
	disconnectedPeers *pslice.PSlice
	metrics           metrics
}

func NewContainer(base cluster.Address) *Container {
	return &Container{
		base:              base,
		connectedPeers:    pslice.New(1, base),
		disconnectedPeers: pslice.New(1, base),
		metrics:           newMetrics(),
	}
}

func (c *Container) Connected(ctx context.Context, peer p2p.Peer) {
	c.peerMu.Lock()
	defer c.peerMu.Unlock()

	addr := peer.Address
	c.connectedPeers.Add(addr)
	c.disconnectedPeers.Remove(addr)

	c.metrics.CurrentlyConnectedPeers.Set(float64(c.connectedPeers.Length()))
	c.metrics.CurrentlyDisconnectedPeers.Set(float64(c.disconnectedPeers.Length()))
}

func (c *Container) Disconnected(peer p2p.Peer) {
	c.peerMu.Lock()
	defer c.peerMu.Unlock()

	addr := peer.Address
	if found := c.connectedPeers.Exists(addr); found {
		c.connectedPeers.Remove(addr)
		c.disconnectedPeers.Add(addr)
	}

	c.metrics.CurrentlyConnectedPeers.Set(float64(c.connectedPeers.Length()))
	c.metrics.CurrentlyDisconnectedPeers.Set(float64(c.disconnectedPeers.Length()))
}

func (c *Container) Count() int {
	return c.connectedPeers.Length()
}

func (c *Container) RandomPeer(not cluster.Address) (cluster.Address, error) {
	c.peerMu.Lock()
	defer c.peerMu.Unlock()
	var (
		cnt   = big.NewInt(int64(c.Count()))
		addr  = cluster.ZeroAddress
		count = int64(0)
	)

PICKPEER:
	i, e := rand.Int(rand.Reader, cnt)
	if e != nil {
		return cluster.ZeroAddress, e
	}
	i64 := i.Int64()

	count = 0
	_ = c.connectedPeers.EachBinRev(func(peer cluster.Address, _ uint8) (bool, bool, error) {
		if count == i64 {
			addr = peer
			return true, false, nil
		}
		count++
		return false, false, nil
	})

	if addr.Equal(not) {
		goto PICKPEER
	}

	return addr, nil
}

func (c *Container) EachPeer(pf topology.EachPeerFunc) error {
	return c.connectedPeers.EachBin(pf)
}

func (c *Container) PeerInfo() topology.BinInfo {
	return topology.BinInfo{
		BinPopulation:     uint(c.connectedPeers.Length()),
		BinConnected:      uint(c.connectedPeers.Length()),
		DisconnectedPeers: peersInfo(c.disconnectedPeers),
		ConnectedPeers:    peersInfo(c.connectedPeers),
	}
}

func peersInfo(s *pslice.PSlice) []*topology.PeerInfo {
	if s.Length() == 0 {
		return nil
	}
	peers := make([]*topology.PeerInfo, 0, s.Length())
	_ = s.EachBin(func(addr cluster.Address, po uint8) (bool, bool, error) {
		peers = append(peers, &topology.PeerInfo{Address: addr})
		return false, false, nil
	})
	return peers
}
