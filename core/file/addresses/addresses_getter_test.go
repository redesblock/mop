package addresses_test

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/file"
	"github.com/redesblock/mop/core/file/addresses"
	"github.com/redesblock/mop/core/file/joiner"
	filetest "github.com/redesblock/mop/core/file/testing"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/redesblock/mop/core/storer/storage/mock"
)

func TestAddressesGetterIterateChunkAddresses(t *testing.T) {
	store := mock.NewStorer()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// create root chunk with 2 references and the referenced data chunks
	rootChunk := filetest.GenerateTestRandomFileChunk(cluster.ZeroAddress, cluster.ChunkSize*2, cluster.SectionSize*2)
	_, err := store.Put(ctx, storage.ModePutUpload, rootChunk)
	if err != nil {
		t.Fatal(err)
	}

	firstAddress := cluster.NewAddress(rootChunk.Data()[8 : cluster.SectionSize+8])
	firstChunk := filetest.GenerateTestRandomFileChunk(firstAddress, cluster.ChunkSize, cluster.ChunkSize)
	_, err = store.Put(ctx, storage.ModePutUpload, firstChunk)
	if err != nil {
		t.Fatal(err)
	}

	secondAddress := cluster.NewAddress(rootChunk.Data()[cluster.SectionSize+8:])
	secondChunk := filetest.GenerateTestRandomFileChunk(secondAddress, cluster.ChunkSize, cluster.ChunkSize)
	_, err = store.Put(ctx, storage.ModePutUpload, secondChunk)
	if err != nil {
		t.Fatal(err)
	}

	createdAddresses := []cluster.Address{rootChunk.Address(), firstAddress, secondAddress}

	foundAddresses := make(map[string]struct{})
	var foundAddressesMu sync.Mutex

	addressIterFunc := func(addr cluster.Address) error {
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

	checkAddressFound := func(t *testing.T, foundAddresses map[string]struct{}, address cluster.Address) {
		t.Helper()

		if _, ok := foundAddresses[address.String()]; !ok {
			t.Fatalf("expected address %s not found", address.String())
		}
	}

	for _, createdAddress := range createdAddresses {
		checkAddressFound(t, foundAddresses, createdAddress)
	}
}
