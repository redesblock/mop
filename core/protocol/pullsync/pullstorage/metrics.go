package pullstorage

import (
	"github.com/prometheus/client_golang/prometheus"
	m "github.com/redesblock/mop/core/metrics"
)

type metrics struct {
	TotalSubscribePullRequests         prometheus.Counter
	TotalSubscribePullRequestsComplete prometheus.Counter
	SubscribePullsStarted              prometheus.Counter
	SubscribePullsComplete             prometheus.Counter
	SubscribePullsFailures             prometheus.Counter
}

func newMetrics() metrics {
	subsystem := "pullstorage"

	return metrics{
		TotalSubscribePullRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "subscribe_pull_requests",
			Help:      "Total subscribe pull requests.",
		}),
		TotalSubscribePullRequestsComplete: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "subscribe_pull_requests_complete",
			Help:      "Total subscribe pull requests completed.",
		}),
		SubscribePullsStarted: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "subscribe_pulls_started",
			Help:      "Total subscribe pulls started.",
		}),
		SubscribePullsComplete: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "subscribe_pulls_complete",
			Help:      "Total subscribe pulls completed.",
		}),
		SubscribePullsFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "subscribe_pulls_failures",
			Help:      "Total subscribe pulls failures.",
		})}
}

func (s *PullStorer) Metrics() []prometheus.Collector {
	return m.PrometheusCollectorsFromFields(s.metrics)
}
