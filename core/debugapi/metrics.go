package debugapi

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redesblock/hop/core/metrics"
)

func newMetricsRegistry() (r *prometheus.Registry) {
	r = prometheus.NewRegistry()

	// register standard metrics
	r.MustRegister(
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{
			Namespace: metrics.Namespace,
		}),
		prometheus.NewGoCollector(),
		prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   metrics.Namespace,
			Name:        "info",
			Help:        "Hop information.",
			ConstLabels: prometheus.Labels{},
		}),
	)

	return r
}

func (s *server) MustRegisterMetrics(cs ...prometheus.Collector) {
	s.metricsRegistry.MustRegister(cs...)
}
