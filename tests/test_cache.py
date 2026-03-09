"""Tests for the response cache."""
import time
import pytest

from ansible_aisnippet.cache import ResponseCache


class TestResponseCache:
    def test_miss_on_empty_cache(self):
        cache = ResponseCache()
        assert cache.get("openai", "sys", "user") is None

    def test_set_and_get(self):
        cache = ResponseCache()
        cache.set("openai", "sys", "user", "response")
        assert cache.get("openai", "sys", "user") == "response"

    def test_different_keys_dont_collide(self):
        cache = ResponseCache()
        cache.set("openai", "sys", "user1", "resp1")
        cache.set("openai", "sys", "user2", "resp2")
        assert cache.get("openai", "sys", "user1") == "resp1"
        assert cache.get("openai", "sys", "user2") == "resp2"

    def test_different_providers_dont_collide(self):
        cache = ResponseCache()
        cache.set("openai", "sys", "user", "openai-resp")
        cache.set("anthropic", "sys", "user", "anthropic-resp")
        assert cache.get("openai", "sys", "user") == "openai-resp"
        assert cache.get("anthropic", "sys", "user") == "anthropic-resp"

    def test_entry_expires_after_ttl(self):
        cache = ResponseCache(ttl=1)
        cache.set("openai", "sys", "user", "response")
        assert cache.get("openai", "sys", "user") == "response"
        time.sleep(1.1)
        assert cache.get("openai", "sys", "user") is None

    def test_clear_empties_cache(self):
        cache = ResponseCache()
        cache.set("openai", "sys", "user", "response")
        cache.clear()
        assert len(cache) == 0
        assert cache.get("openai", "sys", "user") is None

    def test_evicts_oldest_when_full(self):
        cache = ResponseCache(ttl=3600, max_size=2)
        cache.set("openai", "sys", "u1", "r1")
        cache.set("openai", "sys", "u2", "r2")
        # Adding a third entry should evict one of the first two
        cache.set("openai", "sys", "u3", "r3")
        assert len(cache) == 2
        # The newest entry should always be present
        assert cache.get("openai", "sys", "u3") == "r3"

    def test_evicts_expired_entries_before_oldest(self):
        cache = ResponseCache(ttl=1, max_size=2)
        cache.set("openai", "sys", "u1", "r1")
        time.sleep(1.1)
        # u1 is now expired
        cache.set("openai", "sys", "u2", "r2")
        # Adding u3 should evict the expired u1, not u2
        cache.set("openai", "sys", "u3", "r3")
        assert cache.get("openai", "sys", "u2") == "r2"
        assert cache.get("openai", "sys", "u3") == "r3"

    def test_len_tracks_entries(self):
        cache = ResponseCache()
        assert len(cache) == 0
        cache.set("p", "s", "u", "r")
        assert len(cache) == 1
