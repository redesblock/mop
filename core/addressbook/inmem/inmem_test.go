package inmem

import (
	"testing"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/redesblock/hop/core/swarm"
)

func TestInMemStore(t *testing.T) {
	mem := New()
	addr1 := swarm.NewAddress([]byte{0, 1, 2, 3})
	addr2 := swarm.NewAddress([]byte{0, 1, 2, 4})
	multiaddr, err := ma.NewMultiaddr("/ip4/1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}
	//var beep ma.Multiaddr
	exists := mem.Put(addr1, multiaddr)
	if exists {
		t.Fatal("object exists in store but shouldnt")
	}

	_, exists = mem.Get(addr2)
	if exists {
		t.Fatal("value found in store but should not have been")
	}

	v, exists := mem.Get(addr1)
	if !exists {
		t.Fatal("value not found in store but should have been")
	}

	if multiaddr.String() != v.String() {
		t.Fatalf("value retrieved from store not equal to original stored address: %v, want %v", v, multiaddr)
	}
}
