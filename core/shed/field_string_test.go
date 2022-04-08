package shed

import (
	"io/ioutil"
	"testing"

	"github.com/redesblock/hop/core/logging"
	"github.com/syndtr/goleveldb/leveldb"
)

// TestStringField validates put and get operations
// of the StringField.
func TestStringField(t *testing.T) {
	db, cleanupFunc := newTestDB(t)
	defer cleanupFunc()

	logger := logging.New(ioutil.Discard, 0)
	simpleString, err := db.NewStringField("simple-string", logger)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("get empty", func(t *testing.T) {
		got, err := simpleString.Get()
		if err != nil {
			t.Fatal(err)
		}
		want := ""
		if got != want {
			t.Errorf("got string %q, want %q", got, want)
		}
	})

	t.Run("put", func(t *testing.T) {
		want := "simple string value"
		err = simpleString.Put(want)
		if err != nil {
			t.Fatal(err)
		}
		got, err := simpleString.Get()
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("got string %q, want %q", got, want)
		}

		t.Run("overwrite", func(t *testing.T) {
			want := "overwritten string value"
			err = simpleString.Put(want)
			if err != nil {
				t.Fatal(err)
			}
			got, err := simpleString.Get()
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Errorf("got string %q, want %q", got, want)
			}
		})
	})

	t.Run("put in batch", func(t *testing.T) {
		batch := new(leveldb.Batch)
		want := "simple string batch value"
		simpleString.PutInBatch(batch, want)
		err = db.WriteBatch(batch)
		if err != nil {
			t.Fatal(err)
		}
		got, err := simpleString.Get()
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("got string %q, want %q", got, want)
		}

		t.Run("overwrite", func(t *testing.T) {
			batch := new(leveldb.Batch)
			want := "overwritten string batch value"
			simpleString.PutInBatch(batch, want)
			err = db.WriteBatch(batch)
			if err != nil {
				t.Fatal(err)
			}
			got, err := simpleString.Get()
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Errorf("got string %q, want %q", got, want)
			}
		})
	})
}
