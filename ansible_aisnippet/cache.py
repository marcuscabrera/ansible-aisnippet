"""In-memory response cache with TTL expiry and LRU-style eviction."""
from __future__ import annotations

import hashlib
import time
from dataclasses import dataclass
from typing import Dict, Optional


@dataclass
class _CacheEntry:
    value: str
    created_at: float
    ttl: int

    def is_expired(self) -> bool:
        return time.time() - self.created_at > self.ttl


class ResponseCache:
    """Thread-safe in-memory cache for AI provider responses.

    Entries are keyed by a hash of ``(provider, system_message, user_message)``
    and expire after *ttl* seconds.  When the cache is full the oldest entry
    is evicted first; expired entries are cleaned up during every ``set``.

    Args:
        ttl:      Time-to-live for each entry in seconds (default 3 600).
        max_size: Maximum number of entries before eviction (default 100).
    """

    def __init__(self, ttl: int = 3600, max_size: int = 100) -> None:
        self.ttl = ttl
        self.max_size = max_size
        self._cache: Dict[str, _CacheEntry] = {}

    # ------------------------------------------------------------------
    # Public API
    # ------------------------------------------------------------------

    def get(
        self, provider: str, system_message: str, user_message: str
    ) -> Optional[str]:
        """Return cached response or ``None`` if missing / expired."""
        key = self._make_key(provider, system_message, user_message)
        entry = self._cache.get(key)
        if entry is None:
            return None
        if entry.is_expired():
            del self._cache[key]
            return None
        return entry.value

    def set(
        self, provider: str, system_message: str, user_message: str, value: str
    ) -> None:
        """Store *value* in the cache, evicting stale / oldest entries if full."""
        key = self._make_key(provider, system_message, user_message)
        self._evict_if_needed()
        self._cache[key] = _CacheEntry(
            value=value, created_at=time.time(), ttl=self.ttl
        )

    def clear(self) -> None:
        """Remove all cached entries."""
        self._cache.clear()

    def __len__(self) -> int:
        return len(self._cache)

    # ------------------------------------------------------------------
    # Private helpers
    # ------------------------------------------------------------------

    @staticmethod
    def _make_key(provider: str, system_message: str, user_message: str) -> str:
        content = f"{provider}:{system_message}:{user_message}"
        return hashlib.sha256(content.encode()).hexdigest()

    def _evict_if_needed(self) -> None:
        if len(self._cache) < self.max_size:
            return
        # First remove all expired entries
        expired = [k for k, v in self._cache.items() if v.is_expired()]
        for k in expired:
            del self._cache[k]
        # If still full, remove the oldest live entry
        if len(self._cache) >= self.max_size and self._cache:
            oldest_key = min(self._cache, key=lambda k: self._cache[k].created_at)
            del self._cache[oldest_key]
