package debugapi_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/redesblock/hop/core/crypto"
	"github.com/redesblock/hop/core/debugapi"
	"github.com/redesblock/hop/core/hop"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/jsonhttp/jsonhttptest"
	"github.com/redesblock/hop/core/p2p"
	"github.com/redesblock/hop/core/p2p/mock"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
	topmock "github.com/redesblock/hop/core/topology/mock"
)

func TestConnect(t *testing.T) {
	underlay := "/ip4/127.0.0.1/tcp/7070/p2p/16Uiu2HAkx8ULY8cTXhdVAcMmLcH9AsTKz6uBQ7DPLKRjMLgBVYkS"
	errorUnderlay := "/ip4/127.0.0.1/tcp/7070/p2p/16Uiu2HAkw88cjH2orYrB6fDui4eUNdmgkwnDM8W681UbfsPgM9QY"
	testErr := errors.New("test error")

	privateKey, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatal(err)
	}

	overlay := crypto.NewOverlayAddress(privateKey.PublicKey, 0)
	underlama, err := ma.NewMultiaddr(underlay)
	if err != nil {
		t.Fatal(err)
	}

	hopAddress, err := hop.NewAddress(crypto.NewDefaultSigner(privateKey), underlama, overlay, 0)
	if err != nil {
		t.Fatal(err)
	}

	testServer := newTestServer(t, testServerOptions{
		P2P: mock.New(mock.WithConnectFunc(func(ctx context.Context, addr ma.Multiaddr) (*hop.Address, error) {
			if addr.String() == errorUnderlay {
				return nil, testErr
			}
			return hopAddress, nil
		})),
	})

	t.Run("ok", func(t *testing.T) {
		jsonhttptest.ResponseDirect(t, testServer.Client, http.MethodPost, "/connect"+underlay, nil, http.StatusOK, debugapi.PeerConnectResponse{
			Address: overlay.String(),
		})

		hopAddr, err := testServer.Addressbook.Get(overlay)
		if err != nil && errors.Is(err, storage.ErrNotFound) && !hopAddress.Equal(&hopAddr) {
			t.Fatalf("found wrong underlay.  expected: %+v, found: %+v", hopAddress, hopAddr)
		}
	})

	t.Run("error", func(t *testing.T) {
		jsonhttptest.ResponseDirect(t, testServer.Client, http.MethodPost, "/connect"+errorUnderlay, nil, http.StatusInternalServerError, jsonhttp.StatusResponse{
			Code:    http.StatusInternalServerError,
			Message: testErr.Error(),
		})
	})

	t.Run("get method not allowed", func(t *testing.T) {
		jsonhttptest.ResponseDirect(t, testServer.Client, http.MethodGet, "/connect"+underlay, nil, http.StatusMethodNotAllowed, jsonhttp.StatusResponse{
			Code:    http.StatusMethodNotAllowed,
			Message: http.StatusText(http.StatusMethodNotAllowed),
		})
	})

	t.Run("error - add peer", func(t *testing.T) {
		disconnectCalled := false
		testServer := newTestServer(t, testServerOptions{
			P2P: mock.New(mock.WithConnectFunc(func(ctx context.Context, addr ma.Multiaddr) (*hop.Address, error) {
				if addr.String() == errorUnderlay {
					return nil, testErr
				}
				return hopAddress, nil
			}), mock.WithDisconnectFunc(func(addr swarm.Address) error {
				disconnectCalled = true
				return nil
			})),
			TopologyOpts: []topmock.Option{topmock.WithAddPeerErr(testErr)},
		})

		jsonhttptest.ResponseDirect(t, testServer.Client, http.MethodPost, "/connect"+underlay, nil, http.StatusInternalServerError, jsonhttp.StatusResponse{
			Code:    http.StatusInternalServerError,
			Message: testErr.Error(),
		})

		hopAddr, err := testServer.Addressbook.Get(overlay)
		if err != nil && errors.Is(err, storage.ErrNotFound) && !hopAddress.Equal(&hopAddr) {
			t.Fatalf("found wrong underlay.  expected: %+v, found: %+v", hopAddress, hopAddr)
		}

		if !disconnectCalled {
			t.Fatalf("disconnect not called.")
		}
	})

}

func TestDisconnect(t *testing.T) {
	address := swarm.MustParseHexAddress("ca1e9f3938cc1425c6061b96ad9eb93e134dfe8734ad490164ef20af9d1cf59c")
	unknownAdddress := swarm.MustParseHexAddress("ca1e9f3938cc1425c6061b96ad9eb93e134dfe8734ad490164ef20af9d1cf59e")
	errorAddress := swarm.MustParseHexAddress("ca1e9f3938cc1425c6061b96ad9eb93e134dfe8734ad490164ef20af9d1cf59a")
	testErr := errors.New("test error")

	testServer := newTestServer(t, testServerOptions{
		P2P: mock.New(mock.WithDisconnectFunc(func(addr swarm.Address) error {
			if addr.Equal(address) {
				return nil
			}

			if addr.Equal(errorAddress) {
				return testErr
			}

			return p2p.ErrPeerNotFound
		})),
	})

	t.Run("ok", func(t *testing.T) {
		jsonhttptest.ResponseDirect(t, testServer.Client, http.MethodDelete, "/peers/"+address.String(), nil, http.StatusOK, jsonhttp.StatusResponse{
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
		})
	})

	t.Run("unknown", func(t *testing.T) {
		jsonhttptest.ResponseDirect(t, testServer.Client, http.MethodDelete, "/peers/"+unknownAdddress.String(), nil, http.StatusBadRequest, jsonhttp.StatusResponse{
			Code:    http.StatusBadRequest,
			Message: "peer not found",
		})
	})

	t.Run("invalid peer address", func(t *testing.T) {
		jsonhttptest.ResponseDirect(t, testServer.Client, http.MethodDelete, "/peers/invalid-address", nil, http.StatusBadRequest, jsonhttp.StatusResponse{
			Code:    http.StatusBadRequest,
			Message: "invalid peer address",
		})
	})

	t.Run("error", func(t *testing.T) {
		jsonhttptest.ResponseDirect(t, testServer.Client, http.MethodDelete, "/peers/"+errorAddress.String(), nil, http.StatusInternalServerError, jsonhttp.StatusResponse{
			Code:    http.StatusInternalServerError,
			Message: testErr.Error(),
		})
	})
}

func TestPeer(t *testing.T) {
	overlay := swarm.MustParseHexAddress("ca1e9f3938cc1425c6061b96ad9eb93e134dfe8734ad490164ef20af9d1cf59c")
	testServer := newTestServer(t, testServerOptions{
		P2P: mock.New(mock.WithPeersFunc(func() []p2p.Peer {
			return []p2p.Peer{{Address: overlay}}
		})),
	})

	t.Run("ok", func(t *testing.T) {
		jsonhttptest.ResponseDirect(t, testServer.Client, http.MethodGet, "/peers", nil, http.StatusOK, debugapi.PeersResponse{
			Peers: []p2p.Peer{{Address: overlay}},
		})
	})

	t.Run("get method not allowed", func(t *testing.T) {
		jsonhttptest.ResponseDirect(t, testServer.Client, http.MethodPost, "/peers", nil, http.StatusMethodNotAllowed, jsonhttp.StatusResponse{
			Code:    http.StatusMethodNotAllowed,
			Message: http.StatusText(http.StatusMethodNotAllowed),
		})
	})
}
