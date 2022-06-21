package pss

import (
	"github.com/prometheus/client_golang/prometheus"
	m "github.com/redesblock/hop/core/metrics"
)

type metrics struct {
	TotalMessagesSentCounter prometheus.Counter
	MessageMiningDuration    prometheus.Gauge
}

func newMetrics() metrics {
	subsystem := "pss"

	return metrics{
		TotalMessagesSentCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "total_message_sent",
			Help:      "Total messages sent.",
		}),
		MessageMiningDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "mining_duration",
			Help:      "Time duration to mine a message.",
		}),
	}
}

func (s *pss) Metrics() []prometheus.Collector {
	return m.PrometheusCollectorsFromFields(s.metrics)
}
