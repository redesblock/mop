package addresses_test

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/redesblock/mop/core/file"
	"github.com/redesblock/mop/core/file/addresses"
	"github.com/redesblock/mop/core/file/joiner"
	filetest "github.com/redesblock/mop/core/file/testing"
	"github.com/redesblock/mop/core/flock"
	"github.com/redesblock/mop/core/storage"
	"github.com/redesblock/mop/core/storage/mock"
)

func TestAddressesGetterIterateChunkAddresses(t *testing.T) {
	store := mock.NewStorer()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// create root chunk with 2 references and the referenced data chunks
	rootChunk := filetest.GenerateTestRandomFileChunk(flock.ZeroAddress, flock.ChunkSize*2, flock.SectionSize*2)
	_, err := store.Put(ctx, storage.ModePutUpload, rootChunk)
	if err != nil {
		t.Fatal(err)
	}

	firstAddress := flock.NewAddress(rootChunk.Data()[8 : flock.SectionSize+8])
	firstChunk := filetest.GenerateTestRandomFileChunk(firstAddress, flock.ChunkSize, flock.ChunkSize)
	_, err = store.Put(ctx, storage.ModePutUpload, firstChunk)
	if err != nil {
		t.Fatal(err)
	}

	secondAddress := flock.NewAddress(rootChunk.Data()[flock.SectionSize+8:])
	secondChunk := filetest.GenerateTestRandomFileChunk(secondAddress, flock.ChunkSize, flock.ChunkSize)
	_, err = store.Put(ctx, storage.ModePutUpload, secondChunk)
	if err != nil {
		t.Fatal(err)
	}

	createdAddresses := []flock.Address{rootChunk.Address(), firstAddress, secondAddress}

	foundAddresses := make(map[string]struct{})
	var foundAddressesMu sync.Mutex

	addressIterFunc := func(addr flock.Address) error {
		foundAddressesMu.Lock()
		defer foundAddressesMu.Unlock()

		foundAddresses[addr.String()] = struct{}{}
		return nil
	}

	addressesGetter := addresses.NewGetter(store, addressIterFunc)

	j, _, err := joiner.New(ctx, addressesGetter, rootChunk.Address())
	if err != nil {
		t.Fatal(err)
	}

	_, err = file.JoinReadAll(ctx, j, io.Discard)
	if err != nil {
		t.Fatal(err)
	}

	if len(createdAddresses) != len(foundAddresses) {
		t.Fatalf("expected to find %d addresses, got %d", len(createdAddresses), len(foundAddresses))
	}

	checkAddressFound := func(t *testing.T, foundAddresses map[string]struct{}, address flock.Address) {
		t.Helper()

		if _, ok := foundAddresses[address.String()]; !ok {
			t.Fatalf("expected address %s not found", address.String())
		}
	}

	for _, createdAddress := range createdAddresses {
		checkAddressFound(t, foundAddresses, createdAddress)
	}
}
