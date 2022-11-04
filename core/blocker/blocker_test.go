package blocker_test

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/cluster/test"
	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/p2p"

	"github.com/redesblock/mop/core/blocker"
)

var (
	flagTime  = time.Millisecond * 25
	checkTime = time.Millisecond * 100
	blockTime = time.Second
	addr      = test.RandomAddress()
	logger    = log.Noop
)

func TestMain(m *testing.M) {
	defer func(resolution time.Duration) {
		*blocker.SequencerResolution = resolution
	}(*blocker.SequencerResolution)
	*blocker.SequencerResolution = time.Millisecond

	os.Exit(m.Run())
}

func TestBlocksAfterFlagTimeout(t *testing.T) {
	var (
		mu      sync.Mutex
		blocked = make(map[string]time.Duration)
		mock    = mockBlockLister(func(a cluster.Address, d time.Duration, r string) error {
			mu.Lock()
			blocked[a.ByteString()] = d
			mu.Unlock()

			return nil
		})
		b = blocker.New(mock, flagTime, blockTime, time.Millisecond, nil, logger)
	)
	defer b.Close()

	b.Flag(addr)

	mu.Lock()
	if _, ok := blocked[addr.ByteString()]; ok {
		mu.Unlock()
		t.Fatal("blocker did not wait flag duration")
	}
	mu.Unlock()

	midway := time.After(flagTime / 2)
	check := time.After(checkTime * 5)

	<-midway
	b.Flag(addr) // check thats this flag call does not overide previous call
	<-check

	mu.Lock()
	blockedTime, ok := blocked[addr.ByteString()]
	mu.Unlock()
	if !ok {
		t.Fatal("address should be blocked")
	}

	if blockedTime != blockTime {
		t.Fatalf("block time: want %v, got %v", blockTime, blockedTime)
	}
}

func TestUnflagBeforeBlock(t *testing.T) {
	var (
		mu      sync.Mutex
		blocked = make(map[string]time.Duration)
		mock    = mockBlockLister(func(a cluster.Address, d time.Duration, r string) error {
			mu.Lock()
			blocked[a.ByteString()] = d
			mu.Unlock()
			return nil
		})
		b = blocker.New(mock, flagTime, blockTime, time.Millisecond, nil, logger)
	)
	defer b.Close()
	b.Flag(addr)

	mu.Lock()
	if _, ok := blocked[addr.ByteString()]; ok {
		mu.Unlock()
		t.Fatal("blocker did not wait flag duration")
	}
	mu.Unlock()

	b.Unflag(addr)

	time.Sleep(checkTime)

	mu.Lock()
	_, ok := blocked[addr.ByteString()]
	mu.Unlock()

	if ok {
		t.Fatal("address should not be blocked")
	}

}

func TestPruneBeforeBlock(t *testing.T) {
	var (
		mu      sync.Mutex
		blocked = make(map[string]time.Duration)
		mock    = mockBlockLister(func(a cluster.Address, d time.Duration, r string) error {
			mu.Lock()
			blocked[a.ByteString()] = d
			mu.Unlock()
			return nil
		})
		b = blocker.New(mock, flagTime, blockTime, time.Millisecond, nil, logger)
	)
	defer b.Close()

	b.Flag(addr)

	mu.Lock()
	if _, ok := blocked[addr.ByteString()]; ok {
		mu.Unlock()
		t.Fatal("blocker did not wait flag duration")
	}
	mu.Unlock()

	// communicate that we have seen no peers, resulting in the peer being removed
	b.PruneUnseen([]cluster.Address{})

	time.Sleep(checkTime)

	mu.Lock()
	_, ok := blocked[addr.ByteString()]
	mu.Unlock()

	if ok {
		t.Fatal("address should not be blocked")
	}

}

type blocklister struct {
	blocklistFunc func(cluster.Address, time.Duration, string) error
}

func mockBlockLister(f func(cluster.Address, time.Duration, string) error) *blocklister {
	return &blocklister{
		blocklistFunc: f,
	}
}

func (b *blocklister) Blocklist(addr cluster.Address, t time.Duration, r string) error {
	return b.blocklistFunc(addr, t, r)
}

// NetworkStatus implements p2p.NetworkStatuser interface.
// It always returns p2p.NetworkStatusAvailable.
func (b *blocklister) NetworkStatus() p2p.NetworkStatus {
	return p2p.NetworkStatusAvailable
}
