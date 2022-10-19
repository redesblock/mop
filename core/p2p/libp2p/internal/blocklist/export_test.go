package blocklist

import (
	"github.com/redesblock/mop/core/storer/storage"
)

func NewBlocklistWithCurrentTimeFn(store storage.StateStorer, currentTimeFn currentTimeFn) *Blocklist {
	return &Blocklist{
		store:         store,
		currentTimeFn: currentTimeFn,
	}
}
