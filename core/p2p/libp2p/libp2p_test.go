package libp2p_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"sort"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiformats/go-multiaddr"
	"github.com/redesblock/mop/core/address"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/p2p"
	"github.com/redesblock/mop/core/p2p/libp2p"
	"github.com/redesblock/mop/core/p2p/topology/lightnode"
	"github.com/redesblock/mop/core/storer/statestore/mock"
)

type libp2pServiceOpts struct {
	Logger      log.Logger
	Addressbook address.Interface
	PrivateKey  *ecdsa.PrivateKey
	MockPeerKey *ecdsa.PrivateKey
	libp2pOpts  libp2p.Options
	lightNodes  *lightnode.Container
	notifier    p2p.PickyNotifier
}

// newService constructs a new libp2p service.
func newService(t *testing.T, networkID uint64, o libp2pServiceOpts) (s *libp2p.Service, overlay cluster.Address) {
	t.Helper()

	clusterKey, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatal(err)
	}

	nonce := common.HexToHash("0x1").Bytes()

	overlay, err = crypto.NewOverlayAddress(clusterKey.PublicKey, networkID, nonce)
	if err != nil {
		t.Fatal(err)
	}

	addr := ":0"

	if o.Logger == nil {
		o.Logger = log.Noop
	}

	statestore := mock.NewStateStore()
	if o.Addressbook == nil {
		o.Addressbook = address.New(statestore)
	}

	if o.PrivateKey == nil {
		libp2pKey, err := crypto.GenerateSecp256k1Key()
		if err != nil {
			t.Fatal(err)
		}

		o.PrivateKey = libp2pKey
	}

	ctx, cancel := context.WithCancel(context.Background())

	if o.lightNodes == nil {
		o.lightNodes = lightnode.NewContainer(overlay)
	}
	opts := o.libp2pOpts
	opts.Nonce = nonce

	s, err = libp2p.New(ctx, crypto.NewDefaultSigner(clusterKey), networkID, overlay, addr, o.Addressbook, statestore, o.lightNodes, o.Logger, nil, opts)
	if err != nil {
		t.Fatal(err)
	}

	if o.notifier != nil {
		s.SetPickyNotifier(o.notifier)
	}

	_ = s.Ready()

	t.Cleanup(func() {
		cancel()
		s.Close()
	})
	return s, overlay
}

// expectPeers validates that peers with addresses are connected.
func expectPeers(t *testing.T, s *libp2p.Service, addrs ...cluster.Address) {
	t.Helper()

	peers := s.Peers()

	if len(peers) != len(addrs) {
		t.Fatalf("got peers %v, want %v", len(peers), len(addrs))
	}

	sort.Slice(addrs, func(i, j int) bool {
		return bytes.Compare(addrs[i].Bytes(), addrs[j].Bytes()) == -1
	})
	sort.Slice(peers, func(i, j int) bool {
		return bytes.Compare(peers[i].Address.Bytes(), peers[j].Address.Bytes()) == -1
	})

	for i, got := range peers {
		want := addrs[i]
		if !got.Address.Equal(want) {
			t.Errorf("got %v peer %s, want %s", i, got.Address, want)
		}
	}
}

// expectPeersEventually validates that peers with addresses are connected with
// retries. It is supposed to be used to validate asynchronous connecting on the
// peer that is connected to.
func expectPeersEventually(t *testing.T, s *libp2p.Service, addrs ...cluster.Address) {
	t.Helper()

	var peers []p2p.Peer
	for i := 0; i < 100; i++ {
		peers = s.Peers()
		if len(peers) == len(addrs) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if len(peers) != len(addrs) {
		t.Fatalf("got peers %v, want %v", len(peers), len(addrs))
	}

	sort.Slice(addrs, func(i, j int) bool {
		return bytes.Compare(addrs[i].Bytes(), addrs[j].Bytes()) == -1
	})
	sort.Slice(peers, func(i, j int) bool {
		return bytes.Compare(peers[i].Address.Bytes(), peers[j].Address.Bytes()) == -1
	})

	for i, got := range peers {
		want := addrs[i]
		if !got.Address.Equal(want) {
			t.Errorf("got %v peer %s, want %s", i, got.Address, want)
		}
	}
}

func serviceUnderlayAddress(t *testing.T, s *libp2p.Service) multiaddr.Multiaddr {
	t.Helper()

	addrs, err := s.Addresses()
	if err != nil {
		t.Fatal(err)
	}
	return addrs[0]
}
