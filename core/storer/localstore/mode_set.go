package localstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/storer/sharky"
	"github.com/redesblock/mop/core/storer/shed"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/redesblock/mop/core/tags"
	"github.com/syndtr/goleveldb/leveldb"
)

// Set updates database indexes for
// chunks represented by provided addresses.
// Set is required to implement chunk.Store
// interface.
func (db *DB) Set(ctx context.Context, mode storage.ModeSet, addrs ...cluster.Address) (err error) {
	db.metrics.ModeSet.Inc()
	defer totalTimeMetric(db.metrics.TotalTimeSet, time.Now())
	err = db.set(ctx, mode, addrs...)
	if err != nil {
		db.metrics.ModeSetFailure.Inc()
	}
	return err
}

// set updates database indexes for
// chunks represented by provided addresses.
func (db *DB) set(ctx context.Context, mode storage.ModeSet, addrs ...cluster.Address) (err error) {
	// protect parallel updates
	db.batchMu.Lock()
	defer db.batchMu.Unlock()
	if db.gcRunning {
		db.dirtyAddresses = append(db.dirtyAddresses, addrs...)
	}

	batch := new(leveldb.Batch)
	var committedLocations []sharky.Location

	// variables that provide information for operations
	// to be done after write batch function successfully executes
	var (
		gcSizeChange      int64 // number to add or subtract from gcSize
		reserveSizeChange int64 // number of items to add or subtract from reserveSize
	)
	triggerPullFeed := make(map[uint8]struct{}) // signal pull feed subscriptions to iterate

	switch mode {

	case storage.ModeSetSync:
		for _, addr := range addrs {
			c, r, err := db.setSync(batch, addr)
			if err != nil {
				return err
			}
			gcSizeChange += c
			reserveSizeChange += r
		}

	case storage.ModeSetRemove:
		for _, addr := range addrs {
			item := addressToItem(addr)
			item, err = db.retrievalDataIndex.Get(item)
			if err != nil {
				return err
			}
			c, err := db.setRemove(batch, item, true)
			if err != nil {
				return err
			}
			l, err := sharky.LocationFromBinary(item.Location)
			if err != nil {
				return err
			}
			committedLocations = append(committedLocations, l)
			gcSizeChange += c
		}

	case storage.ModeSetPin:
		for _, addr := range addrs {
			item := addressToItem(addr)
			c, err := db.setPin(batch, item)
			if err != nil {
				return err
			}
			gcSizeChange += c
		}
	case storage.ModeSetUnpin:
		for _, addr := range addrs {
			c, err := db.setUnpin(batch, addr)
			if err != nil {
				return err
			}
			gcSizeChange += c
		}
	default:
		return ErrInvalidMode
	}

	err = db.incGCSizeInBatch(batch, gcSizeChange)
	if err != nil {
		return err
	}

	err = db.incReserveSizeInBatch(batch, reserveSizeChange)
	if err != nil {
		return err
	}

	err = db.shed.WriteBatch(batch)
	if err != nil {
		return err
	}

	sharkyErr := new(multierror.Error)
	for _, l := range committedLocations {
		sharkyErr = multierror.Append(sharkyErr, db.sharky.Release(ctx, l))
	}
	if sharkyErr.ErrorOrNil() != nil {
		return sharkyErr.ErrorOrNil()
	}

	for po := range triggerPullFeed {
		db.triggerPullSubscriptions(po)
	}
	return nil
}

// setSync adds the chunk to the garbage collection after syncing by updating indexes
// - ModeSetSync - the corresponding tag is incremented, then item is removed
//   from push chainsync index
// - update to gc index happens given item does not exist in pins index
// Provided batch is updated.
func (db *DB) setSync(batch *leveldb.Batch, addr cluster.Address) (gcSizeChange, reserveSizeChange int64, err error) {
	item := addressToItem(addr)

	// need to get access timestamp here as it is not
	// provided by the access function, and it is not
	// a property of a chunk provided to Accessor.Put.

	i, err := db.retrievalDataIndex.Get(item)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			// chunk is not found,
			// no need to update gc index
			// just delete from the push index
			// if it is there
			err = db.pushIndex.DeleteInBatch(batch, item)
			if err != nil {
				return 0, 0, err
			}
			return 0, 0, nil
		}
		return 0, 0, err
	}
	item.StoreTimestamp = i.StoreTimestamp
	item.BinID = i.BinID
	item.BatchID = i.BatchID
	item.Index = i.Index
	item.Timestamp = i.Timestamp

	i, err = db.pushIndex.Get(item)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			// we handle this error internally, since this is an internal inconsistency of the indices
			// this error can happen if the chunk is put with ModePutRequest or ModePutSync
			// but this function is called with ModeSetSync
			db.logger.Debug("chunk not found in push index", "address", addr)
		} else {
			return 0, 0, err
		}
	}
	if err == nil && db.tags != nil && i.Tag != 0 {
		t, err := db.tags.Get(i.Tag)
		if err != nil {
			// we cannot break or return here since the function needs to
			// run to end from db.pushIndex.DeleteInBatch
			db.logger.Error(err, "get tags on push chainsync set failed", "uid", i.Tag)
		} else {
			err = t.Inc(tags.StateSynced)
			if err != nil {
				return 0, 0, err
			}
		}
	}

	err = db.pushIndex.DeleteInBatch(batch, item)
	if err != nil {
		return 0, 0, err
	}

	i1, err := db.retrievalAccessIndex.Get(item)
	if err != nil {
		if !errors.Is(err, leveldb.ErrNotFound) {
			return 0, 0, err
		}
		item.AccessTimestamp = now()
		err := db.retrievalAccessIndex.PutInBatch(batch, item)
		if err != nil {
			return 0, 0, err
		}
	} else {
		item.AccessTimestamp = i1.AccessTimestamp
	}
	return db.preserveOrCache(batch, item, false, false)
}

// setRemove removes the chunk by updating indexes:
//  - delete from retrieve, pull, gc
// Provided batch is updated.
func (db *DB) setRemove(batch *leveldb.Batch, item shed.Item, check bool) (gcSizeChange int64, err error) {
	if item.AccessTimestamp == 0 {
		i, err := db.retrievalAccessIndex.Get(item)
		switch {
		case err == nil:
			item.AccessTimestamp = i.AccessTimestamp
		case errors.Is(err, leveldb.ErrNotFound):
		default:
			return 0, err
		}
	}

	if item.StoreTimestamp == 0 || item.Location == nil {
		item, err = db.retrievalDataIndex.Get(item)
		if err != nil {
			return 0, err
		}
	}

	db.metrics.GCStoreTimeStamps.Set(float64(item.StoreTimestamp))
	db.metrics.GCStoreAccessTimeStamps.Set(float64(item.AccessTimestamp))

	err = db.retrievalDataIndex.DeleteInBatch(batch, item)
	if err != nil {
		return 0, err
	}
	err = db.retrievalAccessIndex.DeleteInBatch(batch, item)
	if err != nil {
		return 0, err
	}
	err = db.pushIndex.DeleteInBatch(batch, item)
	if err != nil {
		return 0, err
	}
	err = db.pullIndex.DeleteInBatch(batch, item)
	if err != nil {
		return 0, err
	}
	err = db.voucherChunksIndex.DeleteInBatch(batch, item)
	if err != nil {
		return 0, err
	}

	// unless called by GC which iterates through the gcIndex
	// a check is needed for decrementing gcSize
	// as delete is not reporting if the key/value pair is deleted or not
	if check {
		_, err := db.gcIndex.Get(item)
		if err != nil {
			if !errors.Is(err, leveldb.ErrNotFound) {
				return 0, err
			}
			return 0, db.pinIndex.DeleteInBatch(batch, item)
		}
	}
	err = db.gcIndex.DeleteInBatch(batch, item)
	if err != nil {
		return 0, err
	}
	return -1, nil
}

// setPin increments pins counter for the chunk by updating
// pins index and sets the chunk to be excluded from garbage collection.
// Provided batch is updated.
func (db *DB) setPin(batch *leveldb.Batch, item shed.Item) (gcSizeChange int64, err error) {
	// Get the existing pins counter of the chunk
	i, err := db.pinIndex.Get(item)

	// this will not panic because shed.Index.Get returns an instance, not a pointer.
	// we therefore leverage the default value of the pins counter on the item (zero).
	item.PinCounter = i.PinCounter

	if err != nil {
		if !errors.Is(err, leveldb.ErrNotFound) {
			return 0, err
		}
		// if this Address is not pinned yet, then
		i, err := db.retrievalAccessIndex.Get(item)
		if err != nil {
			if !errors.Is(err, leveldb.ErrNotFound) {
				return 0, err
			}
			// not synced yet
		} else {
			item.AccessTimestamp = i.AccessTimestamp
			i, err = db.retrievalDataIndex.Get(item)
			if err != nil {
				return 0, fmt.Errorf("set pin: retrieval data: %w", err)
			}
			item.StoreTimestamp = i.StoreTimestamp
			item.BinID = i.BinID

			err = db.gcIndex.DeleteInBatch(batch, item)
			if err != nil {
				return 0, err
			}
			gcSizeChange = -1
		}
	}

	// Otherwise increase the existing counter by 1
	item.PinCounter++
	err = db.pinIndex.PutInBatch(batch, item)
	if err != nil {
		return 0, err
	}
	return gcSizeChange, nil
}

// setUnpin decrements pins counter for the chunk by updating pins index.
// Provided batch is updated.
func (db *DB) setUnpin(batch *leveldb.Batch, addr cluster.Address) (gcSizeChange int64, err error) {
	item := addressToItem(addr)

	// Get the existing pins counter of the chunk
	i, err := db.pinIndex.Get(item)
	if err != nil {
		return 0, fmt.Errorf("get pins index: %w", err)
	}
	item.PinCounter = i.PinCounter
	// Decrement the pins counter or
	// delete it from pins index if the pins counter has reached 0
	if item.PinCounter > 1 {
		item.PinCounter--
		return 0, db.pinIndex.PutInBatch(batch, item)
	}

	// PinCounter == 0

	err = db.pinIndex.DeleteInBatch(batch, item)
	if err != nil {
		return 0, err
	}
	i, err = db.retrievalDataIndex.Get(item)
	if err != nil {
		return 0, fmt.Errorf("get retrieval data index: %w", err)
	}
	item.StoreTimestamp = i.StoreTimestamp
	item.BinID = i.BinID
	item.BatchID = i.BatchID
	i, err = db.pushIndex.Get(item)
	switch {
	case err == nil:
		// this is a bit odd, but we return a nil here, causing the pending batch to
		// be written to leveldb, removing the item from the pins index, but not moving it to
		// the gc index because it still exists in the push index.
		return 0, nil
	case !errors.Is(err, leveldb.ErrNotFound):
		// err is not leveldb.ErrNotFound
		return 0, fmt.Errorf("get push index: %w", err)
	}

	i, err = db.retrievalAccessIndex.Get(item)
	if err != nil {
		if !errors.Is(err, leveldb.ErrNotFound) {
			return 0, fmt.Errorf("get retrieval access index: %w", err)
		}
		item.AccessTimestamp = now()
		err = db.retrievalAccessIndex.PutInBatch(batch, item)
		if err != nil {
			return 0, err
		}
	} else {
		item.AccessTimestamp = i.AccessTimestamp
	}
	err = db.gcIndex.PutInBatch(batch, item)
	if err != nil {
		return 0, err
	}
	err = db.pullIndex.DeleteInBatch(batch, item)
	if err != nil {
		return 0, err
	}

	gcSizeChange++
	return gcSizeChange, nil
}
