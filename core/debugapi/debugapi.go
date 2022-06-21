// Package debugapi exposes the debug API used to
// control and analyze low-level and runtime
// features and functionalities of hop.
package debugapi

import (
	"crypto/ecdsa"
	"net/http"
	"sync"
	"unicode/utf8"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redesblock/hop/core/accounting"
	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/p2p"
	"github.com/redesblock/hop/core/pingpong"
	"github.com/redesblock/hop/core/settlement"
	"github.com/redesblock/hop/core/settlement/swap"
	"github.com/redesblock/hop/core/settlement/swap/chequebook"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/tags"
	"github.com/redesblock/hop/core/topology"
	"github.com/redesblock/hop/core/tracing"
)

type Service interface {
	http.Handler
	Configure(p2p p2p.DebugService, pingpong pingpong.Interface, topologyDriver topology.Driver, storer storage.Storer, tags *tags.Tags, accounting accounting.Interface, settlement settlement.Interface, chequebookEnabled bool, swap swap.ApiInterface, chequebook chequebook.Service)
	MustRegisterMetrics(cs ...prometheus.Collector)
}

type server struct {
	Overlay            swarm.Address
	PublicKey          ecdsa.PublicKey
	PSSPublicKey       ecdsa.PublicKey
	EthereumAddress    common.Address
	P2P                p2p.DebugService
	Pingpong           pingpong.Interface
	TopologyDriver     topology.Driver
	Storer             storage.Storer
	Logger             logging.Logger
	Tracer             *tracing.Tracer
	Tags               *tags.Tags
	Accounting         accounting.Interface
	Settlement         settlement.Interface
	ChequebookEnabled  bool
	Chequebook         chequebook.Service
	Swap               swap.ApiInterface
	CORSAllowedOrigins []string
	metricsRegistry    *prometheus.Registry
	// handler is changed in the Configure method
	handler   http.Handler
	handlerMu sync.RWMutex
}

// checkOrigin returns true if the origin is not set or is equal to the request host.
func (s *server) checkOrigin(r *http.Request) bool {
	origin := r.Header["Origin"]
	if len(origin) == 0 {
		return true
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	hosts := append(s.CORSAllowedOrigins, scheme+"://"+r.Host)
	for _, v := range hosts {
		if equalASCIIFold(origin[0], v) || v == "*" {
			return true
		}
	}

	return false
}

// equalASCIIFold returns true if s is equal to t with ASCII case folding as
// defined in RFC 4790.
func equalASCIIFold(s, t string) bool {
	for s != "" && t != "" {
		sr, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		tr, size := utf8.DecodeRuneInString(t)
		t = t[size:]
		if sr == tr {
			continue
		}
		if 'A' <= sr && sr <= 'Z' {
			sr = sr + 'a' - 'A'
		}
		if 'A' <= tr && tr <= 'Z' {
			tr = tr + 'a' - 'A'
		}
		if sr != tr {
			return false
		}
	}
	return s == t
}

// New creates a new Debug API Service with only basic routers enabled in order
// to expose /addresses, /health endpoints, Go metrics and pprof. It is useful to expose
// these endpoints before all dependencies are configured and injected to have
// access to basic debugging tools and /health endpoint.
func New(overlay swarm.Address, publicKey, pssPublicKey ecdsa.PublicKey, ethereumAddress common.Address, logger logging.Logger, tracer *tracing.Tracer, corsAllowedOrigins []string) Service {
	s := new(server)
	s.Overlay = overlay
	s.PublicKey = publicKey
	s.PSSPublicKey = pssPublicKey
	s.EthereumAddress = ethereumAddress
	s.Logger = logger
	s.Tracer = tracer
	s.CORSAllowedOrigins = corsAllowedOrigins
	s.metricsRegistry = newMetricsRegistry()

	s.setRouter(s.newBasicRouter())

	return s
}

// Configure injects required dependencies and configuration parameters and
// constructs HTTP routes that depend on them. It is intended and safe to call
// this method only once.
func (s *server) Configure(p2p p2p.DebugService, pingpong pingpong.Interface, topologyDriver topology.Driver, storer storage.Storer, tags *tags.Tags, accounting accounting.Interface, settlement settlement.Interface, chequebookEnabled bool, swap swap.ApiInterface, chequebook chequebook.Service) {
	s.P2P = p2p
	s.Pingpong = pingpong
	s.TopologyDriver = topologyDriver
	s.Storer = storer
	s.Tags = tags
	s.Accounting = accounting
	s.Settlement = settlement
	s.ChequebookEnabled = chequebookEnabled
	s.Chequebook = chequebook
	s.Swap = swap

	s.setRouter(s.newRouter())
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// protect handler as it is changed by the Configure method
	s.handlerMu.RLock()
	h := s.handler
	s.handlerMu.RUnlock()

	h.ServeHTTP(w, r)
}
