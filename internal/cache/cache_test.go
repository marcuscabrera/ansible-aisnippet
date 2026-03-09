package cache_test

import (
	"testing"
	"time"

	"github.com/marcuscabrera/ansible-aisnippet/internal/cache"
)

func TestMissOnEmptyCache(t *testing.T) {
	c := cache.New(3600, 100)
	_, ok := c.Get("openai", "sys", "user")
	if ok {
		t.Error("expected cache miss on empty cache")
	}
}

func TestSetAndGet(t *testing.T) {
	c := cache.New(3600, 100)
	c.Set("openai", "sys", "user", "response")
	val, ok := c.Get("openai", "sys", "user")
	if !ok || val != "response" {
		t.Errorf("expected 'response', got %q (ok=%v)", val, ok)
	}
}

func TestDifferentKeysDontCollide(t *testing.T) {
	c := cache.New(3600, 100)
	c.Set("openai", "sys", "user1", "resp1")
	c.Set("openai", "sys", "user2", "resp2")

	v1, ok1 := c.Get("openai", "sys", "user1")
	v2, ok2 := c.Get("openai", "sys", "user2")

	if !ok1 || v1 != "resp1" {
		t.Errorf("expected resp1, got %q (ok=%v)", v1, ok1)
	}
	if !ok2 || v2 != "resp2" {
		t.Errorf("expected resp2, got %q (ok=%v)", v2, ok2)
	}
}

func TestDifferentProvidersDontCollide(t *testing.T) {
	c := cache.New(3600, 100)
	c.Set("openai", "sys", "user", "openai-resp")
	c.Set("anthropic", "sys", "user", "anthropic-resp")

	v1, ok1 := c.Get("openai", "sys", "user")
	v2, ok2 := c.Get("anthropic", "sys", "user")

	if !ok1 || v1 != "openai-resp" {
		t.Errorf("expected openai-resp, got %q (ok=%v)", v1, ok1)
	}
	if !ok2 || v2 != "anthropic-resp" {
		t.Errorf("expected anthropic-resp, got %q (ok=%v)", v2, ok2)
	}
}

func TestEntryExpiresAfterTTL(t *testing.T) {
	c := cache.New(1, 100) // 1-second TTL
	c.Set("openai", "sys", "user", "response")

	val, ok := c.Get("openai", "sys", "user")
	if !ok || val != "response" {
		t.Fatalf("expected cache hit before TTL, got ok=%v", ok)
	}

	time.Sleep(1100 * time.Millisecond)
	_, ok = c.Get("openai", "sys", "user")
	if ok {
		t.Error("expected cache miss after TTL expiry")
	}
}

func TestClearEmptiesCache(t *testing.T) {
	c := cache.New(3600, 100)
	c.Set("openai", "sys", "user", "response")
	c.Clear()
	if c.Len() != 0 {
		t.Errorf("expected empty cache after Clear, got %d entries", c.Len())
	}
	_, ok := c.Get("openai", "sys", "user")
	if ok {
		t.Error("expected cache miss after Clear")
	}
}

func TestEvictsOldestWhenFull(t *testing.T) {
	c := cache.New(3600, 2)
	c.Set("openai", "sys", "u1", "r1")
	c.Set("openai", "sys", "u2", "r2")
	c.Set("openai", "sys", "u3", "r3") // triggers eviction

	if c.Len() != 2 {
		t.Errorf("expected 2 entries after eviction, got %d", c.Len())
	}
	// The newest entry must always be present.
	v3, ok3 := c.Get("openai", "sys", "u3")
	if !ok3 || v3 != "r3" {
		t.Errorf("expected r3, got %q (ok=%v)", v3, ok3)
	}
}

func TestEvictsExpiredEntriesBeforeOldest(t *testing.T) {
	c := cache.New(1, 2) // 1-second TTL, max 2 entries
	c.Set("openai", "sys", "u1", "r1")
	time.Sleep(1100 * time.Millisecond) // u1 is now expired

	c.Set("openai", "sys", "u2", "r2")
	c.Set("openai", "sys", "u3", "r3") // should evict the expired u1, not u2

	v2, ok2 := c.Get("openai", "sys", "u2")
	v3, ok3 := c.Get("openai", "sys", "u3")

	if !ok2 || v2 != "r2" {
		t.Errorf("expected r2, got %q (ok=%v)", v2, ok2)
	}
	if !ok3 || v3 != "r3" {
		t.Errorf("expected r3, got %q (ok=%v)", v3, ok3)
	}
}

func TestLenTracksEntries(t *testing.T) {
	c := cache.New(3600, 100)
	if c.Len() != 0 {
		t.Errorf("expected 0 initially, got %d", c.Len())
	}
	c.Set("p", "s", "u", "r")
	if c.Len() != 1 {
		t.Errorf("expected 1 after Set, got %d", c.Len())
	}
}
