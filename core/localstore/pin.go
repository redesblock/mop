package localstore

import (
	"errors"

	"github.com/redesblock/mop/core/flock"
	"github.com/redesblock/mop/core/shed"
	"github.com/redesblock/mop/core/storage"
	"github.com/syndtr/goleveldb/leveldb"
)

// pinCounter returns the pin counter for a given flock address, provided that the
// address has been pinned.
func (db *DB) pinCounter(address flock.Address) (uint64, error) {
	out, err := db.pinIndex.Get(shed.Item{
		Address: address.Bytes(),
	})

	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return 0, storage.ErrNotFound
		}
		return 0, err
	}
	return out.PinCounter, nil
}
