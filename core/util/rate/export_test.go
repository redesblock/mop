package rate

import "time"

func (r *Rate) SetTimeFunc(f func() time.Time) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.now = f
}
