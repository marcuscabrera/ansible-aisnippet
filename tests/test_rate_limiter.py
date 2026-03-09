"""Tests for the sliding-window rate limiter."""
import time
import pytest

from ansible_aisnippet.rate_limiter import RateLimiter


class TestRateLimiter:
    def test_acquire_within_limit_does_not_sleep(self):
        limiter = RateLimiter(requests_per_minute=60)
        start = time.time()
        for _ in range(5):
            limiter.acquire()
        elapsed = time.time() - start
        # Should be nearly instant for 5 requests when limit is 60/min
        assert elapsed < 1.0

    def test_tracks_request_count(self):
        limiter = RateLimiter(requests_per_minute=10)
        for _ in range(5):
            limiter.acquire()
        assert len(limiter._timestamps) == 5

    def test_high_limit_never_blocks(self):
        limiter = RateLimiter(requests_per_minute=1000)
        start = time.time()
        for _ in range(50):
            limiter.acquire()
        assert time.time() - start < 0.5

    def test_drop_old_removes_stale_timestamps(self):
        limiter = RateLimiter(requests_per_minute=100)
        # Manually inject an old timestamp
        limiter._timestamps.append(time.time() - 120)
        limiter._drop_old(time.time())
        assert len(limiter._timestamps) == 0
