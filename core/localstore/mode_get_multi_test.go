package localstore

import (
	"context"
	"reflect"
	"testing"

	"github.com/redesblock/hop/core/storage"
)

// TestModeGetMulti stores chunks and validates that GetMulti
// is returning them correctly.
func TestModeGetMulti(t *testing.T) {
	const chunkCount = 10

	for _, mode := range []storage.ModeGet{
		storage.ModeGetRequest,
		storage.ModeGetSync,
		storage.ModeGetLookup,
		storage.ModeGetPin,
	} {
		t.Run(mode.String(), func(t *testing.T) {
			db, cleanupFunc := newTestDB(t, nil)
			defer cleanupFunc()

			chunks := generateTestRandomChunks(chunkCount)

			_, err := db.Put(context.Background(), storage.ModePutUpload, chunks...)
			if err != nil {
				t.Fatal(err)
			}

			if mode == storage.ModeGetPin {
				// pin chunks so that it is not returned as not found by pinIndex
				for i, ch := range chunks {
					err := db.Set(context.Background(), storage.ModeSetPin, ch.Address())
					if err != nil {
						t.Fatal(err)
					}
					chunks[i] = ch.WithPinCounter(1)
				}
			}

			addrs := chunkAddresses(chunks)

			got, err := db.GetMulti(context.Background(), mode, addrs...)
			if err != nil {
				t.Fatal(err)
			}

			for i := 0; i < chunkCount; i++ {
				if !reflect.DeepEqual(got[i], chunks[i]) {
					t.Errorf("got %v chunk %v, want %v", i, got[i], chunks[i])
				}
			}

			missingChunk := generateTestRandomChunk()

			want := storage.ErrNotFound
			_, err = db.GetMulti(context.Background(), mode, append(addrs, missingChunk.Address())...)
			if err != want {
				t.Errorf("got error %v, want %v", err, want)
			}
		})
	}
}
