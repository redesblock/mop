package shed

import (
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

// TestUint64Field validates put and get operations
// of the Uint64Field.
func TestUint64Field(t *testing.T) {
	db := newTestDB(t)

	counter, err := db.NewUint64Field("counter")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("get empty", func(t *testing.T) {
		got, err := counter.Get()
		if err != nil {
			t.Fatal(err)
		}
		var want uint64
		if got != want {
			t.Errorf("got uint64 %v, want %v", got, want)
		}
	})

	t.Run("put", func(t *testing.T) {
		var want uint64 = 42
		err = counter.Put(want)
		if err != nil {
			t.Fatal(err)
		}
		got, err := counter.Get()
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("got uint64 %v, want %v", got, want)
		}

		t.Run("overwrite", func(t *testing.T) {
			var want uint64 = 84
			err = counter.Put(want)
			if err != nil {
				t.Fatal(err)
			}
			got, err := counter.Get()
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Errorf("got uint64 %v, want %v", got, want)
			}
		})
	})

	t.Run("put in batch", func(t *testing.T) {
		batch := new(leveldb.Batch)
		var want uint64 = 42
		counter.PutInBatch(batch, want)
		err = db.WriteBatch(batch)
		if err != nil {
			t.Fatal(err)
		}
		got, err := counter.Get()
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("got uint64 %v, want %v", got, want)
		}

		t.Run("overwrite", func(t *testing.T) {
			batch := new(leveldb.Batch)
			var want uint64 = 84
			counter.PutInBatch(batch, want)
			err = db.WriteBatch(batch)
			if err != nil {
				t.Fatal(err)
			}
			got, err := counter.Get()
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Errorf("got uint64 %v, want %v", got, want)
			}
		})
	})
}

// TestUint64Field_Inc validates Inc operation
// of the Uint64Field.
func TestUint64Field_Inc(t *testing.T) {
	db := newTestDB(t)

	counter, err := db.NewUint64Field("counter")
	if err != nil {
		t.Fatal(err)
	}

	var want uint64 = 1
	got, err := counter.Inc()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got uint64 %v, want %v", got, want)
	}

	want = 2
	got, err = counter.Inc()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got uint64 %v, want %v", got, want)
	}
}

// TestUint64Field_IncInBatch validates IncInBatch operation
// of the Uint64Field.
func TestUint64Field_IncInBatch(t *testing.T) {
	db := newTestDB(t)

	counter, err := db.NewUint64Field("counter")
	if err != nil {
		t.Fatal(err)
	}

	batch := new(leveldb.Batch)
	var want uint64 = 1
	got, err := counter.IncInBatch(batch)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got uint64 %v, want %v", got, want)
	}
	err = db.WriteBatch(batch)
	if err != nil {
		t.Fatal(err)
	}
	got, err = counter.Get()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got uint64 %v, want %v", got, want)
	}

	batch2 := new(leveldb.Batch)
	want = 2
	got, err = counter.IncInBatch(batch2)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got uint64 %v, want %v", got, want)
	}
	err = db.WriteBatch(batch2)
	if err != nil {
		t.Fatal(err)
	}
	got, err = counter.Get()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got uint64 %v, want %v", got, want)
	}
}

// TestUint64Field_Dec validates Dec operation
// of the Uint64Field.
func TestUint64Field_Dec(t *testing.T) {
	db := newTestDB(t)

	counter, err := db.NewUint64Field("counter")
	if err != nil {
		t.Fatal(err)
	}

	// test overflow protection
	var want uint64
	got, err := counter.Dec()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got uint64 %v, want %v", got, want)
	}

	want = 32
	err = counter.Put(want)
	if err != nil {
		t.Fatal(err)
	}

	want = 31
	got, err = counter.Dec()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got uint64 %v, want %v", got, want)
	}
}

// TestUint64Field_DecInBatch validates DecInBatch operation
// of the Uint64Field.
func TestUint64Field_DecInBatch(t *testing.T) {
	db := newTestDB(t)

	counter, err := db.NewUint64Field("counter")
	if err != nil {
		t.Fatal(err)
	}

	batch := new(leveldb.Batch)
	var want uint64
	got, err := counter.DecInBatch(batch)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got uint64 %v, want %v", got, want)
	}
	err = db.WriteBatch(batch)
	if err != nil {
		t.Fatal(err)
	}
	got, err = counter.Get()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got uint64 %v, want %v", got, want)
	}

	batch2 := new(leveldb.Batch)
	want = 42
	counter.PutInBatch(batch2, want)
	err = db.WriteBatch(batch2)
	if err != nil {
		t.Fatal(err)
	}
	got, err = counter.Get()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got uint64 %v, want %v", got, want)
	}

	batch3 := new(leveldb.Batch)
	want = 41
	got, err = counter.DecInBatch(batch3)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got uint64 %v, want %v", got, want)
	}
	err = db.WriteBatch(batch3)
	if err != nil {
		t.Fatal(err)
	}
	got, err = counter.Get()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got uint64 %v, want %v", got, want)
	}
}
