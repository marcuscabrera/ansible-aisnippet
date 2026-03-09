"""Tests for the provider factory and base provider interface."""
import pytest
from unittest.mock import MagicMock, patch

from ansible_aisnippet.providers.factory import ProviderFactory
from ansible_aisnippet.providers.base import BaseProvider
from ansible_aisnippet.config import ProviderConfig


class ConcreteProvider(BaseProvider):
    """Minimal concrete provider for testing the base class."""

    name = "test"

    def generate(self, system_message: str, user_message: str) -> str:
        return f"response to: {user_message}"


class TestBaseProvider:
    def test_validate_config_default_true(self):
        cfg = ProviderConfig()
        provider = ConcreteProvider(cfg)
        assert provider.validate_config() is True

    def test_generate_returns_string(self):
        cfg = ProviderConfig()
        provider = ConcreteProvider(cfg)
        result = provider.generate("system", "user")
        assert isinstance(result, str)

    def test_config_stored(self):
        cfg = ProviderConfig(model="my-model", temperature=0.5)
        provider = ConcreteProvider(cfg)
        assert provider.config.model == "my-model"
        assert provider.config.temperature == 0.5


class TestProviderFactory:
    def test_list_providers_returns_all_ten(self):
        providers = ProviderFactory.list_providers()
        expected = {
            "openai", "anthropic", "google", "azure", "mistral",
            "cohere", "ollama", "lmstudio", "llama", "huggingface",
            "openrouter", "zen",
        }
        assert expected.issubset(set(providers))

    def test_create_openai_provider(self):
        cfg = ProviderConfig(api_key="test-key")
        provider = ProviderFactory.create("openai", cfg)
        assert provider.name == "openai"

    def test_create_anthropic_provider(self):
        cfg = ProviderConfig(api_key="test-key")
        provider = ProviderFactory.create("anthropic", cfg)
        assert provider.name == "anthropic"

    def test_create_google_provider(self):
        cfg = ProviderConfig(api_key="test-key")
        provider = ProviderFactory.create("google", cfg)
        assert provider.name == "google"

    def test_create_azure_provider(self):
        cfg = ProviderConfig(api_key="test-key", base_url="https://example.openai.azure.com/")
        provider = ProviderFactory.create("azure", cfg)
        assert provider.name == "azure"

    def test_create_mistral_provider(self):
        cfg = ProviderConfig(api_key="test-key")
        provider = ProviderFactory.create("mistral", cfg)
        assert provider.name == "mistral"

    def test_create_cohere_provider(self):
        cfg = ProviderConfig(api_key="test-key")
        provider = ProviderFactory.create("cohere", cfg)
        assert provider.name == "cohere"

    def test_create_ollama_provider(self):
        cfg = ProviderConfig()
        provider = ProviderFactory.create("ollama", cfg)
        assert provider.name == "ollama"

    def test_create_lmstudio_provider(self):
        cfg = ProviderConfig()
        provider = ProviderFactory.create("lmstudio", cfg)
        assert provider.name == "lmstudio"

    def test_create_llama_provider(self):
        cfg = ProviderConfig()
        provider = ProviderFactory.create("llama", cfg)
        assert provider.name == "llama"

    def test_create_huggingface_provider(self):
        cfg = ProviderConfig(api_key="hf-token")
        provider = ProviderFactory.create("huggingface", cfg)
        assert provider.name == "huggingface"

    def test_create_openrouter_provider(self):
        cfg = ProviderConfig(api_key="sk-or-key")
        provider = ProviderFactory.create("openrouter", cfg)
        assert provider.name == "openrouter"

    def test_create_zen_provider(self):
        cfg = ProviderConfig(api_key="zen-key")
        provider = ProviderFactory.create("zen", cfg)
        assert provider.name == "zen"

    def test_unknown_provider_raises_value_error(self):
        cfg = ProviderConfig()
        with pytest.raises(ValueError, match="Unknown provider"):
            ProviderFactory.create("nonexistent", cfg)

    def test_case_insensitive_provider_name(self):
        cfg = ProviderConfig(api_key="test-key")
        provider = ProviderFactory.create("OpenAI", cfg)
        assert provider.name == "openai"
