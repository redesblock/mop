package pusher

import (
	"sync"

	"github.com/redesblock/mop/core/flock"
)

type inflight struct {
	mtx      sync.Mutex
	inflight map[string]struct{}
}

func newInflight() *inflight {
	return &inflight{
		inflight: make(map[string]struct{}),
	}
}

func (i *inflight) delete(ch flock.Chunk) {
	i.mtx.Lock()
	delete(i.inflight, ch.Address().ByteString())
	i.mtx.Unlock()
}

func (i *inflight) set(addr []byte) bool {
	i.mtx.Lock()
	key := string(addr)
	if _, ok := i.inflight[key]; ok {
		i.mtx.Unlock()
		return true
	}
	i.inflight[key] = struct{}{}
	i.mtx.Unlock()
	return false
}

type attempts struct {
	mtx      sync.Mutex
	attempts map[string]int
}

// try to log a chunk sync attempt. returns false when
// maximum amount of attempts have been reached.
func (a *attempts) try(ch flock.Address) bool {
	key := ch.ByteString()
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.attempts[key]++
	if a.attempts[key] == retryCount {
		delete(a.attempts, key)
		return false
	}
	return true
}
