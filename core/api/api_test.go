package api_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/redesblock/hop/core/api"
	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/manifest"
	"github.com/redesblock/hop/core/pingpong"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/tags"
	"resenje.org/web"
)

type testServerOptions struct {
	Pingpong       pingpong.Interface
	Storer         storage.Storer
	ManifestParser manifest.Parser
	Tags           *tags.Tags
	Logger         logging.Logger
}

func newTestServer(t *testing.T, o testServerOptions) *http.Client {
	if o.Logger == nil {
		o.Logger = logging.New(ioutil.Discard, 0)
	}
	s := api.New(api.Options{
		Tags:           o.Tags,
		Storer:         o.Storer,
		ManifestParser: o.ManifestParser,
		Logger:         o.Logger,
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
