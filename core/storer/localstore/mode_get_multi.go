package localstore

import (
	"context"
	"errors"
	"time"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/incentives/voucher"
	"github.com/redesblock/mop/core/storer/shed"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/syndtr/goleveldb/leveldb"
)

// GetMulti returns chunks from the database. If one of the chunks is not found
// storage.ErrNotFound will be returned. All required indexes will be updated
// required by the Getter Mode. GetMulti is required to implement chunk.Store
// interface.
func (db *DB) GetMulti(ctx context.Context, mode storage.ModeGet, addrs ...cluster.Address) (chunks []cluster.Chunk, err error) {
	db.metrics.ModeGetMulti.Inc()
	db.metrics.ModeGetMultiChunks.Add(float64(len(addrs)))
	defer totalTimeMetric(db.metrics.TotalTimeGetMulti, time.Now())

	defer func() {
		if err != nil {
			db.metrics.ModeGetMultiFailure.Inc()
		}
	}()

	out, err := db.getMulti(ctx, mode, addrs...)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	chunks = make([]cluster.Chunk, len(out))
	for i, ch := range out {
		chunks[i] = cluster.NewChunk(cluster.NewAddress(ch.Address), ch.Data).
			WithStamp(voucher.NewStamp(ch.BatchID, ch.Index, ch.Timestamp, ch.Sig))
	}
	return chunks, nil
}

// getMulti returns Items from the retrieval index
// and updates other indexes.
func (db *DB) getMulti(ctx context.Context, mode storage.ModeGet, addrs ...cluster.Address) (out []shed.Item, err error) {
	out = make([]shed.Item, len(addrs))
	for i, addr := range addrs {
		if item, err := db.get(ctx, mode, addr); err != nil {
			return nil, err
		} else {
			out[i] = item
		}
	}
	// for i, addr := range addrs {
	// 	out[i].Address = addr.Bytes()
	// }

	// err = db.retrievalDataIndex.Fill(out)
	// if err != nil {
	// 	return nil, err
	// }

	// for i, item := range out {
	// 	l, err := sharky.LocationFromBinary(item.Location)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	out[i].Data = make([]byte, l.Length)
	// 	err = db.sharky.Read(ctx, l, out[i].Data)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	// switch mode {
	// // update the access timestamp and gc index
	// case storage.ModeGetRequest:
	// 	db.updateGCItems(out...)

	// // no updates to indexes
	// case storage.ModeGetSync, storage.ModeGetLookup:
	// default:
	// 	return out, ErrInvalidMode
	// }
	return out, nil
}
