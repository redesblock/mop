package debugapi

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redesblock/hop/core/accounting"
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
	Overlay        swarm.Address
	P2P            p2p.DebugService
	Pingpong       pingpong.Interface
	TopologyDriver topology.Driver
	Storer         storage.Storer
	Logger         logging.Logger
	Tracer         *tracing.Tracer
	Tags           *tags.Tags
	Accounting     accounting.Interface
	http.Handler

	metricsRegistry *prometheus.Registry
}

func New(overlay swarm.Address, p2p p2p.DebugService, pingpong pingpong.Interface, topologyDriver topology.Driver, storer storage.Storer, logger logging.Logger, tracer *tracing.Tracer, tags *tags.Tags, accounting accounting.Interface) Service {
	s := &server{
		Overlay:         overlay,
		P2P:             p2p,
		Pingpong:        pingpong,
		TopologyDriver:  topologyDriver,
		Storer:          storer,
		Logger:          logger,
		Tracer:          tracer,
		Tags:            tags,
		Accounting:      accounting,
		metricsRegistry: newMetricsRegistry(),
	}

	s.setupRouting()

	return s
}
