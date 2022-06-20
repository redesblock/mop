// Package debugapi exposes the debug API used to
// control and analyze low-level and runtime
// features and functionalities of hop.
package debugapi

import (
	"crypto/ecdsa"
	"net/http"
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
	MustRegisterMetrics(cs ...prometheus.Collector)
}

type server struct {
	Overlay           swarm.Address
	PublicKey         ecdsa.PublicKey
	PSSPublicKey      ecdsa.PublicKey
	EthereumAddress   common.Address
	P2P               p2p.DebugService
	Pingpong          pingpong.Interface
	TopologyDriver    topology.Driver
	Storer            storage.Storer
	Logger            logging.Logger
	Tracer            *tracing.Tracer
	Tags              *tags.Tags
	Accounting        accounting.Interface
	Settlement        settlement.Interface
	ChequebookEnabled bool
	Chequebook        chequebook.Service
	Swap              swap.ApiInterface
	metricsRegistry   *prometheus.Registry
	Options
	http.Handler
}

type Options struct {
	CORSAllowedOrigins []string
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

func New(overlay swarm.Address, publicKey, pssPublicKey ecdsa.PublicKey, ethereumAddress common.Address, p2p p2p.DebugService, pingpong pingpong.Interface, topologyDriver topology.Driver, storer storage.Storer, logger logging.Logger, tracer *tracing.Tracer, tags *tags.Tags, accounting accounting.Interface, settlement settlement.Interface, chequebookEnabled bool, swap swap.ApiInterface, chequebook chequebook.Service, o Options) Service {
	s := &server{
		Overlay:           overlay,
		PublicKey:         publicKey,
		PSSPublicKey:      pssPublicKey,
		EthereumAddress:   ethereumAddress,
		P2P:               p2p,
		Pingpong:          pingpong,
		TopologyDriver:    topologyDriver,
		Storer:            storer,
		Logger:            logger,
		Tracer:            tracer,
		Tags:              tags,
		Accounting:        accounting,
		Settlement:        settlement,
		metricsRegistry:   newMetricsRegistry(),
		ChequebookEnabled: chequebookEnabled,
		Chequebook:        chequebook,
		Swap:              swap,
	}

	s.setupRouting()

	return s
}
