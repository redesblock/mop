package leveldb_test

import (
	"os"
	"testing"

	"github.com/redesblock/hop/core/statestore/leveldb"
	"github.com/redesblock/hop/core/statestore/test"
	"github.com/redesblock/hop/core/storage"
)

func TestPersistentStateStore(t *testing.T) {
	test.Run(t, func(t *testing.T) storage.StateStorer {
		dir, err := os.MkdirTemp("", "statestore_test")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.RemoveAll(dir); err != nil {
				t.Fatal(err)
			}
		})

		store, err := leveldb.NewStateStore(dir, nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := store.Close(); err != nil {
				t.Fatal(err)
			}
		})

		return store
	})

	test.RunPersist(t, func(t *testing.T, dir string) storage.StateStorer {
		store, err := leveldb.NewStateStore(dir, nil)
		if err != nil {
			t.Fatal(err)
		}

		return store
	})
}

func TestGetSchemaName(t *testing.T) {
	dir, err := os.MkdirTemp("", "statestore_test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	})

	store, err := leveldb.NewStateStore(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatal(err)
		}
	})
	n, err := store.GetSchemaName() // expect current
	if err != nil {
		t.Fatal(err)
	}
	if n != leveldb.DbSchemaCurrent {
		t.Fatalf("wanted current db schema but got '%s'", n)
	}
}
