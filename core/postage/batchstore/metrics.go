package batchstore

import (
	"github.com/prometheus/client_golang/prometheus"
	m "github.com/redesblock/hop/core/metrics"
)

type metrics struct {
	AvailableCapacity prometheus.Gauge
	Inner             prometheus.Gauge
	Outer             prometheus.Gauge
	Radius            prometheus.Gauge
}

func newMetrics() metrics {
	subsystem := "batchstore"

	return metrics{
		AvailableCapacity: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "available_capacity",
			Help:      "Available capacity observed by the batchstore.",
		}),
		Inner: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "inner",
			Help:      "Inner storage tier value observed by the batchstore.",
		}),
		Outer: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "outer",
			Help:      "Outer storage tier value observed by the batchstore.",
		}),
		Radius: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "radius",
			Help:      "Radius of responsibility observed by the batchstore.",
		}),
	}
}

func (s *store) Metrics() []prometheus.Collector {
	return m.PrometheusCollectorsFromFields(s.metrics)
}
