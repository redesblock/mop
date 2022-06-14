package debugapi_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/multiformats/go-multiaddr"
	accountingmock "github.com/redesblock/hop/core/accounting/mock"
	"github.com/redesblock/hop/core/api"
	"github.com/redesblock/hop/core/debugapi"
	"github.com/redesblock/hop/core/logging"
	p2pmock "github.com/redesblock/hop/core/p2p/mock"
	"github.com/redesblock/hop/core/pingpong"
	"github.com/redesblock/hop/core/resolver"
	resolverMock "github.com/redesblock/hop/core/resolver/mock"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/tags"
	topologymock "github.com/redesblock/hop/core/topology/mock"
	"resenje.org/web"
)

type testServerOptions struct {
	Overlay        swarm.Address
	P2P            *p2pmock.Service
	Pingpong       pingpong.Interface
	Storer         storage.Storer
	Resolver       resolver.Interface
	TopologyOpts   []topologymock.Option
	Tags           *tags.Tags
	AccountingOpts []accountingmock.Option
}

type testServer struct {
	Client  *http.Client
	P2PMock *p2pmock.Service
}

func newTestServer(t *testing.T, o testServerOptions) *testServer {
	topologyDriver := topologymock.NewTopologyDriver(o.TopologyOpts...)
	acc := accountingmock.NewAccounting(o.AccountingOpts...)

	s := debugapi.New(o.Overlay, o.P2P, o.Pingpong, topologyDriver, o.Storer, logging.New(ioutil.Discard, 0), nil, o.Tags, acc)
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

func newHOPTestServer(t *testing.T, o testServerOptions) *http.Client {
	if o.Resolver == nil {
		o.Resolver = resolverMock.NewResolver()
	}
	s := api.New(o.Tags, o.Storer, o.Resolver, nil, logging.New(ioutil.Discard, 0), nil)
	ts := httptest.NewServer(s)
	t.Cleanup(ts.Close)

	return &http.Client{
		Transport: web.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
			u, err := url.Parse(ts.URL + r.URL.String())
			if err != nil {
				return nil, err
			}
			r.URL = u
			return ts.Client().Transport.RoundTrip(r)
		}),
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
