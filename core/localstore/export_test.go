package localstore

import (
	"bytes"
	"context"
	"testing"

	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
)

// TestExportImport constructs two databases, one to put and export
// chunks and another one to import and validate that all chunks are
// imported.
func TestExportImport(t *testing.T) {
	db1 := newTestDB(t, nil)

	var chunkCount = 100

	chunks := make(map[string][]byte, chunkCount)
	for i := 0; i < chunkCount; i++ {
		ch := generateTestRandomChunk()

		_, err := db1.Put(context.Background(), storage.ModePutUpload, ch)
		if err != nil {
			t.Fatal(err)
		}
		stamp, err := ch.Stamp().MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		chunks[ch.Address().String()] = append(stamp, ch.Data()...)
	}

	var buf bytes.Buffer

	c, err := db1.Export(&buf)
	if err != nil {
		t.Fatal(err)
	}
	wantChunksCount := int64(len(chunks))
	if c != wantChunksCount {
		t.Errorf("got export count %v, want %v", c, wantChunksCount)
	}

	db2 := newTestDB(t, nil)

	c, err = db2.Import(context.Background(), &buf)
	if err != nil {
		t.Fatal(err)
	}
	if c != wantChunksCount {
		t.Errorf("got import count %v, want %v", c, wantChunksCount)
	}

	for a, want := range chunks {
		addr := swarm.MustParseHexAddress(a)
		ch, err := db2.Get(context.Background(), storage.ModeGetRequest, addr)
		if err != nil {
			t.Fatal(err)
		}
		stamp, err := ch.Stamp().MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		got := append(stamp, ch.Data()...)
		if !bytes.Equal(got, want) {
			t.Fatalf("chunk %s: got stamp+data %x, want %x", addr, got, want)
		}
	}
}
