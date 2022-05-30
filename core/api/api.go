package api

import (
	"net/http"

	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/manifest"
	m "github.com/redesblock/hop/core/metrics"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/tags"
	"github.com/redesblock/hop/core/tracing"
)

type Service interface {
	http.Handler
	m.Collector
}

type server struct {
	Options
	http.Handler
	metrics metrics
}

type Options struct {
	Tags               *tags.Tags
	Storer             storage.Storer
	ManifestParser     manifest.Parser
	CORSAllowedOrigins []string
	Logger             logging.Logger
	Tracer             *tracing.Tracer
}

const (
	// TargetsRecoveryHeader defines the Header for Recovery targets in Global Pinning
	TargetsRecoveryHeader = "swarm-recovery-targets"
)

func New(o Options) Service {
	s := &server{
		Options: o,
		metrics: newMetrics(),
	}

	s.setupRouting()

	return s
}
