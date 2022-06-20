package factory

import (
	"github.com/redesblock/hop/core/feeds"
	"github.com/redesblock/hop/core/feeds/epochs"
	"github.com/redesblock/hop/core/feeds/sequence"
	"github.com/redesblock/hop/core/storage"
)

type factory struct {
	storage.Getter
}

func New(getter storage.Getter) feeds.Factory {
	return &factory{getter}
}

func (f *factory) NewLookup(t feeds.Type, feed *feeds.Feed) (feeds.Lookup, error) {
	switch t {
	case feeds.Sequence:
		return sequence.NewAsyncFinder(f.Getter, feed), nil
	case feeds.Epoch:
		return epochs.NewAsyncFinder(f.Getter, feed), nil
	}

	return nil, feeds.ErrFeedTypeNotFound
}
