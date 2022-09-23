package testing

import (
	"testing"

	"github.com/redesblock/mop/core/tags"
)

// CheckTag checks the first tag in the api struct to be in a certain state
func CheckTag(t *testing.T, tag *tags.Tag, split, stored, seen, sent, synced, total int64) {
	t.Helper()
	if tag == nil {
		t.Fatal("no tag found")
	}
	tSplit := tag.Get(tags.StateSplit)
	if tSplit != split {
		t.Fatalf("should have had split chunks, got %d want %d", tSplit, split)
	}

	tSeen := tag.Get(tags.StateSeen)
	if tSeen != seen {
		t.Fatalf("should have had seen chunks, got %d want %d", tSeen, seen)
	}

	tStored := tag.Get(tags.StateStored)
	if tStored != stored {
		t.Fatalf("mismatch stored chunks, got %d want %d", tStored, stored)
	}

	tSent := tag.Get(tags.StateSent)
	if tStored != stored {
		t.Fatalf("mismatch sent chunks, got %d want %d", tSent, sent)
	}

	tSynced := tag.Get(tags.StateSynced)
	if tSynced != synced {
		t.Fatalf("mismatch synced chunks, got %d want %d", tSynced, synced)
	}

	tTotal := tag.TotalCounter()
	if tTotal != total {
		t.Fatalf("mismatch total chunks, got %d want %d", tTotal, total)
	}
}
