package shed

import (
	"encoding/binary"

	"github.com/dgraph-io/badger/v2"
	"github.com/redesblock/hop/core/logging"
)

// Uint64Vector provides a way to have multiple counters in the database.
// It transparently encodes uint64 type value to bytes.
type Uint64Vector struct {
	db     *DB
	key    []byte
	logger logging.Logger
}

// NewUint64Vector returns a new Uint64Vector.
// It validates its name and type against the database schema.
func (db *DB) NewUint64Vector(name string, logger logging.Logger) (f Uint64Vector, err error) {
	key, err := db.schemaFieldKey(name, "vector-uint64")
	if err != nil {
		return f, err
	}
	return Uint64Vector{
		db:     db,
		key:    key,
		logger: logger,
	}, nil
}

// Get retrieves a uint64 value at index i from the database.
// If the value is not found in the database a 0 value
// is returned and no error.
func (f Uint64Vector) Get(i uint64) (val uint64, err error) {
	b, err := f.db.Get(f.indexKey(i))
	if err != nil {
		if err == ErrNotFound {
			return 0, nil
		}
		return 0, err
	}
	return binary.BigEndian.Uint64(b), nil
}

// Put encodes uin64 value and stores it in the database.
func (f Uint64Vector) Put(i, val uint64) (err error) {
	return f.db.Put(f.indexKey(i), encodeUint64(val))
}

// PutInBatch stores a uint64 value at index i in a batch
// that can be saved later in the database.
func (f Uint64Vector) PutInBatch(batch *badger.Txn, i, val uint64) (err error) {
	return batch.Set(f.indexKey(i), encodeUint64(val))
}

// Inc increments a uint64 value in the database.
// This operation is not goroutine safe.
func (f Uint64Vector) Inc(i uint64) (val uint64, err error) {
	val, err = f.Get(i)
	if err != nil {
		return 0, err
	}
	val++
	return val, f.Put(i, val)
}

// IncInBatch increments a uint64 value at index i in the batch
// by retreiving a value from the database, not the same batch.
// This operation is not goroutine safe.
func (f Uint64Vector) IncInBatch(batch *badger.Txn, i uint64) (val uint64, err error) {
	val, err = f.Get(i)
	if err != nil {
		return 0, err
	}
	val++
	err = f.PutInBatch(batch, i, val)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// Dec decrements a uint64 value at index i in the database.
// This operation is not goroutine safe.
// The field is protected from overflow to a negative value.
func (f Uint64Vector) Dec(i uint64) (val uint64, err error) {
	val, err = f.Get(i)
	if err != nil {
		if err == ErrNotFound {
			val = 0
		} else {
			f.logger.Debugf("error getiing value while doing Dec. Error: %s", err.Error())
			return 0, err
		}
	}
	if val != 0 {
		val--
	}
	return val, f.Put(i, val)
}

// DecInBatch decrements a uint64 value at index i in the batch
// by retreiving a value from the database, not the same batch.
// This operation is not goroutine safe.
// The field is protected from overflow to a negative value.
func (f Uint64Vector) DecInBatch(batch *badger.Txn, i uint64) (val uint64, err error) {
	val, err = f.Get(i)
	if err != nil {
		return 0, err
	}
	if val != 0 {
		val--
	}
	err = f.PutInBatch(batch, i, val)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// indexKey concatenates field prefix and vector index
// returning a unique database key for a specific vector element.
func (f Uint64Vector) indexKey(i uint64) (key []byte) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return append(f.key, b...)
}
