package api_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/redesblock/hop/core/api"
	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/pingpong"
	"resenje.org/web"
)

type testServerOptions struct {
	Pingpong pingpong.Interface
}

func newTestServer(t *testing.T, o testServerOptions) (client *http.Client, cleanup func()) {
	s := api.New(api.Options{
		Pingpong: o.Pingpong,
		Logger:   logging.New(ioutil.Discard, 0),
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
