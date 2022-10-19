package blocklist

import (
	"strings"
	"time"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/p2p"
	"github.com/redesblock/mop/core/storer/storage"
)

var keyPrefix = "blocklist-"

type currentTimeFn = func() time.Time

type Blocklist struct {
	store         storage.StateStorer
	currentTimeFn currentTimeFn
}

func NewBlocklist(store storage.StateStorer) *Blocklist {
	return &Blocklist{
		store:         store,
		currentTimeFn: time.Now,
	}
}

type entry struct {
	Timestamp time.Time `json:"timestamp"`
	Duration  string    `json:"duration"` // Duration is string because the time.Duration does not implement MarshalJSON/UnmarshalJSON methods.
}

func (b *Blocklist) Exists(overlay cluster.Address) (bool, error) {
	key := generateKey(overlay)
	timestamp, duration, err := b.get(key)
	if err != nil {
		if err == storage.ErrNotFound {
			return false, nil
		}

		return false, err
	}

	if b.currentTimeFn().Sub(timestamp) > duration && duration != 0 {
		_ = b.store.Delete(key)
		return false, nil
	}

	return true, nil
}

func (b *Blocklist) Add(overlay cluster.Address, duration time.Duration) (err error) {
	key := generateKey(overlay)
	_, d, err := b.get(key)
	if err != nil {
		if err != storage.ErrNotFound {
			return err
		}
	}

	// if peer is already blacklisted, blacklist it for the maximum amount of time
	if duration < d && duration != 0 || d == 0 {
		duration = d
	}

	return b.store.Put(key, &entry{
		Timestamp: b.currentTimeFn(),
		Duration:  duration.String(),
	})
}

// Peers returns all currently blocklisted peers.
func (b *Blocklist) Peers() ([]p2p.Peer, error) {
	var peers []p2p.Peer
	if err := b.store.Iterate(keyPrefix, func(k, v []byte) (bool, error) {
		if !strings.HasPrefix(string(k), keyPrefix) {
			return true, nil
		}
		addr, err := unmarshalKey(string(k))
		if err != nil {
			return true, err
		}

		t, d, err := b.get(string(k))
		if err != nil {
			return true, err
		}

		if b.currentTimeFn().Sub(t) > d && d != 0 {
			// skip to the next item
			return false, nil
		}

		p := p2p.Peer{Address: addr}
		peers = append(peers, p)
		return false, nil
	}); err != nil {
		return nil, err
	}

	return peers, nil
}

func (b *Blocklist) get(key string) (timestamp time.Time, duration time.Duration, err error) {
	var e entry
	if err := b.store.Get(key, &e); err != nil {
		return time.Time{}, -1, err
	}

	duration, err = time.ParseDuration(e.Duration)
	if err != nil {
		return time.Time{}, -1, err
	}

	return e.Timestamp, duration, nil
}

func generateKey(overlay cluster.Address) string {
	return keyPrefix + overlay.String()
}

func unmarshalKey(s string) (cluster.Address, error) {
	addr := s[len(keyPrefix):] // trim prefix
	return cluster.ParseHexAddress(addr)
}
