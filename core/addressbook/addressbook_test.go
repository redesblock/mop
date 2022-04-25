package addressbook_test

import (
	"testing"

	"github.com/redesblock/hop/core/addressbook"
	"github.com/redesblock/hop/core/statestore/mock"
	"github.com/redesblock/hop/core/swarm"

	ma "github.com/multiformats/go-multiaddr"
)

type bookFunc func(t *testing.T) (book addressbook.GetPutter)

func TestInMem(t *testing.T) {
	run(t, func(t *testing.T) addressbook.GetPutter {
		store := mock.NewStateStore()
		book := addressbook.New(store)

		return book
	})
}

func run(t *testing.T, f bookFunc) {
	store := f(t)

	addr1 := swarm.NewAddress([]byte{0, 1, 2, 3})
	addr2 := swarm.NewAddress([]byte{0, 1, 2, 4})
	multiaddr, err := ma.NewMultiaddr("/ip4/1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}

	err = store.Put(addr1, multiaddr)
	if err != nil {
		t.Fatal(err)
	}

	v, err := store.Get(addr1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Get(addr2)
	if err == nil {
		t.Fatal("value found in store but should not have been")
	}

	if multiaddr.String() != v.String() {
		t.Fatalf("value retrieved from store not equal to original stored address: %v, want %v", v, multiaddr)
	}
}
