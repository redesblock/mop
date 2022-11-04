package localstore

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/redesblock/mop/core/storer/storage"
)

// TestHas validates that Has method is returning true for
// the stored chunk and false for one that is not stored.
func TestHas(t *testing.T) {
	db := newTestDB(t, nil)

	ch := generateTestRandomChunk()

	_, err := db.Put(context.Background(), storage.ModePutUpload, ch)
	if err != nil {
		t.Fatal(err)
	}

	has, err := db.Has(context.Background(), ch.Address())
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Error("chunk not found")
	}

	missingChunk := generateTestRandomChunk()

	has, err = db.Has(context.Background(), missingChunk.Address())
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Error("unexpected chunk is found")
	}
}

// TestHasMulti validates that HasMulti method is returning correct boolean
// slice for stored chunks.
func TestHasMulti(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, tc := range multiChunkTestCases {
		t.Run(tc.name, func(t *testing.T) {
			db := newTestDB(t, nil)

			chunks := generateTestRandomChunks(tc.count)
			want := make([]bool, tc.count)

			for i, ch := range chunks {
				if r.Intn(2) == 0 {
					// randomly exclude half of the chunks
					continue
				}
				_, err := db.Put(context.Background(), storage.ModePutUpload, ch)
				if err != nil {
					t.Fatal(err)
				}
				want[i] = true
			}

			got, err := db.HasMulti(context.Background(), chunkAddresses(chunks)...)
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprint(got) != fmt.Sprint(want) {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}
