package debugapi_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/multiformats/go-multiaddr"
	"github.com/redesblock/hop/core/api"
	"github.com/redesblock/hop/core/debugapi"
	"github.com/redesblock/hop/core/logging"
	mockp2p "github.com/redesblock/hop/core/p2p/mock"
	"github.com/redesblock/hop/core/pingpong"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/tags"
	"github.com/redesblock/hop/core/topology/mock"
	"resenje.org/web"
)

type testServerOptions struct {
	Overlay      swarm.Address
	P2P          *mockp2p.Service
	Pingpong     pingpong.Interface
	Storer       storage.Storer
	TopologyOpts []mock.Option
	Tags         *tags.Tags
}

type testServer struct {
	Client  *http.Client
	P2PMock *mockp2p.Service
}

func newTestServer(t *testing.T, o testServerOptions) *testServer {
	topologyDriver := mock.NewTopologyDriver(o.TopologyOpts...)

	s := debugapi.New(debugapi.Options{
		Overlay:        o.Overlay,
		P2P:            o.P2P,
		Pingpong:       o.Pingpong,
		Tags:           o.Tags,
		Logger:         logging.New(ioutil.Discard, 0),
		Storer:         o.Storer,
		TopologyDriver: topologyDriver,
	})
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

func newHopTestServer(t *testing.T, o testServerOptions) *http.Client {
	s := api.New(api.Options{
		Storer: o.Storer,
		Tags:   o.Tags,
		Logger: logging.New(ioutil.Discard, 0),
	})
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
