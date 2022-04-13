package localstore

import (
	"context"
	"time"

	"github.com/redesblock/hop/core/shed"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
)

// GetMulti returns chunks from the database. If one of the chunks is not found
// storage.ErrNotFound will be returned. All required indexes will be updated
// required by the Getter Mode. GetMulti is required to implement chunk.Store
// interface.
func (db *DB) GetMulti(ctx context.Context, mode storage.ModeGet, addrs ...swarm.Address) (chunks []swarm.Chunk, err error) {
	db.metrics.ModeGetMulti.Inc()
	defer totalTimeMetric(db.metrics.TotalTimeGetMulti, time.Now())

	defer func() {
		if err != nil {
			db.metrics.ModeGetMultiFailure.Inc()
		}
	}()

	out, err := db.getMulti(mode, addrs...)
	if err != nil {
		if err == shed.ErrNotFound {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	chunks = make([]swarm.Chunk, len(out))
	for i, ch := range out {
		chunks[i] = swarm.NewChunk(swarm.NewAddress(ch.Address), ch.Data).WithPinCounter(ch.PinCounter)
	}
	return chunks, nil
}

// getMulti returns Items from the retrieval index
// and updates other indexes.
func (db *DB) getMulti(mode storage.ModeGet, addrs ...swarm.Address) (out []shed.Item, err error) {
	out = make([]shed.Item, len(addrs))
	for i, addr := range addrs {
		out[i].Address = addr.Bytes()
	}

	err = db.retrievalDataIndex.Fill(out)
	if err != nil {
		return nil, err
	}

	switch mode {
	// update the access timestamp and gc index
	case storage.ModeGetRequest:
		db.updateGCItems(out...)

	case storage.ModeGetPin:
		err := db.pinIndex.Fill(out)
		if err != nil {
			return nil, err
		}

	// no updates to indexes
	case storage.ModeGetSync:
	case storage.ModeGetLookup:
	default:
		return out, ErrInvalidMode
	}
	return out, nil
}
