"""Tests for the centralized configuration module."""
import os
import pytest
from unittest.mock import patch

from ansible_aisnippet.config import (
    CacheConfig,
    Config,
    ProviderConfig,
    RateLimitConfig,
)


class TestProviderConfig:
    def test_defaults(self):
        cfg = ProviderConfig()
        assert cfg.api_key is None
        assert cfg.model is None
        assert cfg.base_url is None
        assert cfg.temperature == 0.0
        assert cfg.max_retries == 3
        assert cfg.timeout == 30
        assert cfg.extra == {}

    def test_custom_values(self):
        cfg = ProviderConfig(api_key="key", model="gpt-4", temperature=0.7)
        assert cfg.api_key == "key"
        assert cfg.model == "gpt-4"
        assert cfg.temperature == 0.7


class TestConfig:
    def test_defaults(self):
        cfg = Config()
        assert cfg.provider == "openai"
        assert cfg.fallback_providers == []
        assert isinstance(cfg.cache, CacheConfig)
        assert isinstance(cfg.rate_limit, RateLimitConfig)

    def test_from_env_defaults(self):
        with patch.dict(os.environ, {}, clear=False):
            # Remove relevant keys if present
            env_copy = {
                k: v for k, v in os.environ.items()
                if not k.startswith("AI_")
            }
            with patch.dict(os.environ, env_copy, clear=True):
                cfg = Config.from_env()
        assert cfg.provider == "openai"
        assert cfg.fallback_providers == []

    def test_from_env_custom_provider(self):
        with patch.dict(os.environ, {"AI_PROVIDER": "anthropic"}):
            cfg = Config.from_env()
        assert cfg.provider == "anthropic"

    def test_from_env_fallback_providers(self):
        with patch.dict(os.environ, {"AI_FALLBACK_PROVIDERS": "ollama, lmstudio"}):
            cfg = Config.from_env()
        assert cfg.fallback_providers == ["ollama", "lmstudio"]

    def test_from_env_cache_disabled(self):
        with patch.dict(os.environ, {"AI_CACHE_ENABLED": "false"}):
            cfg = Config.from_env()
        assert cfg.cache.enabled is False

    def test_from_env_rate_limit_rpm(self):
        with patch.dict(os.environ, {"AI_RATE_LIMIT_RPM": "30"}):
            cfg = Config.from_env()
        assert cfg.rate_limit.requests_per_minute == 30

    def test_get_provider_config_defaults(self):
        cfg = Config()
        pcfg = cfg.get_provider_config("openai")
        assert isinstance(pcfg, ProviderConfig)
        assert pcfg.api_key is None

    def test_get_provider_config_from_providers_dict(self):
        cfg = Config(
            providers={
                "openai": {"api_key": "sk-test", "model": "gpt-4", "temperature": 0.5}
            }
        )
        pcfg = cfg.get_provider_config("openai")
        assert pcfg.api_key == "sk-test"
        assert pcfg.model == "gpt-4"
        assert pcfg.temperature == 0.5

    def test_get_provider_config_uses_active_provider_when_name_omitted(self):
        cfg = Config(
            provider="mistral",
            providers={"mistral": {"api_key": "key-m"}},
        )
        pcfg = cfg.get_provider_config()
        assert pcfg.api_key == "key-m"

    def test_from_file(self, tmp_path):
        config_file = tmp_path / "config.yml"
        config_file.write_text(
            "provider: anthropic\n"
            "fallback_providers:\n"
            "  - ollama\n"
            "cache:\n"
            "  enabled: false\n"
            "  ttl: 60\n"
            "rate_limit:\n"
            "  requests_per_minute: 10\n"
            "providers:\n"
            "  anthropic:\n"
            "    api_key: ant-key\n"
            "    model: claude-3-haiku-20240307\n"
        )
        cfg = Config.from_file(str(config_file))
        assert cfg.provider == "anthropic"
        assert cfg.fallback_providers == ["ollama"]
        assert cfg.cache.enabled is False
        assert cfg.cache.ttl == 60
        assert cfg.rate_limit.requests_per_minute == 10
        pcfg = cfg.get_provider_config("anthropic")
        assert pcfg.api_key == "ant-key"
        assert pcfg.model == "claude-3-haiku-20240307"

    def test_from_file_invalid_raises(self, tmp_path):
        config_file = tmp_path / "bad.yml"
        config_file.write_text("- item1\n- item2\n")
        with pytest.raises(ValueError):
            Config.from_file(str(config_file))
