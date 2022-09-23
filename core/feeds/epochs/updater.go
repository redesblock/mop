package epochs

import (
	"context"

	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/feeds"
	"github.com/redesblock/mop/core/storage"
)

var _ feeds.Updater = (*updater)(nil)

// Updater encapsulates a feeds putter to generate successive updates for epoch based feeds
// it persists the last update
type updater struct {
	*feeds.Putter
	last  int64
	epoch feeds.Index
}

// NewUpdater constructs a feed updater
func NewUpdater(putter storage.Putter, signer crypto.Signer, topic []byte) (feeds.Updater, error) {
	p, err := feeds.NewPutter(putter, signer, topic)
	if err != nil {
		return nil, err
	}
	return &updater{Putter: p}, nil
}

// Update pushes an update to the feed through the chunk stores
func (u *updater) Update(ctx context.Context, at int64, payload []byte) error {
	e := next(u.epoch, u.last, uint64(at))
	err := u.Put(ctx, e, at, payload)
	if err != nil {
		return err
	}
	u.last = at
	u.epoch = e
	return nil
}

func (u *updater) Feed() *feeds.Feed {
	return u.Putter.Feed
}
