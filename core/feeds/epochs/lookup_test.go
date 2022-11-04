package epochs_test

import (
	"testing"

	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/feeds"
	"github.com/redesblock/mop/core/feeds/epochs"
	feedstesting "github.com/redesblock/mop/core/feeds/testing"
	"github.com/redesblock/mop/core/storer/storage"
)

func TestFinder(t *testing.T) {
	t.Skip("test flakes")
	testf := func(t *testing.T, finderf func(storage.Getter, *feeds.Feed) feeds.Lookup, updaterf func(putter storage.Putter, signer crypto.Signer, topic []byte) (feeds.Updater, error)) {
		t.Run("basic", func(t *testing.T) {
			feedstesting.TestFinderBasic(t, finderf, updaterf)
		})
		i := int64(0)
		nextf := func() (bool, int64) {
			defer func() { i++ }()
			return i < 50, i
		}
		t.Run("fixed", func(t *testing.T) {
			feedstesting.TestFinderFixIntervals(t, nextf, finderf, updaterf)
		})
		t.Run("random", func(t *testing.T) {
			feedstesting.TestFinderRandomIntervals(t, finderf, updaterf)
		})
	}
	t.Run("chainsync", func(t *testing.T) {
		testf(t, epochs.NewFinder, epochs.NewUpdater)
	})
	t.Run("async", func(t *testing.T) {
		testf(t, epochs.NewAsyncFinder, epochs.NewUpdater)
	})
}
