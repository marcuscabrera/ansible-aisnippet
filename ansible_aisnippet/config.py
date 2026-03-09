"""Centralized configuration for ansible-aisnippet.

Supports loading from environment variables or a YAML file.

Environment variable hierarchy (first match wins):
  - Variables explicitly set by the caller
  - OS environment variables
  - Default values
"""
from __future__ import annotations

import os
from dataclasses import dataclass, field
from typing import List, Optional

from .helpers import load_yaml


@dataclass
class ProviderConfig:
    """Configuration for a single AI provider."""

    api_key: Optional[str] = None
    model: Optional[str] = None
    base_url: Optional[str] = None
    temperature: float = 0.0
    max_retries: int = 3
    timeout: int = 30
    extra: dict = field(default_factory=dict)


@dataclass
class CacheConfig:
    """Configuration for the response cache."""

    enabled: bool = True
    ttl: int = 3600
    max_size: int = 100


@dataclass
class RateLimitConfig:
    """Configuration for the rate limiter."""

    enabled: bool = True
    requests_per_minute: int = 60


@dataclass
class Config:
    """Top-level configuration object for ansible-aisnippet."""

    provider: str = "openai"
    fallback_providers: List[str] = field(default_factory=list)
    providers: dict = field(default_factory=dict)
    cache: CacheConfig = field(default_factory=CacheConfig)
    rate_limit: RateLimitConfig = field(default_factory=RateLimitConfig)

    def get_provider_config(self, provider_name: Optional[str] = None) -> ProviderConfig:
        """Return a :class:`ProviderConfig` for the given provider name.

        Falls back to :attr:`provider` if *provider_name* is ``None``.
        Provider-specific settings from :attr:`providers` are merged with
        defaults so callers always receive a fully-populated object.
        """
        name = provider_name or self.provider
        data = self.providers.get(name, {})
        return ProviderConfig(
            api_key=data.get("api_key"),
            model=data.get("model"),
            base_url=data.get("base_url"),
            temperature=data.get("temperature", 0.0),
            max_retries=data.get("max_retries", 3),
            timeout=data.get("timeout", 30),
            extra=data.get("extra", {}),
        )

    @classmethod
    def from_env(cls) -> "Config":
        """Build a :class:`Config` from OS environment variables.

        Supported variables:

        ===============================  =================================
        Variable                         Description
        ===============================  =================================
        ``AI_PROVIDER``                  Active provider (default openai)
        ``AI_FALLBACK_PROVIDERS``        Comma-separated fallback list
        ``AI_CACHE_ENABLED``             ``true`` / ``false``
        ``AI_CACHE_TTL``                 Seconds (int)
        ``AI_CACHE_MAX_SIZE``            Items (int)
        ``AI_RATE_LIMIT_ENABLED``        ``true`` / ``false``
        ``AI_RATE_LIMIT_RPM``            Requests per minute (int)
        ===============================  =================================
        """
        provider = os.getenv("AI_PROVIDER", "openai")
        fallback_raw = os.getenv("AI_FALLBACK_PROVIDERS", "")
        fallback_providers = [p.strip() for p in fallback_raw.split(",") if p.strip()]

        cache = CacheConfig(
            enabled=os.getenv("AI_CACHE_ENABLED", "true").lower() != "false",
            ttl=int(os.getenv("AI_CACHE_TTL", "3600")),
            max_size=int(os.getenv("AI_CACHE_MAX_SIZE", "100")),
        )
        rate_limit = RateLimitConfig(
            enabled=os.getenv("AI_RATE_LIMIT_ENABLED", "true").lower() != "false",
            requests_per_minute=int(os.getenv("AI_RATE_LIMIT_RPM", "60")),
        )
        return cls(
            provider=provider,
            fallback_providers=fallback_providers,
            cache=cache,
            rate_limit=rate_limit,
        )

    @classmethod
    def from_file(cls, path: str) -> "Config":
        """Load configuration from a YAML file.

        Expected YAML layout::

            provider: openai
            fallback_providers:
              - anthropic
              - ollama
            cache:
              enabled: true
              ttl: 3600
              max_size: 100
            rate_limit:
              enabled: true
              requests_per_minute: 60
            providers:
              openai:
                api_key: sk-...
                model: gpt-4
              anthropic:
                api_key: sk-ant-...
                model: claude-3-haiku-20240307
        """
        data = load_yaml(path)
        if not isinstance(data, dict):
            raise ValueError(f"Configuration file {path} must contain a YAML mapping.")

        cache_data = data.get("cache", {}) or {}
        rate_data = data.get("rate_limit", {}) or {}

        cache = CacheConfig(
            enabled=cache_data.get("enabled", True),
            ttl=cache_data.get("ttl", 3600),
            max_size=cache_data.get("max_size", 100),
        )
        rate_limit = RateLimitConfig(
            enabled=rate_data.get("enabled", True),
            requests_per_minute=rate_data.get("requests_per_minute", 60),
        )
        return cls(
            provider=data.get("provider", "openai"),
            fallback_providers=data.get("fallback_providers", []),
            providers=data.get("providers", {}),
            cache=cache,
            rate_limit=rate_limit,
        )
