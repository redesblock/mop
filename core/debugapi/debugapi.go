// Package debugapi exposes the debug API used to
// control and analyze low-level and runtime
// features and functionalities of hop.
package debugapi

import (
	"crypto/ecdsa"
	"net/http"

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
	http.Handler
}

func New(overlay swarm.Address, publicKey, pssPublicKey ecdsa.PublicKey, ethereumAddress common.Address, p2p p2p.DebugService, pingpong pingpong.Interface, topologyDriver topology.Driver, storer storage.Storer, logger logging.Logger, tracer *tracing.Tracer, tags *tags.Tags, accounting accounting.Interface, settlement settlement.Interface, chequebookEnabled bool, swap swap.ApiInterface, chequebook chequebook.Service) Service {
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
