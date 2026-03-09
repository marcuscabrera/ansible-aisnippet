"""Tests for the fallback manager."""
import pytest
from unittest.mock import MagicMock

from ansible_aisnippet.fallback import FallbackManager
from ansible_aisnippet.providers.base import BaseProvider
from ansible_aisnippet.config import ProviderConfig


def _make_provider(name: str, response: str = None, raises: Exception = None):
    """Create a mock provider that either returns a value or raises."""

    _name = name

    class _Provider(BaseProvider):
        name = _name

        def generate(self, system_message: str, user_message: str) -> str:
            return ""

    provider = _Provider(ProviderConfig())
    if raises is not None:
        provider.generate = MagicMock(side_effect=raises)
    else:
        provider.generate = MagicMock(return_value=response)
    return provider


class TestFallbackManager:
    def test_empty_providers_raises(self):
        with pytest.raises(ValueError):
            FallbackManager([])

    def test_returns_first_provider_response(self):
        p1 = _make_provider("p1", response="hello from p1")
        p2 = _make_provider("p2", response="hello from p2")
        manager = FallbackManager([p1, p2])
        result, used = manager.generate("sys", "user")
        assert result == "hello from p1"
        assert used == "p1"
        p1.generate.assert_called_once()
        p2.generate.assert_not_called()

    def test_falls_back_on_first_provider_error(self):
        p1 = _make_provider("p1", raises=RuntimeError("API error"))
        p2 = _make_provider("p2", response="hello from p2")
        manager = FallbackManager([p1, p2])
        result, used = manager.generate("sys", "user")
        assert result == "hello from p2"
        assert used == "p2"

    def test_raises_runtime_error_when_all_fail(self):
        p1 = _make_provider("p1", raises=RuntimeError("err1"))
        p2 = _make_provider("p2", raises=RuntimeError("err2"))
        manager = FallbackManager([p1, p2])
        with pytest.raises(RuntimeError, match="All providers failed"):
            manager.generate("sys", "user")

    def test_error_message_includes_all_provider_errors(self):
        p1 = _make_provider("p1", raises=RuntimeError("boom1"))
        p2 = _make_provider("p2", raises=RuntimeError("boom2"))
        manager = FallbackManager([p1, p2])
        with pytest.raises(RuntimeError) as exc_info:
            manager.generate("sys", "user")
        assert "p1" in str(exc_info.value)
        assert "p2" in str(exc_info.value)

    def test_single_provider_success(self):
        p1 = _make_provider("p1", response="solo response")
        manager = FallbackManager([p1])
        result, used = manager.generate("sys", "user")
        assert result == "solo response"
        assert used == "p1"
