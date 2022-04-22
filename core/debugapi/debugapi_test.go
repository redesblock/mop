package debugapi_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/multiformats/go-multiaddr"
	"github.com/redesblock/hop/core/addressbook"
	"github.com/redesblock/hop/core/api"
	"github.com/redesblock/hop/core/debugapi"
	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/p2p"
	mockstore "github.com/redesblock/hop/core/statestore/mock"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/topology"
	"github.com/redesblock/hop/core/topology/mock"
	"resenje.org/web"
)

type testServerOptions struct {
	Overlay      swarm.Address
	P2P          p2p.Service
	Storer       storage.Storer
	TopologyOpts []mock.Option
}

type testServer struct {
	Client         *http.Client
	Addressbook    addressbook.GetPutter
	TopologyDriver topology.Driver
	Cleanup        func()
}

func newTestServer(t *testing.T, o testServerOptions) *testServer {
	statestore := mockstore.NewStateStore()
	addrbook := addressbook.New(statestore)
	topologyDriver := mock.NewTopologyDriver(o.TopologyOpts...)

	s := debugapi.New(debugapi.Options{
		Overlay:        o.Overlay,
		P2P:            o.P2P,
		Logger:         logging.New(ioutil.Discard, 0),
		Addressbook:    addrbook,
		TopologyDriver: topologyDriver,
		Storer:         o.Storer,
	})
	ts := httptest.NewServer(s)
	cleanup := ts.Close

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
		Client:         client,
		Addressbook:    addrbook,
		TopologyDriver: topologyDriver,
		Cleanup:        cleanup,
	}
}

func newHOPTestServer(t *testing.T, o testServerOptions) (client *http.Client, cleanup func()) {
	s := api.New(api.Options{
		Storer: o.Storer,
		Logger: logging.New(ioutil.Discard, 0),
	})
	ts := httptest.NewServer(s)
	cleanup = ts.Close

	client = &http.Client{
		Transport: web.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
			u, err := url.Parse(ts.URL + r.URL.String())
			if err != nil {
				return nil, err
			}
			r.URL = u
			return ts.Client().Transport.RoundTrip(r)
		}),
	}
	return client, cleanup
}

func mustMultiaddr(t *testing.T, s string) multiaddr.Multiaddr {
	t.Helper()

	a, err := multiaddr.NewMultiaddr(s)
	if err != nil {
		t.Fatal(err)
	}
	return a
}
