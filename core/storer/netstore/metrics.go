package netstore

import (
	"github.com/prometheus/client_golang/prometheus"
	m "github.com/redesblock/mop/core/metrics"
)

type metrics struct {
	LocalChunksCounter        prometheus.Counter
	InvalidLocalChunksCounter prometheus.Counter
	RetrievedChunksCounter    prometheus.Counter
	RetrievedMemChunksCounter prometheus.Counter
}

func newMetrics() metrics {
	subsystem := "netstore"

	return metrics{
		LocalChunksCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "local_chunks_retrieved",
			Help:      "Total no. of chunks retrieved locally.",
		}),
		InvalidLocalChunksCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "invalid_local_chunks_retrieved",
			Help:      "Total no. of chunks retrieved locally that are invalid.",
		}),
		RetrievedChunksCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "chunks_retrieved_from_network",
			Help:      "Total no. of chunks retrieved from network.",
		}),
		RetrievedMemChunksCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "chunks_retrieved_from_memory",
			Help:      "Total no. of chunks retrieved from memory.",
		}),
	}
}

func (s *store) Metrics() []prometheus.Collector {
	return m.PrometheusCollectorsFromFields(s.metrics)
}
