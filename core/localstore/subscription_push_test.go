package localstore

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
)

// TestDB_SubscribePush uploads some chunks before and after
// push syncing subscription is created and validates if
// all addresses are received in the right order.
func TestDB_SubscribePush(t *testing.T) {
	db, cleanupFunc := newTestDB(t, nil)
	defer cleanupFunc()

	chunks := make([]swarm.Chunk, 0)
	var chunksMu sync.Mutex

	uploadRandomChunks := func(count int) {
		chunksMu.Lock()
		defer chunksMu.Unlock()

		for i := 0; i < count; i++ {
			ch := generateTestRandomChunk()

			_, err := db.Put(context.Background(), storage.ModePutUpload, ch)
			if err != nil {
				t.Fatal(err)
			}

			chunks = append(chunks, ch)
		}
	}

	// prepopulate database with some chunks
	// before the subscription
	uploadRandomChunks(10)

	// set a timeout on subscription
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// collect all errors from validating addresses, even nil ones
	// to validate the number of addresses received by the subscription
	errChan := make(chan error)

	ch, stop := db.SubscribePush(ctx)
	defer stop()

	// receive and validate addresses from the subscription
	go func() {
		var i int // address index
		for {
			select {
			case got, ok := <-ch:
				if !ok {
					return
				}
				chunksMu.Lock()
				want := chunks[i]
				chunksMu.Unlock()
				var err error
				if !bytes.Equal(got.Data(), want.Data()) {
					err = fmt.Errorf("got chunk %v data %x, want %x", i, got.Data(), want.Data())
				}
				if !got.Address().Equal(want.Address()) {
					err = fmt.Errorf("got chunk %v address %s, want %s", i, got.Address(), want.Address())
				}
				i++
				// send one and only one error per received address
				select {
				case errChan <- err:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// upload some chunks just after subscribe
	uploadRandomChunks(5)

	time.Sleep(200 * time.Millisecond)

	// upload some chunks after some short time
	// to ensure that subscription will include them
	// in a dynamic environment
	uploadRandomChunks(3)

	checkErrChan(ctx, t, errChan, len(chunks))
}

// TestDB_SubscribePush_multiple uploads chunks before and after
// multiple push syncing subscriptions are created and
// validates if all addresses are received in the right order.
func TestDB_SubscribePush_multiple(t *testing.T) {
	db, cleanupFunc := newTestDB(t, nil)
	defer cleanupFunc()

	addrs := make([]swarm.Address, 0)
	var addrsMu sync.Mutex

	uploadRandomChunks := func(count int) {
		addrsMu.Lock()
		defer addrsMu.Unlock()

		for i := 0; i < count; i++ {
			ch := generateTestRandomChunk()

			_, err := db.Put(context.Background(), storage.ModePutUpload, ch)
			if err != nil {
				t.Fatal(err)
			}

			addrs = append(addrs, ch.Address())
		}
	}

	// prepopulate database with some chunks
	// before the subscription
	uploadRandomChunks(10)

	// set a timeout on subscription
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// collect all errors from validating addresses, even nil ones
	// to validate the number of addresses received by the subscription
	errChan := make(chan error)

	subsCount := 10

	// start a number of subscriptions
	// that all of them will write every addresses error to errChan
	for j := 0; j < subsCount; j++ {
		ch, stop := db.SubscribePush(ctx)
		defer stop()

		// receive and validate addresses from the subscription
		go func(j int) {
			var i int // address index
			for {
				select {
				case got, ok := <-ch:
					if !ok {
						return
					}
					addrsMu.Lock()
					want := addrs[i]
					addrsMu.Unlock()
					var err error
					if !got.Address().Equal(want) {
						err = fmt.Errorf("got chunk %v address on subscription %v %s, want %s", i, j, got, want)
					}
					i++
					// send one and only one error per received address
					select {
					case errChan <- err:
					case <-ctx.Done():
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}(j)
	}

	// upload some chunks just after subscribe
	uploadRandomChunks(5)

	time.Sleep(200 * time.Millisecond)

	// upload some chunks after some short time
	// to ensure that subscription will include them
	// in a dynamic environment
	uploadRandomChunks(3)

	// number of addresses received by all subscriptions
	wantedChunksCount := len(addrs) * subsCount

	checkErrChan(ctx, t, errChan, wantedChunksCount)
}
