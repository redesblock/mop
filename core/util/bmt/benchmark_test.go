package bmt_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/util/bmt"
	"github.com/redesblock/mop/core/util/bmt/reference"
	"golang.org/x/sync/errgroup"
)

//
func BenchmarkBMT(t *testing.B) {
	for size := testChunkSize; size >= testSegmentSectionSize; size /= 2 {
		t.Run(fmt.Sprintf("%v_size_%v", "SHA3", size), func(t *testing.B) {
			benchmarkSHA3(t, size)
		})
		//t.Run(fmt.Sprintf("%v_size_%v", "Baseline", size), func(t *testing.B) {
		//	benchmarkBMTBaseline(t, size)
		//})
		//t.Run(fmt.Sprintf("%v_size_%v", "REF", size), func(t *testing.B) {
		//	benchmarkRefHasher(t, size)
		//})
		t.Run(fmt.Sprintf("%v_size_%v_num_%v", "BMT", size, calcNumber(int(math.Ceil(float64(size)/float64(testSegmentSectionSize))))), func(t *testing.B) {
			benchmarkBMT(t, size)
		})

		t.Run(fmt.Sprintf("%v_size_%v", "BMTParallel", size), func(t *testing.B) {
			benchmarkBMTParallel(t, size)
		})
	}
}

func benchmarkBMTParallel(b *testing.B, n int) {
	pool := bmt.NewPool(bmt.NewConf(cluster.NewHasher, testSegmentCount, testSegmentSectionSize, testPoolSize))
	testData := randomBytes(b, seed)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h := pool.Get()
			if _, err := syncHash(h, testData[:n]); err != nil {
				b.Fatalf("seed %d: %v", seed, err)
			}
			pool.Put(h)
		}
	})
}

func BenchmarkPool(t *testing.B) {
	for _, c := range []int{1, 8, 16, 32, 64, 128, 256, 512, 1024, 2048} {
		t.Run(fmt.Sprintf("poolsize_%v", c), func(t *testing.B) {
			benchmarkPool(t, c)
		})
	}
}

// benchmarks simple sha3 hash on chunks
func benchmarkSHA3(t *testing.B, n int) {
	testData := randomBytes(t, seed)

	t.ReportAllocs()
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		if _, err := bmt.Sha3hash(testData[:n]); err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
	}
}

// benchmarks the minimum hashing time for a balanced (for simplicity) BMT
// by doing count/segmentsize parallel hashings of 2*segmentsize bytes
// doing it on n testPoolSize each reusing the base hasher
// the premise is that this is the minimum computation needed for a BMT
// therefore this serves as a theoretical optimum for concurrent implementations
func benchmarkBMTBaseline(t *testing.B, n int) {
	testData := randomBytes(t, seed)

	t.ReportAllocs()
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		eg := new(errgroup.Group)
		for j := 0; j < testSegmentCount; j++ {
			eg.Go(func() error {
				_, err := bmt.Sha3hash(testData[:hashSize])
				return err
			})
		}
		if err := eg.Wait(); err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
	}
}

// benchmarks BMT Hasher
func benchmarkBMT(t *testing.B, n int) {
	testData := randomBytes(t, seed)

	pool := bmt.NewPool(bmt.NewConf(cluster.NewHasher, testSegmentCount, testSegmentSectionSize, testPoolSize))
	h := pool.Get()
	defer pool.Put(h)

	t.ReportAllocs()
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		if _, err := syncHash(h, testData[:n]); err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
	}
}

// benchmarks 100 concurrent bmt hashes with pool capacity
func benchmarkPool(t *testing.B, poolsize int) {
	testData := randomBytes(t, seed)

	pool := bmt.NewPool(bmt.NewConf(cluster.NewHasher, testSegmentCount, testSegmentSectionSize, poolsize))
	cycles := poolsize

	t.ReportAllocs()
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		eg := new(errgroup.Group)
		for j := 0; j < cycles; j++ {
			eg.Go(func() error {
				h := pool.Get()
				defer pool.Put(h)
				_, err := syncHash(h, testData[:h.Capacity()])
				return err
			})
		}
		if err := eg.Wait(); err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
	}
}

// benchmarks the reference hasher
func benchmarkRefHasher(t *testing.B, n int) {
	testData := randomBytes(t, seed)

	rbmt := reference.NewRefHasher(cluster.NewHasher(), 128)

	t.ReportAllocs()
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		_, err := rbmt.Hash(testData[:n])
		if err != nil {
			t.Fatal(err)
		}
	}
}

func calcNumber(count int) int {
	num := math.Ceil(float64(count) / float64(2))
	total := num
	for c := 2; c < testSegmentCount; c *= 2 {
		num = math.Ceil(float64(num) / float64(2))
		total += num
	}
	return int(total)
}
