package api

import (
	"net/http"

	"github.com/redesblock/hop/core/logging"
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
	Tags               *tags.Tags
	Storer             storage.Storer
	CORSAllowedOrigins []string
	Logger             logging.Logger
	Tracer             *tracing.Tracer
	http.Handler
	metrics metrics
}

const (
	// TargetsRecoveryHeader defines the Header for Recovery targets in Global Pinning
	TargetsRecoveryHeader = "swarm-recovery-targets"
)

func New(tags *tags.Tags, storer storage.Storer, corsAllowedOrigins []string, logger logging.Logger, tracer *tracing.Tracer) Service {
	s := &server{
		Tags:               tags,
		Storer:             storer,
		CORSAllowedOrigins: corsAllowedOrigins,
		Logger:             logger,
		Tracer:             tracer,
		metrics:            newMetrics(),
	}

	s.setupRouting()

	return s
}
