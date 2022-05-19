package debugapi

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/p2p"
	"github.com/redesblock/hop/core/pingpong"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/tags"
	"github.com/redesblock/hop/core/topology"
	"github.com/redesblock/hop/core/tracing"
)

type Service interface {
	http.Handler
	MustRegisterMetrics(cs ...prometheus.Collector)
}

type server struct {
	Options
	http.Handler

	metricsRegistry *prometheus.Registry
}

type Options struct {
	Overlay        swarm.Address
	P2P            p2p.Service
	Pingpong       pingpong.Interface
	TopologyDriver topology.PeerAdder
	Storer         storage.Storer
	Logger         logging.Logger
	Tracer         *tracing.Tracer
	Tags           *tags.Tags
}

func New(o Options) Service {
	s := &server{
		Options:         o,
		metricsRegistry: newMetricsRegistry(),
	}

	s.setupRouting()

	return s
}
