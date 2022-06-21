package localstore

import (
	"context"
	"errors"
	"testing"

	"github.com/redesblock/hop/core/storage"
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
				for _, ch := range chunks {
					wantErr := leveldb.ErrNotFound
					_, err := db.retrievalDataIndex.Get(addressToItem(ch.Address()))
					if !errors.Is(err, wantErr) {
						t.Errorf("got error %v, want %v", err, wantErr)
					}

					// access index should not be set
					_, err = db.retrievalAccessIndex.Get(addressToItem(ch.Address()))
					if !errors.Is(err, wantErr) {
						t.Errorf("got error %v, want %v", err, wantErr)
					}
				}

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
