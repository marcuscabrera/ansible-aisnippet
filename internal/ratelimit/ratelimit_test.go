package ratelimit_test

import (
	"testing"
	"time"

	"github.com/marcuscabrera/ansible-aisnippet/internal/ratelimit"
)

func TestAcquireWithinLimitDoesNotBlock(t *testing.T) {
	limiter := ratelimit.New(60)
	start := time.Now()
	for i := 0; i < 5; i++ {
		limiter.Acquire()
	}
	elapsed := time.Since(start)
	if elapsed >= time.Second {
		t.Errorf("expected near-instant acquisition for 5 requests at 60 RPM, took %v", elapsed)
	}
}

func TestHighLimitNeverBlocks(t *testing.T) {
	limiter := ratelimit.New(1000)
	start := time.Now()
	for i := 0; i < 50; i++ {
		limiter.Acquire()
	}
	if time.Since(start) >= 500*time.Millisecond {
		t.Error("expected 50 acquires at 1000 RPM to complete in < 500ms")
	}
}
