package localstore

import (
	"context"
	"time"

	"github.com/redesblock/mop/core/cluster"
)

// Has returns true if the chunk is stored in database.
func (db *DB) Has(ctx context.Context, addr cluster.Address) (bool, error) {

	db.metrics.ModeHas.Inc()
	defer totalTimeMetric(db.metrics.TotalTimeHas, time.Now())

	has, err := db.retrievalDataIndex.Has(addressToItem(addr))
	if err != nil {
		db.metrics.ModeHasFailure.Inc()
	}
	return has, err
}

// HasMulti returns a slice of booleans which represent if the provided chunks
// are stored in database.
func (db *DB) HasMulti(ctx context.Context, addrs ...cluster.Address) ([]bool, error) {

	db.metrics.ModeHasMulti.Inc()
	defer totalTimeMetric(db.metrics.TotalTimeHasMulti, time.Now())

	have, err := db.retrievalDataIndex.HasMulti(addressesToItems(addrs...)...)
	if err != nil {
		db.metrics.ModeHasMultiFailure.Inc()
	}
	return have, err
}
