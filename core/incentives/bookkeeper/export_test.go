package bookkeeper

import (
	"time"

	"github.com/redesblock/mop/core/cluster"
)

func (a *Accounting) SetTimeNow(f func() time.Time) {
	a.timeNow = f
}

func (a *Accounting) SetTime(k int64) {
	a.SetTimeNow(func() time.Time {
		return time.Unix(k, 0)
	})
}

func (a *Accounting) IsPaymentOngoing(peer cluster.Address) bool {
	return a.getAccountingPeer(peer).paymentOngoing
}
