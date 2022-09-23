package sequence_test

import (
	"testing"

	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/feeds"
	"github.com/redesblock/mop/core/feeds/sequence"
	feedstesting "github.com/redesblock/mop/core/feeds/testing"
	"github.com/redesblock/mop/core/storage"
)

func TestFinder(t *testing.T) {
	testf := func(t *testing.T, finderf func(storage.Getter, *feeds.Feed) feeds.Lookup, updaterf func(putter storage.Putter, signer crypto.Signer, topic []byte) (feeds.Updater, error)) {
		t.Run("basic", func(t *testing.T) {
			feedstesting.TestFinderBasic(t, finderf, updaterf)
		})
		i := 0
		nextf := func() (bool, int64) {
			i++
			return i == 40, int64(i)
		}
		t.Run("fixed", func(t *testing.T) {
			feedstesting.TestFinderFixIntervals(t, nextf, finderf, updaterf)
		})
		t.Run("random", func(t *testing.T) {
			feedstesting.TestFinderRandomIntervals(t, finderf, updaterf)
		})
	}
	t.Run("sync", func(t *testing.T) {
		testf(t, sequence.NewFinder, sequence.NewUpdater)
	})
	t.Run("async", func(t *testing.T) {
		testf(t, sequence.NewAsyncFinder, sequence.NewUpdater)
	})
}
