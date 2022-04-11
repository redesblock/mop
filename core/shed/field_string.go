package shed

import (
	"github.com/dgraph-io/badger/v2"
	"github.com/redesblock/hop/core/logging"
)

// StringField is the most simple field implementation
// that stores an arbitrary string under a specific LevelDB key.
type StringField struct {
	db     *DB
	key    []byte
	logger logging.Logger
}

// NewStringField retruns a new Instance of StringField.
// It validates its name and type against the database schema.
func (db *DB) NewStringField(name string, logger logging.Logger) (f StringField, err error) {
	key, err := db.schemaFieldKey(name, "string")
	if err != nil {
		return f, err
	}
	return StringField{
		db:     db,
		key:    key,
		logger: logger,
	}, nil
}

// Get returns a string value from database.
// If the value is not found, an empty string is returned
// an no error.
func (f StringField) Get() (val string, err error) {
	b, err := f.db.Get(f.key)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Put stores a string in the database.
func (f StringField) Put(val string) (err error) {
	return f.db.Put(f.key, []byte(val))
}

// PutInBatch stores a string in a batch that can be
// saved later in database.
func (f StringField) PutInBatch(batch *badger.Txn, val string) (err error) {
	return batch.Set(f.key, []byte(val))
}
