package warden_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"sync"
	"testing"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/file/pipeline/builder"
	"github.com/redesblock/mop/core/p2p/topology"
	"github.com/redesblock/mop/core/protocol/pushsync"
	psmock "github.com/redesblock/mop/core/protocol/pushsync/mock"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/redesblock/mop/core/storer/storage/mock"
	"github.com/redesblock/mop/core/traverser"
	"github.com/redesblock/mop/core/warden"
)

func TestSteward(t *testing.T) {
	var (
		ctx            = context.Background()
		chunks         = 1000
		data           = make([]byte, chunks*4096) //1k chunks
		store          = mock.NewStorer()
		traverser      = traverser.New(store)
		loggingStorer  = &loggingStore{Storer: store}
		traversedAddrs = make(map[string]struct{})
		mu             sync.Mutex
		fn             = func(_ context.Context, ch cluster.Chunk) (*pushsync.Receipt, error) {
			mu.Lock()
			traversedAddrs[ch.Address().String()] = struct{}{}
			mu.Unlock()
			return nil, nil
		}
		ps = psmock.New(fn)
		s  = warden.New(store, traverser, loggingStorer, ps)
	)
	n, err := rand.Read(data)
	if n != cap(data) {
		t.Fatal("short read")
	}
	if err != nil {
		t.Fatal(err)
	}

	pipe := builder.NewPipelineBuilder(ctx, loggingStorer, storage.ModePutUpload, false)
	addr, err := builder.FeedPipeline(ctx, pipe, bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	err = s.Reupload(ctx, addr)
	if err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()

	isRetrievable, err := s.IsRetrievable(ctx, addr)
	if err != nil {
		t.Fatal(err)
	}
	if !isRetrievable {
		t.Fatalf("re-uploaded content on %q should be retrievable", addr)
	}

	// check that everything that was stored is also traversed
	for _, a := range loggingStorer.addrs {
		if _, ok := traversedAddrs[a.String()]; !ok {
			t.Fatalf("expected address %s to be traversed", a.String())
		}
	}
}

func TestSteward_ErrWantSelf(t *testing.T) {
	var (
		ctx           = context.Background()
		chunks        = 10
		data          = make([]byte, chunks*4096)
		store         = mock.NewStorer()
		traverser     = traverser.New(store)
		loggingStorer = &loggingStore{Storer: store}
		fn            = func(_ context.Context, ch cluster.Chunk) (*pushsync.Receipt, error) {
			return nil, topology.ErrWantSelf
		}
		ps = psmock.New(fn)
		s  = warden.New(store, traverser, loggingStorer, ps)
	)
	n, err := rand.Read(data)
	if n != cap(data) {
		t.Fatal("short read")
	}
	if err != nil {
		t.Fatal(err)
	}

	pipe := builder.NewPipelineBuilder(ctx, loggingStorer, storage.ModePutUpload, false)
	addr, err := builder.FeedPipeline(ctx, pipe, bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	err = s.Reupload(ctx, addr)
	if err != nil {
		t.Fatal(err)
	}
}

type loggingStore struct {
	storage.Storer
	addrs []cluster.Address
}

func (ls *loggingStore) Put(ctx context.Context, mode storage.ModePut, chs ...cluster.Chunk) (exist []bool, err error) {
	for _, c := range chs {
		ls.addrs = append(ls.addrs, c.Address())
	}
	return ls.Storer.Put(ctx, mode, chs...)
}

func (ls *loggingStore) RetrieveChunk(ctx context.Context, addr, sourceAddr cluster.Address) (chunk cluster.Chunk, err error) {
	return ls.Get(ctx, storage.ModeGetRequest, addr)
}
