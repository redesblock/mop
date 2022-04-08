package shed

import (
	"io/ioutil"
	"testing"

	"github.com/redesblock/hop/core/logging"
	"github.com/syndtr/goleveldb/leveldb"
)

// TestUint64Vector validates put and get operations
// of the Uint64Vector.
func TestUint64Vector(t *testing.T) {
	db, cleanupFunc := newTestDB(t)
	defer cleanupFunc()
	logger := logging.New(ioutil.Discard, 0)
	bins, err := db.NewUint64Vector("bins", logger)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("get empty", func(t *testing.T) {
		got, err := bins.Get(0)
		if err != nil {
			t.Fatal(err)
		}
		var want uint64
		if got != want {
			t.Errorf("got uint64 %v, want %v", got, want)
		}
	})

	t.Run("put", func(t *testing.T) {
		for _, index := range []uint64{0, 1, 2, 5, 100} {
			var want uint64 = 42 + index
			err = bins.Put(index, want)
			if err != nil {
				t.Fatal(err)
			}
			got, err := bins.Get(index)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Errorf("got %v uint64 %v, want %v", index, got, want)
			}

			t.Run("overwrite", func(t *testing.T) {
				var want uint64 = 84 + index
				err = bins.Put(index, want)
				if err != nil {
					t.Fatal(err)
				}
				got, err := bins.Get(index)
				if err != nil {
					t.Fatal(err)
				}
				if got != want {
					t.Errorf("got %v uint64 %v, want %v", index, got, want)
				}
			})
		}
	})

	t.Run("put in batch", func(t *testing.T) {
		for _, index := range []uint64{0, 1, 2, 3, 5, 10} {
			batch := new(leveldb.Batch)
			var want uint64 = 43 + index
			bins.PutInBatch(batch, index, want)
			err = db.WriteBatch(batch)
			if err != nil {
				t.Fatal(err)
			}
			got, err := bins.Get(index)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Errorf("got %v uint64 %v, want %v", index, got, want)
			}

			t.Run("overwrite", func(t *testing.T) {
				batch := new(leveldb.Batch)
				var want uint64 = 85 + index
				bins.PutInBatch(batch, index, want)
				err = db.WriteBatch(batch)
				if err != nil {
					t.Fatal(err)
				}
				got, err := bins.Get(index)
				if err != nil {
					t.Fatal(err)
				}
				if got != want {
					t.Errorf("got %v uint64 %v, want %v", index, got, want)
				}
			})
		}
	})
}

// TestUint64Vector_Inc validates Inc operation
// of the Uint64Vector.
func TestUint64Vector_Inc(t *testing.T) {
	db, cleanupFunc := newTestDB(t)
	defer cleanupFunc()
	logger := logging.New(ioutil.Discard, 0)
	bins, err := db.NewUint64Vector("bins", logger)
	if err != nil {
		t.Fatal(err)
	}

	for _, index := range []uint64{0, 1, 2, 3, 5, 10} {
		var want uint64 = 1
		got, err := bins.Inc(index)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("got %v uint64 %v, want %v", index, got, want)
		}

		want = 2
		got, err = bins.Inc(index)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("got %v uint64 %v, want %v", index, got, want)
		}
	}
}

// TestUint64Vector_IncInBatch validates IncInBatch operation
// of the Uint64Vector.
func TestUint64Vector_IncInBatch(t *testing.T) {
	db, cleanupFunc := newTestDB(t)
	defer cleanupFunc()
	logger := logging.New(ioutil.Discard, 0)
	bins, err := db.NewUint64Vector("bins", logger)
	if err != nil {
		t.Fatal(err)
	}

	for _, index := range []uint64{0, 1, 2, 3, 5, 10} {
		batch := new(leveldb.Batch)
		var want uint64 = 1
		got, err := bins.IncInBatch(batch, index)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("got %v uint64 %v, want %v", index, got, want)
		}
		err = db.WriteBatch(batch)
		if err != nil {
			t.Fatal(err)
		}
		got, err = bins.Get(index)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("got %v uint64 %v, want %v", index, got, want)
		}

		batch2 := new(leveldb.Batch)
		want = 2
		got, err = bins.IncInBatch(batch2, index)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("got %v uint64 %v, want %v", index, got, want)
		}
		err = db.WriteBatch(batch2)
		if err != nil {
			t.Fatal(err)
		}
		got, err = bins.Get(index)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("got %v uint64 %v, want %v", index, got, want)
		}
	}
}

// TestUint64Vector_Dec validates Dec operation
// of the Uint64Vector.
func TestUint64Vector_Dec(t *testing.T) {
	db, cleanupFunc := newTestDB(t)
	defer cleanupFunc()
	logger := logging.New(ioutil.Discard, 0)
	bins, err := db.NewUint64Vector("bins", logger)
	if err != nil {
		t.Fatal(err)
	}

	for _, index := range []uint64{0, 1, 2, 3, 5, 10} {
		// test overflow protection
		var want uint64
		got, err := bins.Dec(index)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("got %v uint64 %v, want %v", index, got, want)
		}

		want = 32 + index
		err = bins.Put(index, want)
		if err != nil {
			t.Fatal(err)
		}

		want = 31 + index
		got, err = bins.Dec(index)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("got %v uint64 %v, want %v", index, got, want)
		}
	}
}

// TestUint64Vector_DecInBatch validates DecInBatch operation
// of the Uint64Vector.
func TestUint64Vector_DecInBatch(t *testing.T) {
	db, cleanupFunc := newTestDB(t)
	defer cleanupFunc()
	logger := logging.New(ioutil.Discard, 0)
	bins, err := db.NewUint64Vector("bins", logger)
	if err != nil {
		t.Fatal(err)
	}

	for _, index := range []uint64{0, 1, 2, 3, 5, 10} {
		batch := new(leveldb.Batch)
		var want uint64
		got, err := bins.DecInBatch(batch, index)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("got %v uint64 %v, want %v", index, got, want)
		}
		err = db.WriteBatch(batch)
		if err != nil {
			t.Fatal(err)
		}
		got, err = bins.Get(index)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("got %v uint64 %v, want %v", index, got, want)
		}

		batch2 := new(leveldb.Batch)
		want = 42 + index
		bins.PutInBatch(batch2, index, want)
		err = db.WriteBatch(batch2)
		if err != nil {
			t.Fatal(err)
		}
		got, err = bins.Get(index)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("got %v uint64 %v, want %v", index, got, want)
		}

		batch3 := new(leveldb.Batch)
		want = 41 + index
		got, err = bins.DecInBatch(batch3, index)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("got %v uint64 %v, want %v", index, got, want)
		}
		err = db.WriteBatch(batch3)
		if err != nil {
			t.Fatal(err)
		}
		got, err = bins.Get(index)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("got %v uint64 %v, want %v", index, got, want)
		}
	}
}
