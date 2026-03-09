"""Sliding-window rate limiter for AI provider requests."""
from __future__ import annotations

import threading
import time
from collections import deque


class RateLimiter:
    """Thread-safe sliding-window rate limiter.

    Tracks request timestamps within the last 60 seconds and blocks (sleeps)
    when the per-minute limit is reached.

    Args:
        requests_per_minute: Maximum number of requests allowed per 60-second
                             window (default 60).
    """

    _WINDOW_SECONDS = 60.0

    def __init__(self, requests_per_minute: int = 60) -> None:
        self.requests_per_minute = requests_per_minute
        self._timestamps: deque[float] = deque()
        self._lock = threading.Lock()

    def acquire(self) -> None:
        """Block the calling thread until a request slot is available."""
        with self._lock:
            now = time.time()
            self._drop_old(now)

            if len(self._timestamps) >= self.requests_per_minute:
                # Sleep until the oldest request leaves the window
                wait_time = self._WINDOW_SECONDS - (now - self._timestamps[0])
                if wait_time > 0:
                    time.sleep(wait_time)
                    now = time.time()
                    self._drop_old(now)

            self._timestamps.append(now)

    def _drop_old(self, now: float) -> None:
        """Remove timestamps older than the sliding window."""
        cutoff = now - self._WINDOW_SECONDS
        while self._timestamps and self._timestamps[0] < cutoff:
            self._timestamps.popleft()
