package localstore

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/redesblock/mop/core/storer/storage"
)

// TestModeGetMulti stores chunks and validates that GetMulti
// is returning them correctly.
func TestModeGetMulti(t *testing.T) {
	const chunkCount = 10

	for _, mode := range []storage.ModeGet{
		storage.ModeGetRequest,
		storage.ModeGetSync,
		storage.ModeGetLookup,
	} {
		t.Run(mode.String(), func(t *testing.T) {
			db := newTestDB(t, nil)

			chunks := generateTestRandomChunks(chunkCount)

			_, err := db.Put(context.Background(), storage.ModePutUpload, chunks...)
			if err != nil {
				t.Fatal(err)
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
			if !errors.Is(err, want) {
				t.Errorf("got error %v, want %v", err, want)
			}
		})
	}
}
