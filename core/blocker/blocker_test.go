package blocker_test

import (
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/swarm/test"

	"github.com/redesblock/hop/core/blocker"
)

func TestBlocksAfterFlagTimeout(t *testing.T) {

	mux := sync.Mutex{}
	blocked := make(map[string]time.Duration)

	mock := mockBlockLister(func(a swarm.Address, d time.Duration, r string) error {
		mux.Lock()
		blocked[a.ByteString()] = d
		mux.Unlock()

		return nil
	})

	logger := logging.New(ioutil.Discard, 0)

	flagTime := time.Millisecond * 25
	checkTime := time.Millisecond * 100
	blockTime := time.Second

	b := blocker.New(mock, flagTime, blockTime, time.Millisecond, logger)

	addr := test.RandomAddress()
	b.Flag(addr)

	if _, ok := blocked[addr.ByteString()]; ok {
		t.Fatal("blocker did not wait flag duration")
	}

	midway := time.After(flagTime / 2)
	check := time.After(checkTime)

	<-midway
	b.Flag(addr) // check thats this flag call does not overide previous call
	<-check

	mux.Lock()
	blockedTime, ok := blocked[addr.ByteString()]
	mux.Unlock()
	if !ok {
		t.Fatal("address should be blocked")
	}

	if blockedTime != blockTime {
		t.Fatalf("block time: want %v, got %v", blockTime, blockedTime)
	}

	b.Close()
}

func TestUnflagBeforeBlock(t *testing.T) {

	mux := sync.Mutex{}
	blocked := make(map[string]time.Duration)

	mock := mockBlockLister(func(a swarm.Address, d time.Duration, r string) error {
		mux.Lock()
		blocked[a.ByteString()] = d
		mux.Unlock()
		return nil
	})

	logger := logging.New(ioutil.Discard, 0)

	flagTime := time.Millisecond * 25
	checkTime := time.Millisecond * 100
	blockTime := time.Second

	b := blocker.New(mock, flagTime, blockTime, time.Millisecond, logger)

	addr := test.RandomAddress()
	b.Flag(addr)

	if _, ok := blocked[addr.ByteString()]; ok {
		t.Fatal("blocker did not wait flag duration")
	}

	b.Unflag(addr)

	time.Sleep(checkTime)

	mux.Lock()
	_, ok := blocked[addr.ByteString()]
	mux.Unlock()

	if ok {
		t.Fatal("address should not be blocked")
	}

	b.Close()
}

type blocklister struct {
	blocklistFunc func(swarm.Address, time.Duration, string) error
}

func mockBlockLister(f func(swarm.Address, time.Duration, string) error) *blocklister {
	return &blocklister{
		blocklistFunc: f,
	}
}

func (b *blocklister) Blocklist(addr swarm.Address, t time.Duration, r string) error {
	return b.blocklistFunc(addr, t, r)
}
