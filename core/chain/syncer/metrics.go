package syncer

import (
	"github.com/prometheus/client_golang/prometheus"
	m "github.com/redesblock/mop/core/metrics"
)

type metrics struct {
	SyncedPeers      prometheus.Counter
	UnsyncedPeers    prometheus.Counter
	PeerErrors       prometheus.Counter
	InvalidProofs    prometheus.Counter
	TotalTimeWaiting prometheus.Counter
	PositiveProofs   prometheus.Gauge
}

func newMetrics() metrics {
	subsystem := "syncer"

	return metrics{
		SyncedPeers: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "synced_peers_count",
			Help:      "total number of synced peers. duplicate increments are expected for the same peer",
		}),
		UnsyncedPeers: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "unsynced_peers_count",
			Help:      "total number of peers marked as unsynced",
		}),
		PeerErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "peer_error_count",
			Help:      "total number of errors we've received when asking for a proof from a peer. duplicates increments are expected",
		}),
		InvalidProofs: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "invalid_proof_count",
			Help:      "total number of invalid proofs we've received. duplicate increments are expected",
		}),
		TotalTimeWaiting: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "total_time_waiting",
			Help:      "total time spent waiting for proofs (not per peer but rather per cycle)",
		}),
		PositiveProofs: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "positive_proofs",
			Help:      "percentage of positive proofs logged in every round of chain-chainsync challenge-response iterations",
		}),
	}
}

// Metrics returns the prometheus Collector for the bookkeeper service.
func (c *ChainSyncer) Metrics() []prometheus.Collector {
	return m.PrometheusCollectorsFromFields(c.metrics)
}
