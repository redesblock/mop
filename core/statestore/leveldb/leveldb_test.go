package leveldb_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/redesblock/hop/core/statestore/leveldb"
	"github.com/redesblock/hop/core/statestore/test"
	"github.com/redesblock/hop/core/storage"
)

func TestPersistentStateStore(t *testing.T) {
	test.Run(t, func(t *testing.T) (storage.StateStorer, func()) {
		dir, err := ioutil.TempDir("", "statestore_test")
		if err != nil {
			t.Fatal(err)
		}

		store, err := leveldb.NewStateStore(dir)
		if err != nil {
			t.Fatal(err)
		}

		return store, func() { os.RemoveAll(dir) }
	})

	test.RunPersist(t, func(t *testing.T, dir string) storage.StateStorer {
		store, err := leveldb.NewStateStore(dir)
		if err != nil {
			t.Fatal(err)
		}

		return store
	})
}
