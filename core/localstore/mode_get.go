package localstore

import (
	"context"
	"errors"
	"time"

	"github.com/redesblock/mop/core/postage"
	"github.com/redesblock/mop/core/shed"
	"github.com/redesblock/mop/core/storage"
	"github.com/redesblock/mop/core/swarm"
	"github.com/syndtr/goleveldb/leveldb"
)

// Get returns a chunk from the database. If the chunk is
// not found storage.ErrNotFound will be returned.
// All required indexes will be updated required by the
// Getter Mode. Get is required to implement chunk.Store
// interface.
func (db *DB) Get(ctx context.Context, mode storage.ModeGet, addr swarm.Address) (ch swarm.Chunk, err error) {
	db.metrics.ModeGet.Inc()
	defer totalTimeMetric(db.metrics.TotalTimeGet, time.Now())

	defer func() {
		if err != nil {
			db.metrics.ModeGetFailure.Inc()
		}
	}()

	out, err := db.get(mode, addr)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return swarm.NewChunk(swarm.NewAddress(out.Address), out.Data).
		WithVouch(postage.NewVouch(out.BatchID, out.Index, out.Timestamp, out.Sig)), nil
}

// get returns Item from the retrieval index
// and updates other indexes.
func (db *DB) get(mode storage.ModeGet, addr swarm.Address) (out shed.Item, err error) {
	item := addressToItem(addr)

	out, err = db.retrievalDataIndex.Get(item)
	if err != nil {
		return out, err
	}
	switch mode {
	// update the access timestamp and gc index
	case storage.ModeGetRequest:
		db.updateGCItems(out)

	// no updates to indexes
	case storage.ModeGetSync, storage.ModeGetLookup:
	default:
		return out, ErrInvalidMode
	}
	return out, nil
}

// updateGCItems is called when ModeGetRequest is used
// for Get or GetMulti to update access time and gc indexes
// for all returned chunks.
func (db *DB) updateGCItems(items ...shed.Item) {
	if db.updateGCSem != nil {
		// wait before creating new goroutines
		// if updateGCSem buffer id full
		db.updateGCSem <- struct{}{}
	}
	db.updateGCWG.Add(1)
	go func() {
		defer db.updateGCWG.Done()
		if db.updateGCSem != nil {
			// free a spot in updateGCSem buffer
			// for a new goroutine
			defer func() { <-db.updateGCSem }()
		}

		db.metrics.GCUpdate.Inc()
		defer totalTimeMetric(db.metrics.TotalTimeUpdateGC, time.Now())

		for _, item := range items {
			err := db.updateGC(item)
			if err != nil {
				db.metrics.GCUpdateError.Inc()
				db.logger.Errorf("localstore update gc: %v", err)
			}
		}
		// if gc update hook is defined, call it
		if testHookUpdateGC != nil {
			testHookUpdateGC()
		}
	}()
}

// updateGC updates garbage collection index for
// a single item. Provided item is expected to have
// only Address and Data fields with non zero values,
// which is ensured by the get function.
func (db *DB) updateGC(item shed.Item) (err error) {
	db.batchMu.Lock()
	defer db.batchMu.Unlock()
	if db.gcRunning {
		db.dirtyAddresses = append(db.dirtyAddresses, swarm.NewAddress(item.Address))
	}

	batch := new(leveldb.Batch)

	// update accessTimeStamp in retrieve, gc

	i, err := db.retrievalAccessIndex.Get(item)
	switch {
	case err == nil:
		item.AccessTimestamp = i.AccessTimestamp
	case errors.Is(err, leveldb.ErrNotFound):
		// no chunk accesses
	default:
		return err
	}
	if item.AccessTimestamp == 0 {
		// chunk is not yet synced
		// do not add it to the gc index
		return nil
	}
	// delete current entry from the gc index
	err = db.gcIndex.DeleteInBatch(batch, item)
	if err != nil {
		return err
	}

	// update the gc item timestamp in case
	// it exists
	_, err = db.gcIndex.Get(item)
	item.AccessTimestamp = now()
	if err == nil {
		err = db.gcIndex.PutInBatch(batch, item)
		if err != nil {
			return err
		}
	} else if !errors.Is(err, leveldb.ErrNotFound) {
		return err
	}
	// if the item is not in the gc we don't
	// update the gc index, since the item is
	// in the reserve.

	// update retrieve access index
	err = db.retrievalAccessIndex.PutInBatch(batch, item)
	if err != nil {
		return err
	}

	return db.shed.WriteBatch(batch)
}

// testHookUpdateGC is a hook that can provide
// information when a garbage collection index is updated.
var testHookUpdateGC func()
