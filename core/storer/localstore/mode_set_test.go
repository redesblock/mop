package localstore

import (
	"context"
	"testing"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/syndtr/goleveldb/leveldb"
)

// TestModeSetRemove validates ModeSetRemove index values on the provided DB.
func TestModeSetRemove(t *testing.T) {
	for _, tc := range multiChunkTestCases {
		t.Run(tc.name, func(t *testing.T) {
			db := newTestDB(t, nil)

			chunks := generateTestRandomChunks(tc.count)

			_, err := db.Put(context.Background(), storage.ModePutUpload, chunks...)
			if err != nil {
				t.Fatal(err)
			}

			err = db.Set(context.Background(), storage.ModeSetRemove, chunkAddresses(chunks)...)
			if err != nil {
				t.Fatal(err)
			}

			t.Run("retrieve indexes", func(t *testing.T) {

				t.Run("retrieve data index count", newItemsCountTest(db.retrievalDataIndex, 0))
				t.Run("retrieve access index count", newItemsCountTest(db.retrievalAccessIndex, 0))
			})

			for _, ch := range chunks {
				newPullIndexTest(db, ch, 0, leveldb.ErrNotFound)(t)
			}

			t.Run("pull index count", newItemsCountTest(db.pullIndex, 0))

			t.Run("gc index count", newItemsCountTest(db.gcIndex, 0))

			t.Run("gc size", newIndexGCSizeTest(db))
		})
	}
}

// TestModeSetRemove_WithSync validates ModeSetRemove index values on the provided DB
// with the syncing flow for a reserved chunk that has been marked for removal.
func TestModeSetRemove_WithSync(t *testing.T) {
	for _, tc := range multiChunkTestCases {
		t.Run(tc.name, func(t *testing.T) {
			db := newTestDB(t, nil)
			var chs []cluster.Chunk
			for i := 0; i < tc.count; i++ {
				ch := generateTestRandomChunkAt(cluster.NewAddress(db.baseKey), 2).WithBatch(2, 3, 2, false)
				_, err := db.UnreserveBatch(ch.Stamp().BatchID(), 2)
				if err != nil {
					t.Fatal(err)
				}
				_, err = db.Put(context.Background(), storage.ModePutUpload, ch)
				if err != nil {
					t.Fatal(err)
				}
				err = db.Set(context.Background(), storage.ModeSetSync, ch.Address())
				if err != nil {
					t.Fatal(err)
				}

				chs = append(chs, ch)
			}

			err := db.Set(context.Background(), storage.ModeSetRemove, chunkAddresses(chs)...)
			if err != nil {
				t.Fatal(err)
			}

			t.Run("retrieve indexes", func(t *testing.T) {

				t.Run("retrieve data index count", newItemsCountTest(db.retrievalDataIndex, 0))
				t.Run("retrieve access index count", newItemsCountTest(db.retrievalAccessIndex, 0))
			})

			t.Run("voucher chunks index count", newItemsCountTest(db.voucherChunksIndex, 0))

			t.Run("voucher index index count", newItemsCountTest(db.voucherIndexIndex, tc.count))

			t.Run("pull index count", newItemsCountTest(db.pullIndex, 0))

			t.Run("gc index count", newItemsCountTest(db.gcIndex, 0))

			t.Run("gc size", newIndexGCSizeTest(db))
		})
	}
}
