package debugapi_test

import (
	"crypto/ecdsa"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiformats/go-multiaddr"
	accountingmock "github.com/redesblock/hop/core/accounting/mock"
	"github.com/redesblock/hop/core/debugapi"
	"github.com/redesblock/hop/core/logging"
	p2pmock "github.com/redesblock/hop/core/p2p/mock"
	"github.com/redesblock/hop/core/pingpong"
	"github.com/redesblock/hop/core/resolver"
	settlementmock "github.com/redesblock/hop/core/settlement/pseudosettle/mock"
	chequebookmock "github.com/redesblock/hop/core/settlement/swap/chequebook/mock"
	swapmock "github.com/redesblock/hop/core/settlement/swap/mock"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/tags"
	topologymock "github.com/redesblock/hop/core/topology/mock"
	"resenje.org/web"
)

type testServerOptions struct {
	Overlay         swarm.Address
	PublicKey       ecdsa.PublicKey
	PSSPublicKey    ecdsa.PublicKey
	EthereumAddress common.Address
	P2P             *p2pmock.Service
	Pingpong        pingpong.Interface
	Storer          storage.Storer
	Resolver        resolver.Interface
	TopologyOpts    []topologymock.Option
	Tags            *tags.Tags
	AccountingOpts  []accountingmock.Option
	SettlementOpts  []settlementmock.Option
	ChequebookOpts  []chequebookmock.Option
	SwapOpts        []swapmock.Option
}

type testServer struct {
	Client  *http.Client
	P2PMock *p2pmock.Service
}

func newTestServer(t *testing.T, o testServerOptions) *testServer {
	topologyDriver := topologymock.NewTopologyDriver(o.TopologyOpts...)
	acc := accountingmock.NewAccounting(o.AccountingOpts...)
	settlement := settlementmock.NewSettlement(o.SettlementOpts...)
	chequebook := chequebookmock.NewChequebook(o.ChequebookOpts...)
	swapserv := swapmock.NewApiInterface(o.SwapOpts...)
	s := debugapi.New(o.Overlay, o.PublicKey, o.PSSPublicKey, o.EthereumAddress, o.P2P, o.Pingpong, topologyDriver, o.Storer, logging.New(ioutil.Discard, 0), nil, o.Tags, acc, settlement, true, swapserv, chequebook)
	ts := httptest.NewServer(s)
	t.Cleanup(ts.Close)

	client := &http.Client{
		Transport: web.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
			u, err := url.Parse(ts.URL + r.URL.String())
			if err != nil {
				return nil, err
			}
			r.URL = u
			return ts.Client().Transport.RoundTrip(r)
		}),
	}
	return &testServer{
		Client:  client,
		P2PMock: o.P2P,
	}
}

func mustMultiaddr(t *testing.T, s string) multiaddr.Multiaddr {
	t.Helper()

	a, err := multiaddr.NewMultiaddr(s)
	if err != nil {
		t.Fatal(err)
	}
	return a
}
