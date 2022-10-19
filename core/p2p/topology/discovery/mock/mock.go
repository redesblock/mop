package mock

import (
	"context"
	"sync"

	"github.com/redesblock/mop/core/cluster"
)

type Discovery struct {
	mtx           sync.Mutex
	ctr           int //how many ops
	records       map[string][]cluster.Address
	broadcastFunc func(context.Context, cluster.Address, ...cluster.Address) error
}

type Option interface {
	apply(*Discovery)
}
type optionFunc func(*Discovery)

func (f optionFunc) apply(r *Discovery) { f(r) }

func WithBroadcastPeers(f func(context.Context, cluster.Address, ...cluster.Address) error) optionFunc {
	return optionFunc(func(r *Discovery) {
		r.broadcastFunc = f
	})
}

func NewDiscovery(opts ...Option) *Discovery {
	d := &Discovery{
		records: make(map[string][]cluster.Address),
	}
	for _, opt := range opts {
		opt.apply(d)
	}
	return d
}

func (d *Discovery) BroadcastPeers(ctx context.Context, addressee cluster.Address, peers ...cluster.Address) error {
	if d.broadcastFunc != nil {
		return d.broadcastFunc(ctx, addressee, peers...)
	}
	for _, peer := range peers {
		d.mtx.Lock()
		d.records[addressee.String()] = append(d.records[addressee.String()], peer)
		d.mtx.Unlock()
	}
	d.mtx.Lock()
	d.ctr++
	d.mtx.Unlock()
	return nil
}

func (d *Discovery) Broadcasts() int {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	return d.ctr
}

func (d *Discovery) AddresseeRecords(addressee cluster.Address) (peers []cluster.Address, exists bool) {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	peers, exists = d.records[addressee.String()]
	return
}

func (d *Discovery) Reset() {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	d.ctr = 0
	d.records = make(map[string][]cluster.Address)
}
