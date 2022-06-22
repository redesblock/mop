package accounting

import (
	"github.com/prometheus/client_golang/prometheus"
	m "github.com/redesblock/hop/core/metrics"
)

type metrics struct {
	// all metrics fields must be exported
	// to be able to return them by Metrics()
	// using reflection
	TotalDebitedAmount                       prometheus.Counter
	TotalCreditedAmount                      prometheus.Counter
	DebitEventsCount                         prometheus.Counter
	CreditEventsCount                        prometheus.Counter
	AccountingDisconnectsEnforceRefreshCount prometheus.Counter
	AccountingDisconnectsOverdrawCount       prometheus.Counter
	AccountingDisconnectsGhostOverdrawCount  prometheus.Counter
	AccountingDisconnectsReconnectCount      prometheus.Counter
	AccountingBlocksCount                    prometheus.Counter
	AccountingReserveCount                   prometheus.Counter
	TotalOriginatedCreditedAmount            prometheus.Counter
	OriginatedCreditEventsCount              prometheus.Counter
}

func newMetrics() metrics {
	subsystem := "accounting"

	return metrics{
		TotalDebitedAmount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "total_debited_amount",
			Help:      "Amount of HOP debited to peers (potential income of the node)",
		}),
		TotalCreditedAmount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "total_credited_amount",
			Help:      "Amount of HOP credited to peers (potential cost of the node)",
		}),
		DebitEventsCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "debit_events_count",
			Help:      "Number of occurrences of HOP debit events towards peers",
		}),
		CreditEventsCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "credit_events_count",
			Help:      "Number of occurrences of HOP credit events towards peers",
		}),
		AccountingDisconnectsEnforceRefreshCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "disconnects_enforce_refresh_count",
			Help:      "Number of occurrences of peers disconnected based on failed refreshment attempts",
		}),
		AccountingDisconnectsOverdrawCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "disconnects_overdraw_count",
			Help:      "Number of occurrences of peers disconnected based on payment thresholds",
		}),
		AccountingDisconnectsGhostOverdrawCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "disconnects_ghost_overdraw_count",
			Help:      "Number of occurrences of peers disconnected based on undebitable requests thresholds",
		}),
		AccountingDisconnectsReconnectCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "disconnects_reconnect_count",
			Help:      "Number of occurrences of peers disconnected based on early attempt to reconnect",
		}),

		AccountingBlocksCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "accounting_blocks_count",
			Help:      "Number of occurrences of temporarily skipping a peer to avoid crossing their disconnect thresholds",
		}),
		AccountingReserveCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "accounting_reserve_count",
			Help:      "Number of reserve calls",
		}),
		TotalOriginatedCreditedAmount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "total_originated_credited_amount",
			Help:      "Amount of HOP credited to peers (potential cost of the node) for originated traffic",
		}),
		OriginatedCreditEventsCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "originated_credit_events_count",
			Help:      "Number of occurrences of HOP credit events as originator towards peers",
		}),
	}
}

// Metrics returns the prometheus Collector for the accounting service.
func (a *Accounting) Metrics() []prometheus.Collector {
	return m.PrometheusCollectorsFromFields(a.metrics)
}
