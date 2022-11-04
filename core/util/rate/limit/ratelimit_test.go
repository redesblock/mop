package limit_test

import (
	"testing"
	"time"

	ratelimit "github.com/redesblock/mop/core/util/rate/limit"
)

func TestRateLimit(t *testing.T) {

	var (
		key1  = "test1"
		key2  = "test2"
		rate  = time.Second
		burst = 10
	)

	limiter := ratelimit.New(rate, burst)

	if !limiter.Allow(key1, burst) {
		t.Fatal("want allowed")
	}

	if limiter.Allow(key1, burst) {
		t.Fatalf("want not allowed")
	}

	limiter.Clear(key1)

	if !limiter.Allow(key1, burst) {
		t.Fatal("want allowed")
	}

	if !limiter.Allow(key2, burst) {
		t.Fatal("want allowed")
	}
}
