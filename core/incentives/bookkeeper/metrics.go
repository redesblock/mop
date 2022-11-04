package bookkeeper

import (
	"github.com/prometheus/client_golang/prometheus"
	m "github.com/redesblock/mop/core/metrics"
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
	AccountingRefreshAttemptCount            prometheus.Counter
	AccountingNonFatalRefreshFailCount       prometheus.Counter
	AccountingDisconnectsOverdrawCount       prometheus.Counter
	AccountingDisconnectsGhostOverdrawCount  prometheus.Counter
	AccountingDisconnectsReconnectCount      prometheus.Counter
	AccountingBlocksCount                    prometheus.Counter
	AccountingReserveCount                   prometheus.Counter
	TotalOriginatedCreditedAmount            prometheus.Counter
	OriginatedCreditEventsCount              prometheus.Counter
	SettleErrorCount                         prometheus.Counter
	PaymentAttemptCount                      prometheus.Counter
	PaymentErrorCount                        prometheus.Counter
	ErrTimeOutOfSyncAlleged                  prometheus.Counter
	ErrTimeOutOfSyncRecent                   prometheus.Counter
	ErrTimeOutOfSyncInterval                 prometheus.Counter
	ErrRefreshmentBelowExpected              prometheus.Counter
	ErrRefreshmentAboveExpected              prometheus.Counter
}

func newMetrics() metrics {
	subsystem := "bookkeeper"

	return metrics{
		TotalDebitedAmount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "total_debited_amount",
			Help:      "Amount of MOP debited to peers (potential income of the node)",
		}),
		TotalCreditedAmount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "total_credited_amount",
			Help:      "Amount of MOP credited to peers (potential cost of the node)",
		}),
		DebitEventsCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "debit_events_count",
			Help:      "Number of occurrences of MOP debit events towards peers",
		}),
		CreditEventsCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "credit_events_count",
			Help:      "Number of occurrences of MOP credit events towards peers",
		}),
		AccountingDisconnectsEnforceRefreshCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "disconnects_enforce_refresh_count",
			Help:      "Number of occurrences of peers disconnected based on failed refreshment attempts",
		}),
		AccountingRefreshAttemptCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "refresh_attempt_count",
			Help:      "Number of attempts of refresh op",
		}),
		AccountingNonFatalRefreshFailCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "non_fatal_refresh_fail_count",
			Help:      "Number of occurrences of refreshments failing for peers because of peer timestamp ahead of ours",
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
			Help:      "Amount of MOP credited to peers (potential cost of the node) for originated traffic",
		}),
		OriginatedCreditEventsCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "originated_credit_events_count",
			Help:      "Number of occurrences of MOP credit events as originator towards peers",
		}),
		SettleErrorCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "settle_error_count",
			Help:      "Number of errors occurring in settle method",
		}),
		PaymentErrorCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "payment_error_count",
			Help:      "Number of errors occurring during payment op",
		}),
		PaymentAttemptCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "payment_attempt_count",
			Help:      "Number of attempts of payment op",
		}),
		ErrRefreshmentBelowExpected: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "refreshment_below_expected",
			Help:      "Number of times the peer received a refreshment that is below expected",
		}),
		ErrRefreshmentAboveExpected: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "refreshment_above_expected",
			Help:      "Number of times the peer received a refreshment that is above expected",
		}),
		ErrTimeOutOfSyncAlleged: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "time_out_of_sync_alleged",
			Help:      "Number of times the timestamps from peer were decreasing",
		}),
		ErrTimeOutOfSyncRecent: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "time_out_of_sync_recent",
			Help:      "Number of times the timestamps from peer differed from our measurement by more than 2 seconds",
		}),
		ErrTimeOutOfSyncInterval: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "time_out_of_sync_interval",
			Help:      "Number of times the time interval from peer differed from local interval by more than 3 seconds",
		}),
	}
}

// Metrics returns the prometheus Collector for the bookkeeper service.
func (a *Accounting) Metrics() []prometheus.Collector {
	return m.PrometheusCollectorsFromFields(a.metrics)
}
