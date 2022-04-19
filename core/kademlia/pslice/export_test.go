package pslice

import "github.com/redesblock/hop/core/swarm"

func PSlicePeers(p *PSlice) []swarm.Address {
	return p.peers
}

func PSliceBins(p *PSlice) []uint {
	return p.bins
}
