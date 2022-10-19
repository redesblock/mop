package blocklist_test

import (
	"testing"
	"time"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/p2p"
	"github.com/redesblock/mop/core/p2p/libp2p/internal/blocklist"
	"github.com/redesblock/mop/core/storer/statestore/mock"
)

func TestExist(t *testing.T) {
	addr1 := cluster.NewAddress([]byte{0, 1, 2, 3})
	addr2 := cluster.NewAddress([]byte{4, 5, 6, 7})
	ctMock := &currentTimeMock{}

	bl := blocklist.NewBlocklistWithCurrentTimeFn(mock.NewStateStore(), ctMock.Time)

	exists, err := bl.Exists(addr1)
	if err != nil {
		t.Fatal(err)
	}

	if exists {
		t.Fatal("got exists, expected not exists")
	}

	// add forever
	if err := bl.Add(addr1, 0); err != nil {
		t.Fatal(err)
	}

	// add for 50 miliseconds
	if err := bl.Add(addr2, time.Millisecond*50); err != nil {
		t.Fatal(err)
	}

	ctMock.SetTime(time.Now().Add(100 * time.Millisecond))

	exists, err = bl.Exists(addr1)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("got not exists, expected exists")
	}

	exists, err = bl.Exists(addr2)
	if err != nil {
		t.Fatal(err)
	}

	if exists {
		t.Fatal("got  exists, expected not exists")
	}
}

func TestPeers(t *testing.T) {
	addr1 := cluster.NewAddress([]byte{0, 1, 2, 3})
	addr2 := cluster.NewAddress([]byte{4, 5, 6, 7})
	ctMock := &currentTimeMock{}

	bl := blocklist.NewBlocklistWithCurrentTimeFn(mock.NewStateStore(), ctMock.Time)

	// add forever
	if err := bl.Add(addr1, 0); err != nil {
		t.Fatal(err)
	}

	// add for 50 miliseconds
	if err := bl.Add(addr2, time.Millisecond*50); err != nil {
		t.Fatal(err)
	}

	peers, err := bl.Peers()
	if err != nil {
		t.Fatal(err)
	}
	if !isIn(addr1, peers) {
		t.Fatalf("expected addr1 to exist in peers: %v", addr1)
	}

	if !isIn(addr2, peers) {
		t.Fatalf("expected addr2 to exist in peers: %v", addr2)
	}

	ctMock.SetTime(time.Now().Add(100 * time.Millisecond))

	// now expect just one
	peers, err = bl.Peers()
	if err != nil {
		t.Fatal(err)
	}
	if !isIn(addr1, peers) {
		t.Fatalf("expected addr1 to exist in peers: %v", peers)
	}

	if isIn(addr2, peers) {
		t.Fatalf("expected addr2 to not exist in peers: %v", peers)
	}
}

func isIn(p cluster.Address, peers []p2p.Peer) bool {
	for _, v := range peers {
		if v.Address.Equal(p) {
			return true
		}
	}
	return false
}

type currentTimeMock struct {
	time time.Time
}

func (c *currentTimeMock) Time() time.Time {
	return c.time
}

func (c *currentTimeMock) SetTime(t time.Time) {
	c.time = t
}
