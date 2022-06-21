// Package debugapi exposes the debug API used to
// control and analyze low-level and runtime
// features and functionalities of hop.
package debugapi

import (
	"crypto/ecdsa"
	"net/http"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redesblock/hop/core/accounting"
	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/p2p"
	"github.com/redesblock/hop/core/pingpong"
	"github.com/redesblock/hop/core/postage"
	"github.com/redesblock/hop/core/settlement"
	"github.com/redesblock/hop/core/settlement/swap"
	"github.com/redesblock/hop/core/settlement/swap/chequebook"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/tags"
	"github.com/redesblock/hop/core/topology"
	"github.com/redesblock/hop/core/topology/lightnode"
	"github.com/redesblock/hop/core/tracing"
)

// Service implements http.Handler interface to be used in HTTP server.
type Service struct {
	overlay            swarm.Address
	publicKey          ecdsa.PublicKey
	pssPublicKey       ecdsa.PublicKey
	ethereumAddress    common.Address
	p2p                p2p.DebugService
	pingpong           pingpong.Interface
	topologyDriver     topology.Driver
	storer             storage.Storer
	logger             logging.Logger
	tracer             *tracing.Tracer
	tags               *tags.Tags
	accounting         accounting.Interface
	settlement         settlement.Interface
	chequebookEnabled  bool
	chequebook         chequebook.Service
	swap               swap.ApiInterface
	batchStore         postage.Storer
	corsAllowedOrigins []string
	metricsRegistry    *prometheus.Registry
	lightNodes         *lightnode.Container
	// handler is changed in the Configure method
	handler   http.Handler
	handlerMu sync.RWMutex
}

// New creates a new Debug API Service with only basic routers enabled in order
// to expose /addresses, /health endpoints, Go metrics and pprof. It is useful to expose
// these endpoints before all dependencies are configured and injected to have
// access to basic debugging tools and /health endpoint.
func New(overlay swarm.Address, publicKey, pssPublicKey ecdsa.PublicKey, ethereumAddress common.Address, logger logging.Logger, tracer *tracing.Tracer, corsAllowedOrigins []string) *Service {
	s := new(Service)
	s.overlay = overlay
	s.publicKey = publicKey
	s.pssPublicKey = pssPublicKey
	s.ethereumAddress = ethereumAddress
	s.logger = logger
	s.tracer = tracer
	s.corsAllowedOrigins = corsAllowedOrigins
	s.metricsRegistry = newMetricsRegistry()

	s.setRouter(s.newBasicRouter())

	return s
}

// Configure injects required dependencies and configuration parameters and
// constructs HTTP routes that depend on them. It is intended and safe to call
// this method only once.
func (s *Service) Configure(p2p p2p.DebugService, pingpong pingpong.Interface, topologyDriver topology.Driver, lightNodes *lightnode.Container, storer storage.Storer, tags *tags.Tags, accounting accounting.Interface, settlement settlement.Interface, chequebookEnabled bool, swap swap.ApiInterface, chequebook chequebook.Service, batchStore postage.Storer) {
	s.p2p = p2p
	s.pingpong = pingpong
	s.topologyDriver = topologyDriver
	s.storer = storer
	s.tags = tags
	s.accounting = accounting
	s.settlement = settlement
	s.chequebookEnabled = chequebookEnabled
	s.chequebook = chequebook
	s.swap = swap
	s.lightNodes = lightNodes
	s.batchStore = batchStore

	s.setRouter(s.newRouter())
}

// ServeHTTP implements http.Handler interface.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// protect handler as it is changed by the Configure method
	s.handlerMu.RLock()
	h := s.handler
	s.handlerMu.RUnlock()

	h.ServeHTTP(w, r)
}
