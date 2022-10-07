package accounting

import (
	"time"

	"github.com/redesblock/mop/core/flock"
)

func (s *Accounting) SetTimeNow(f func() time.Time) {
	s.timeNow = f
}

func (s *Accounting) SetTime(k int64) {
	s.SetTimeNow(func() time.Time {
		return time.Unix(k, 0)
	})
}

func (a *Accounting) IsPaymentOngoing(peer flock.Address) bool {
	return a.getAccountingPeer(peer).paymentOngoing
}
