package libp2p

import (
	"github.com/prometheus/client_golang/prometheus"
	m "github.com/redesblock/hop/core/metrics"
)

type metrics struct {
	// all metrics fields must be exported
	// to be able to return them by Metrics()
	// using reflection
	CreatedConnectionCount prometheus.Counter
	HandledConnectionCount prometheus.Counter
	CreatedStreamCount     prometheus.Counter
	HandledStreamCount     prometheus.Counter
}

func newMetrics() (m metrics) {
	return metrics{
		CreatedConnectionCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "libp2p_created_connection_count",
			Help: "Number of initiated outgoing libp2p connections.",
		}),
		HandledConnectionCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "libp2p_handled_connection_count",
			Help: "Number of handled incoming libp2p connections.",
		}),
		CreatedStreamCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "libp2p_created_stream_count",
			Help: "Number of initiated outgoing libp2p streams.",
		}),
		HandledStreamCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "libp2p_handled_stream_count",
			Help: "Number of handled incoming libp2p streams.",
		}),
	}
}

func (s *Service) Metrics() []prometheus.Collector {
	return m.PrometheusCollectorsFromFields(s.metrics)
}
