package blocker_test

import (
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/redesblock/mop/core/flock"
	"github.com/redesblock/mop/core/flock/test"
	"github.com/redesblock/mop/core/logging"

	"github.com/redesblock/mop/core/blocker"
)

func TestBlocksAfterFlagTimeout(t *testing.T) {

	mu := sync.Mutex{}
	blocked := make(map[string]time.Duration)

	mock := mockBlockLister(func(a flock.Address, d time.Duration, r string) error {
		mu.Lock()
		blocked[a.ByteString()] = d
		mu.Unlock()

		return nil
	})

	logger := logging.New(ioutil.Discard, 0)

	flagTime := time.Millisecond * 25
	checkTime := time.Millisecond * 100
	blockTime := time.Second

	b := blocker.New(mock, flagTime, blockTime, time.Millisecond, nil, logger)

	addr := test.RandomAddress()
	b.Flag(addr)

	mu.Lock()
	if _, ok := blocked[addr.ByteString()]; ok {
		mu.Unlock()
		t.Fatal("blocker did not wait flag duration")
	}
	mu.Unlock()

	midway := time.After(flagTime / 2)
	check := time.After(checkTime)

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

	b.Close()
}

func TestUnflagBeforeBlock(t *testing.T) {

	mu := sync.Mutex{}
	blocked := make(map[string]time.Duration)

	mock := mockBlockLister(func(a flock.Address, d time.Duration, r string) error {
		mu.Lock()
		blocked[a.ByteString()] = d
		mu.Unlock()
		return nil
	})

	logger := logging.New(ioutil.Discard, 0)

	flagTime := time.Millisecond * 25
	checkTime := time.Millisecond * 100
	blockTime := time.Second

	b := blocker.New(mock, flagTime, blockTime, time.Millisecond, nil, logger)

	addr := test.RandomAddress()
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

	b.Close()
}

type blocklister struct {
	blocklistFunc func(flock.Address, time.Duration, string) error
}

func mockBlockLister(f func(flock.Address, time.Duration, string) error) *blocklister {
	return &blocklister{
		blocklistFunc: f,
	}
}

func (b *blocklister) Blocklist(addr flock.Address, t time.Duration, r string) error {
	return b.blocklistFunc(addr, t, r)
}
