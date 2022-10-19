package localstore

import (
	"errors"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/storer/shed"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/syndtr/goleveldb/leveldb"
)

// pinCounter returns the pins counter for a given cluster address, provided that the
// address has been pinned.
func (db *DB) pinCounter(address cluster.Address) (uint64, error) {
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
