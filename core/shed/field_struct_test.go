package shed

import (
	"testing"
)

// TestStructField validates put and get operations
// of the StructField.
func TestStructField(t *testing.T) {
	db, cleanupFunc := newTestDB(t)
	defer cleanupFunc()

	complexField, err := db.NewStructField("complex-field")
	if err != nil {
		t.Fatal(err)
	}

	type complexStructure struct {
		A string
	}

	t.Run("get empty", func(t *testing.T) {
		var s complexStructure
		err := complexField.Get(&s)
		if err != ErrNotFound {
			t.Fatalf("got error %v, want %v", err, ErrNotFound)
		}
		want := ""
		if s.A != want {
			t.Errorf("got string %q, want %q", s.A, want)
		}
	})

	t.Run("put", func(t *testing.T) {
		want := complexStructure{
			A: "simple string value",
		}
		err = complexField.Put(want)
		if err != nil {
			t.Fatal(err)
		}
		var got complexStructure
		err = complexField.Get(&got)
		if err != nil {
			t.Fatal(err)
		}
		if got.A != want.A {
			t.Errorf("got string %q, want %q", got.A, want.A)
		}

		t.Run("overwrite", func(t *testing.T) {
			want := complexStructure{
				A: "overwritten string value",
			}
			err = complexField.Put(want)
			if err != nil {
				t.Fatal(err)
			}
			var got complexStructure
			err = complexField.Get(&got)
			if err != nil {
				t.Fatal(err)
			}
			if got.A != want.A {
				t.Errorf("got string %q, want %q", got.A, want.A)
			}
		})
	})

	t.Run("put in batch", func(t *testing.T) {
		batch := db.GetBatch(true)
		want := complexStructure{
			A: "simple string batch value",
		}
		err = complexField.PutInBatch(batch, want)
		if err != nil {
			t.Fatal(err)
		}
		err = db.WriteBatch(batch)
		if err != nil {
			t.Fatal(err)
		}
		var got complexStructure
		err := complexField.Get(&got)
		if err != nil {
			t.Fatal(err)
		}
		if got.A != want.A {
			t.Errorf("got string %q, want %q", got, want)
		}

		t.Run("overwrite", func(t *testing.T) {
			batch := db.GetBatch(true)
			want := complexStructure{
				A: "overwritten string batch value",
			}
			err = complexField.PutInBatch(batch, want)
			if err != nil {
				t.Fatal(err)
			}
			err = db.WriteBatch(batch)
			if err != nil {
				t.Fatal(err)
			}
			var got complexStructure
			err := complexField.Get(&got)
			if err != nil {
				t.Fatal(err)
			}
			if got.A != want.A {
				t.Errorf("got string %q, want %q", got, want)
			}
		})
	})
}
