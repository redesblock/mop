package intervalstore

import (
	"os"
	"testing"

	"github.com/redesblock/hop/core/statestore/leveldb"
	"github.com/redesblock/hop/core/statestore/mock"
	"github.com/redesblock/hop/core/storage"
)

// TestInmemoryStore tests basic functionality of InmemoryStore.
func TestInmemoryStore(t *testing.T) {
	testStore(t, mock.NewStateStore())
}

// TestDBStore tests basic functionality of DBStore.
func TestDBStore(t *testing.T) {
	dir, err := os.MkdirTemp("", "intervals_test_db_store")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	store, err := leveldb.NewStateStore(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	testStore(t, store)
}

// testStore is a helper function to test various Store implementations.
func testStore(t *testing.T, s storage.StateStorer) {
	key1 := "key1"
	i1 := NewIntervals(0)
	i1.Add(10, 20)
	if err := s.Put(key1, i1); err != nil {
		t.Fatal(err)
	}
	i := &Intervals{}
	err := s.Get(key1, i)
	if err != nil {
		t.Fatal(err)
	}
	if i.String() != i1.String() {
		t.Errorf("expected interval %s, got %s", i1, i)
	}

	key2 := "key2"
	i2 := NewIntervals(0)
	i2.Add(10, 20)
	if err := s.Put(key2, i2); err != nil {
		t.Fatal(err)
	}
	err = s.Get(key2, i)
	if err != nil {
		t.Fatal(err)
	}
	if i.String() != i2.String() {
		t.Errorf("expected interval %s, got %s", i2, i)
	}

	if err := s.Delete(key1); err != nil {
		t.Fatal(err)
	}
	if err := s.Get(key1, i); err != storage.ErrNotFound {
		t.Errorf("expected error %v, got %s", storage.ErrNotFound, err)
	}
	if err := s.Get(key2, i); err != nil {
		t.Errorf("expected error %v, got %s", nil, err)
	}

	if err := s.Delete(key2); err != nil {
		t.Fatal(err)
	}
	if err := s.Get(key2, i); err != storage.ErrNotFound {
		t.Errorf("expected error %v, got %s", storage.ErrNotFound, err)
	}
}
