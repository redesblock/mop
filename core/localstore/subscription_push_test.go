package localstore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/redesblock/mop/core/storage"
	"github.com/redesblock/mop/core/swarm"
)

// TestDB_SubscribePush uploads some chunks before and after
// push syncing subscription is created and validates if
// all addresses are received in the right order.
func TestDB_SubscribePush(t *testing.T) {
	db := newTestDB(t, nil)

	chunks := make([]swarm.Chunk, 0)
	var chunksMu sync.Mutex

	chunkProcessedTimes := make([]int, 0)

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

			chunkProcessedTimes = append(chunkProcessedTimes, 0)
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

	ch, _, stop := db.SubscribePush(ctx, func(_ []byte) bool { return false })
	defer stop()

	// receive and validate addresses from the subscription
	go func() {
		var (
			err, ierr           error
			i                   int // address index
			gotStamp, wantStamp []byte
		)
		for {
			select {
			case got, ok := <-ch:
				if !ok {
					return
				}
				chunksMu.Lock()
				cIndex := i
				want := chunks[cIndex]
				chunkProcessedTimes[cIndex]++
				chunksMu.Unlock()
				if !bytes.Equal(got.Data(), want.Data()) {
					err = fmt.Errorf("got chunk %v data %x, want %x", i, got.Data(), want.Data())
				}
				if !got.Address().Equal(want.Address()) {
					err = fmt.Errorf("got chunk %v address %s, want %s", i, got.Address(), want.Address())
				}
				if gotStamp, ierr = got.Stamp().MarshalBinary(); ierr != nil {
					err = ierr
				}
				if wantStamp, ierr = want.Stamp().MarshalBinary(); ierr != nil {
					err = ierr
				}
				if !bytes.Equal(gotStamp, wantStamp) {
					err = errors.New("stamps don't match")
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

	chunksMu.Lock()
	for i, pc := range chunkProcessedTimes {
		if pc != 1 {
			t.Fatalf("chunk on address %s processed %d times, should be only once", chunks[i].Address(), pc)
		}
	}
	chunksMu.Unlock()
}

// TestDB_SubscribePush_multiple uploads chunks before and after
// multiple push syncing subscriptions are created and
// validates if all addresses are received in the right order.
func TestDB_SubscribePush_multiple(t *testing.T) {
	db := newTestDB(t, nil)

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
		ch, _, stop := db.SubscribePush(ctx, func(_ []byte) bool { return false })
		defer stop()

		// receive and validate addresses from the subscription
		go func(j int) {
			var err error
			var i int // address index
			for {
				select {
				case got, ok := <-ch:
					if !ok {
						return
					}
					addrsMu.Lock()
					aIndex := i
					want := addrs[aIndex]
					addrsMu.Unlock()
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

// TestDB_SubscribePush_iterator_restart tests that the
// iterator restart functionality works correctly.
func TestDB_SubscribePush_iterator_restart(t *testing.T) {
	db := newTestDB(t, nil)

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

	uploadRandomChunks(10)

	// set a timeout on subscription
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	skip := false
	ch, restart, stop := db.SubscribePush(ctx, func(addr []byte) bool {
		// in a later case we would like to skip the first item
		if skip && bytes.Equal(addr, addrs[0].Bytes()) {
			return true
		}
		return false
	})
	defer stop()

	consume := func(start int) {
		for i := start; i < len(addrs); i++ {
			got := <-ch
			want := addrs[i]
			if !got.Address().Equal(want) {
				t.Fatalf("got wrong chunk %v address on subscription %s, want %s", i, got, want)
			}
		}
	}
	consume(0) // first pass
	restart()  // trigger again and expect all 10 entries to be iterated on
	consume(0)

	skip = true // expect that first item is skipped
	restart()
	consume(1)

	skip = false // now reset again, expect 10 entries
	restart()
	consume(0)
}
