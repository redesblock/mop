package shed

import (
	"testing"
)

// TestStringField validates put and get operations
// of the StringField.
func TestStringField(t *testing.T) {
	db, cleanupFunc := newTestDB(t)
	defer cleanupFunc()

	simpleString, err := db.NewStringField("simple-string")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("get empty", func(t *testing.T) {
		got, err := simpleString.Get()
		if err == nil {
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
		batch := db.GetBatch(true)
		want := "simple string batch value"
		err = simpleString.PutInBatch(batch, want)
		if err != nil {
			t.Fatal(err)
		}
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
			batch := db.GetBatch(true)
			want := "overwritten string batch value"
			err = simpleString.PutInBatch(batch, want)
			if err != nil {
				t.Fatal(err)
			}
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
