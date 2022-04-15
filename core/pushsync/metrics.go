package pushsync

import (
	"github.com/prometheus/client_golang/prometheus"
	m "github.com/redesblock/hop/core/metrics"
)

type metrics struct {
	// all metrics fields must be exported
	// to be able to return them by Metrics()
	// using reflection

	SendChunkCounter      prometheus.Counter
	SendChunkTimer        prometheus.Counter
	SendChunkErrorCounter prometheus.Counter
	MarkAndSweepTimer     prometheus.Counter

	ChunksInBatch prometheus.Gauge
}

func newMetrics() metrics {
	subsystem := "pushsync"

	return metrics{
		SendChunkCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "send_chunk",
			Help:      "Total chunks to be sent.",
		}),
		SendChunkTimer: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "send_chunk_time_taken",
			Help:      "Total time taken to send a chunk.",
		}),
		SendChunkErrorCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "send_chunk_error",
			Help:      "Total no of time error received while sending chunk.",
		}),
		MarkAndSweepTimer: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "mark_and_sweep_time",
			Help:      "Total time spent in mark and sweep.",
		}),

		ChunksInBatch: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "chunks_in_batch",
			Help:      "Chunks in batch at a given time.",
		}),
	}
}

func (s *PushSync) Metrics() []prometheus.Collector {
	return m.PrometheusCollectorsFromFields(s.metrics)
}
