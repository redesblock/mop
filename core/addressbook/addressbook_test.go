package addressbook_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/redesblock/mop/core/addressbook"
	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/flock"
	"github.com/redesblock/mop/core/mop"
	"github.com/redesblock/mop/core/statestore/mock"

	ma "github.com/multiformats/go-multiaddr"
)

type bookFunc func(t *testing.T) (book addressbook.Interface)

func TestInMem(t *testing.T) {
	run(t, func(t *testing.T) addressbook.Interface {
		store := mock.NewStateStore()
		book := addressbook.New(store)
		return book
	})
}

func run(t *testing.T, f bookFunc) {
	store := f(t)
	addr1 := flock.NewAddress([]byte{0, 1, 2, 3})
	addr2 := flock.NewAddress([]byte{0, 1, 2, 4})
	trxHash := common.HexToHash("0x1").Bytes()
	multiaddr, err := ma.NewMultiaddr("/ip4/1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}

	pk, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatal(err)
	}

	mopAddr, err := mop.NewAddress(crypto.NewDefaultSigner(pk), multiaddr, addr1, 1, trxHash)
	if err != nil {
		t.Fatal(err)
	}

	err = store.Put(addr1, *mopAddr)
	if err != nil {
		t.Fatal(err)
	}

	v, err := store.Get(addr1)
	if err != nil {
		t.Fatal(err)
	}

	if !mopAddr.Equal(v) {
		t.Fatalf("expectted: %s, want %s", v, multiaddr)
	}

	notFound, err := store.Get(addr2)
	if err != addressbook.ErrNotFound {
		t.Fatal(err)
	}

	if notFound != nil {
		t.Fatalf("expected nil got %s", v)
	}

	overlays, err := store.Overlays()
	if err != nil {
		t.Fatal(err)
	}

	if len(overlays) != 1 {
		t.Fatalf("expected overlay len %v, got %v", 1, len(overlays))
	}

	addresses, err := store.Addresses()
	if err != nil {
		t.Fatal(err)
	}

	if len(addresses) != 1 {
		t.Fatalf("expected addresses len %v, got %v", 1, len(addresses))
	}
}
