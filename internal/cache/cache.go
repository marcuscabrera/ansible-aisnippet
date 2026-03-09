// Package cache provides an in-memory response cache with TTL expiry and LRU-style eviction.
package cache

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

type entry struct {
	value     string
	createdAt time.Time
	ttl       time.Duration
}

func (e *entry) isExpired() bool {
	return time.Since(e.createdAt) > e.ttl
}

// ResponseCache is a thread-safe in-memory cache for AI provider responses.
//
// Entries are keyed by a hash of (provider, systemMessage, userMessage) and
// expire after ttl. When the cache is full, expired entries are removed first;
// if still full the oldest live entry is evicted.
type ResponseCache struct {
	ttl     time.Duration
	maxSize int
	mu      sync.Mutex
	store   map[string]*entry
	order   []string // insertion order for LRU eviction
}

// New creates a ResponseCache with the given TTL (seconds) and maximum number of entries.
func New(ttlSeconds, maxSize int) *ResponseCache {
	return &ResponseCache{
		ttl:     time.Duration(ttlSeconds) * time.Second,
		maxSize: maxSize,
		store:   make(map[string]*entry),
	}
}

// Get returns the cached response or an empty string and false if missing or expired.
func (c *ResponseCache) Get(provider, systemMessage, userMessage string) (string, bool) {
	key := makeKey(provider, systemMessage, userMessage)
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.store[key]
	if !ok {
		return "", false
	}
	if e.isExpired() {
		c.deleteKey(key)
		return "", false
	}
	return e.value, true
}

// Set stores a response in the cache, evicting stale or oldest entries if full.
func (c *ResponseCache) Set(provider, systemMessage, userMessage, value string) {
	key := makeKey(provider, systemMessage, userMessage)
	c.mu.Lock()
	defer c.mu.Unlock()

	c.evictIfNeeded()

	if _, exists := c.store[key]; !exists {
		c.order = append(c.order, key)
	}
	c.store[key] = &entry{
		value:     value,
		createdAt: time.Now(),
		ttl:       c.ttl,
	}
}

// Clear removes all cached entries.
func (c *ResponseCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store = make(map[string]*entry)
	c.order = nil
}

// Len returns the number of entries currently in the cache.
func (c *ResponseCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.store)
}

// makeKey produces a SHA-256 hash key from the provider and messages.
func makeKey(provider, systemMessage, userMessage string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s:%s:%s", provider, systemMessage, userMessage)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// deleteKey removes a single entry (must be called with mu held).
func (c *ResponseCache) deleteKey(key string) {
	delete(c.store, key)
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
}

// evictIfNeeded removes expired entries and, if still needed, the oldest live entry.
// Must be called with mu held.
func (c *ResponseCache) evictIfNeeded() {
	if len(c.store) < c.maxSize {
		return
	}
	// Remove all expired entries first.
	for key, e := range c.store {
		if e.isExpired() {
			c.deleteKey(key)
		}
	}
	// If still full, remove the oldest live entry (first in insertion order).
	if len(c.store) >= c.maxSize && len(c.order) > 0 {
		oldest := c.order[0]
		c.deleteKey(oldest)
	}
}
