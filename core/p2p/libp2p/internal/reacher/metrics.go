package reacher

import (
	"github.com/prometheus/client_golang/prometheus"
	m "github.com/redesblock/mop/core/metrics"
)

type metrics struct {
	Pings    prometheus.CounterVec
	PingTime prometheus.HistogramVec
}

func newMetrics() metrics {
	subsystem := "reacher"

	return metrics{
		Pings: *prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "pings",
			Help:      "Ping counter.",
		}, []string{"status"}),
		PingTime: *prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "ping_timer",
			Help:      "Ping timer.",
		}, []string{"status"}),
	}
}

func (s *reacher) Metrics() []prometheus.Collector {
	return m.PrometheusCollectorsFromFields(s.metrics)
}
