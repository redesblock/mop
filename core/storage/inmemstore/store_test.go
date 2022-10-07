package inmemstore_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/redesblock/mop/core/flock"
	"github.com/redesblock/mop/core/storage"
	"github.com/redesblock/mop/core/storage/inmemstore"
)

func TestStorePutGet(t *testing.T) {
	s := inmemstore.New()

	keyFound, err := flock.ParseHexAddress("aabbcc")
	if err != nil {
		t.Fatal(err)
	}
	keyNotFound, err := flock.ParseHexAddress("bbccdd")
	if err != nil {
		t.Fatal(err)
	}

	valueFound := []byte("data data data")

	ctx := context.Background()
	if _, err := s.Get(ctx, storage.ModeGetRequest, keyFound); err != storage.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if _, err := s.Get(ctx, storage.ModeGetRequest, keyNotFound); err != storage.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if _, err := s.Put(ctx, storage.ModePutUpload, flock.NewChunk(keyFound, valueFound)); err != nil {
		t.Fatalf("expected not error but got: %v", err)
	}

	if chunk, err := s.Get(ctx, storage.ModeGetRequest, keyFound); err != nil {
		t.Fatalf("expected no error but got: %v", err)
	} else if !bytes.Equal(chunk.Data(), valueFound) {
		t.Fatalf("expected value %s but got %s", valueFound, chunk.Data())
	}
}
