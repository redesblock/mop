package puller

import (
	"github.com/prometheus/client_golang/prometheus"
	m "github.com/redesblock/mop/core/metrics"
)

type metrics struct {
	HistWorkerIterCounter prometheus.Counter // counts the number of historical syncing iterations
	HistWorkerDoneCounter prometheus.Counter // count number of finished historical syncing jobs
	HistWorkerErrCounter  prometheus.Counter // count number of errors
	LiveWorkerIterCounter prometheus.Counter // counts the number of live syncing iterations
	LiveWorkerErrCounter  prometheus.Counter // count number of errors
	MaxUintErrCounter     prometheus.Counter // how many times we got maxuint as topmost
}

func newMetrics() metrics {
	subsystem := "puller"

	return metrics{
		HistWorkerIterCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "hist_worker_iterations",
			Help:      "Total history worker iterations.",
		}),
		HistWorkerDoneCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "hist_worker_done",
			Help:      "Total history worker jobs done.",
		}),
		HistWorkerErrCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "hist_worker_errors",
			Help:      "Total history worker errors.",
		}),
		LiveWorkerIterCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "live_worker_iterations",
			Help:      "Total live worker iterations.",
		}),
		LiveWorkerErrCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "live_worker_errors",
			Help:      "Total live worker errors.",
		}),
		MaxUintErrCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "max_uint_errors",
			Help:      "Total max uint errors.",
		}),
	}
}

func (p *Puller) Metrics() []prometheus.Collector {
	return m.PrometheusCollectorsFromFields(p.metrics)
}
