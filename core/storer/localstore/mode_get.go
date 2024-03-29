package localstore

import (
	"context"
	"errors"
	"time"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/incentives/voucher"
	"github.com/redesblock/mop/core/storer/sharky"
	"github.com/redesblock/mop/core/storer/shed"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/syndtr/goleveldb/leveldb"
)

// Get returns a chunk from the database. If the chunk is
// not found storage.ErrNotFound will be returned.
// All required indexes will be updated required by the
// Getter Mode. Get is required to implement chunk.Store
// interface.
func (db *DB) Get(ctx context.Context, mode storage.ModeGet, addr cluster.Address) (ch cluster.Chunk, err error) {
	db.metrics.ModeGet.Inc()
	defer totalTimeMetric(db.metrics.TotalTimeGet, time.Now())

	defer func() {
		if err != nil {
			db.metrics.ModeGetFailure.Inc()
		}
	}()

	out, err := db.get(ctx, mode, addr)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return cluster.NewChunk(cluster.NewAddress(out.Address), out.Data).
		WithStamp(voucher.NewStamp(out.BatchID, out.Index, out.Timestamp, out.Sig)), nil
}

// get returns Item from the retrieval index
// and updates other indexes.
func (db *DB) get(ctx context.Context, mode storage.ModeGet, addr cluster.Address) (out shed.Item, err error) {
	addrStr := addr.String()

	found := false
	if db.lru != nil {
		if val, ok := db.lru.Get(addrStr); ok {
			out = *val.(*shed.Item)
			found = true
			db.metrics.ModeGetMem.Inc()
		}
	}
	if !found {
		item := addressToItem(addr)

		out, err = db.retrievalDataIndex.Get(item)
		if err != nil {
			return out, err
		}

		l, err := sharky.LocationFromBinary(out.Location)
		if err != nil {
			return out, err
		}

		out.Data = make([]byte, l.Length)
		err = db.sharky.Read(ctx, l, out.Data)
		if err != nil {
			return out, err
		}
		if db.lru != nil {
			db.lru.Add(addr.String(), &out)
		}
	}

	switch mode {
	// update the access timestamp and gc index
	case storage.ModeGetRequest:
		var items []shed.Item
		db.updateGCItemKeysMu.Lock()
		db.updateGCItemKeys[addrStr] = &out
		if cnt := len(db.updateGCItemKeys); cnt > maxParallelUpdateGC {
			for _, item := range db.updateGCItemKeys {
				items = append(items, *item)
			}
			db.updateGCItemKeys = make(map[string]*shed.Item)
		}
		db.updateGCItemKeysMu.Unlock()
		if len(items) > 0 {
			db.updateGCItems(items...)
		}
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
		select {
		case db.updateGCSem <- struct{}{}:
		default:
			return
		}
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
				db.logger.Error(err, "localstore update gc failed")
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
		db.dirtyAddresses = append(db.dirtyAddresses, cluster.NewAddress(item.Address))
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
