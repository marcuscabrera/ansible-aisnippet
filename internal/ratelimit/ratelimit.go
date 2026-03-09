// Package ratelimit provides a thread-safe sliding-window rate limiter.
package ratelimit

import (
	"sync"
	"time"
)

const windowSeconds = 60 * time.Second

// RateLimiter is a thread-safe sliding-window rate limiter.
//
// It tracks request timestamps within the last 60 seconds and blocks (sleeps)
// when the per-minute limit is reached.
type RateLimiter struct {
	requestsPerMinute int
	timestamps        []time.Time
	mu                sync.Mutex
}

// New creates a RateLimiter that allows up to requestsPerMinute requests per 60-second window.
func New(requestsPerMinute int) *RateLimiter {
	return &RateLimiter{
		requestsPerMinute: requestsPerMinute,
	}
}

// Acquire blocks the calling goroutine until a request slot is available.
func (r *RateLimiter) Acquire() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	r.dropOld(now)

	if len(r.timestamps) >= r.requestsPerMinute {
		// Sleep until the oldest request leaves the window.
		waitUntil := r.timestamps[0].Add(windowSeconds)
		waitTime := time.Until(waitUntil)
		if waitTime > 0 {
			r.mu.Unlock()
			time.Sleep(waitTime)
			r.mu.Lock()
			now = time.Now()
			r.dropOld(now)
		}
	}

	r.timestamps = append(r.timestamps, now)
}

// dropOld removes timestamps older than the sliding window. Must be called with mu held.
func (r *RateLimiter) dropOld(now time.Time) {
	cutoff := now.Add(-windowSeconds)
	i := 0
	for i < len(r.timestamps) && r.timestamps[i].Before(cutoff) {
		i++
	}
	r.timestamps = r.timestamps[i:]
}
