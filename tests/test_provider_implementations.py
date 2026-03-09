"""Tests for individual AI provider implementations."""
import os
import pytest
from unittest.mock import MagicMock, patch

from ansible_aisnippet.config import ProviderConfig


class TestOpenAIProvider:
    def test_api_key_from_config(self):
        from ansible_aisnippet.providers.openai_provider import OpenAIProvider
        cfg = ProviderConfig(api_key="cfg-key")
        provider = OpenAIProvider(cfg)
        assert provider.api_key == "cfg-key"

    def test_api_key_from_env(self):
        from ansible_aisnippet.providers.openai_provider import OpenAIProvider
        with patch.dict(os.environ, {"OPENAI_KEY": "env-key"}):
            cfg = ProviderConfig()
            provider = OpenAIProvider(cfg)
        assert provider.api_key == "env-key"

    def test_model_default(self):
        from ansible_aisnippet.providers.openai_provider import OpenAIProvider
        cfg = ProviderConfig()
        provider = OpenAIProvider(cfg)
        assert provider.model == OpenAIProvider.DEFAULT_MODEL

    def test_model_from_config(self):
        from ansible_aisnippet.providers.openai_provider import OpenAIProvider
        cfg = ProviderConfig(model="gpt-4")
        provider = OpenAIProvider(cfg)
        assert provider.model == "gpt-4"

    def test_validate_config_with_key(self):
        from ansible_aisnippet.providers.openai_provider import OpenAIProvider
        cfg = ProviderConfig(api_key="key")
        assert OpenAIProvider(cfg).validate_config() is True

    def test_validate_config_without_key(self):
        from ansible_aisnippet.providers.openai_provider import OpenAIProvider
        with patch.dict(os.environ, {}, clear=True):
            cfg = ProviderConfig()
            provider = OpenAIProvider(cfg)
        assert provider.validate_config() is False

    def test_generate_legacy_api(self):
        """Test generation with legacy openai < 1.0 API."""
        from ansible_aisnippet.providers.openai_provider import OpenAIProvider

        mock_openai = MagicMock()
        # Simulate old API: no OpenAI class
        del mock_openai.OpenAI
        mock_response = {"choices": [{"message": {"content": "task_json"}}]}
        mock_openai.ChatCompletion.create.return_value = mock_response

        cfg = ProviderConfig(api_key="key")
        provider = OpenAIProvider(cfg)

        with patch.dict("sys.modules", {"openai": mock_openai}):
            result = provider.generate("sys", "user")

        assert result == "task_json"

    def test_generate_new_api(self):
        """Test generation with new openai >= 1.0 API."""
        from ansible_aisnippet.providers.openai_provider import OpenAIProvider

        mock_response = MagicMock()
        mock_response.choices[0].message.content = "task_json_new"

        mock_client = MagicMock()
        mock_client.chat.completions.create.return_value = mock_response
        mock_openai_module = MagicMock()
        mock_openai_module.OpenAI.return_value = mock_client

        cfg = ProviderConfig(api_key="key")
        provider = OpenAIProvider(cfg)

        with patch.dict("sys.modules", {"openai": mock_openai_module}):
            result = provider.generate("sys", "user")

        assert result == "task_json_new"


class TestAnthropicProvider:
    def test_api_key_from_env(self):
        from ansible_aisnippet.providers.anthropic_provider import AnthropicProvider
        with patch.dict(os.environ, {"ANTHROPIC_API_KEY": "ant-key"}):
            provider = AnthropicProvider(ProviderConfig())
        assert provider.api_key == "ant-key"

    def test_default_model(self):
        from ansible_aisnippet.providers.anthropic_provider import AnthropicProvider
        provider = AnthropicProvider(ProviderConfig())
        assert provider.model == AnthropicProvider.DEFAULT_MODEL

    def test_validate_config_requires_key(self):
        from ansible_aisnippet.providers.anthropic_provider import AnthropicProvider
        with patch.dict(os.environ, {}, clear=True):
            provider = AnthropicProvider(ProviderConfig())
        assert provider.validate_config() is False

    def test_generate_calls_api(self):
        from ansible_aisnippet.providers.anthropic_provider import AnthropicProvider
        mock_response = MagicMock()
        mock_response.json.return_value = {
            "content": [{"text": "anthropic response"}]
        }
        mock_response.raise_for_status = MagicMock()

        cfg = ProviderConfig(api_key="key")
        provider = AnthropicProvider(cfg)

        with patch("ansible_aisnippet.providers.anthropic_provider.requests.post") as mock_post:
            mock_post.return_value = mock_response
            result = provider.generate("sys", "user")

        assert result == "anthropic response"
        mock_post.assert_called_once()


class TestGoogleGeminiProvider:
    def test_api_key_from_env(self):
        from ansible_aisnippet.providers.google_provider import GoogleGeminiProvider
        with patch.dict(os.environ, {"GOOGLE_API_KEY": "g-key"}):
            provider = GoogleGeminiProvider(ProviderConfig())
        assert provider.api_key == "g-key"

    def test_generate_calls_api(self):
        from ansible_aisnippet.providers.google_provider import GoogleGeminiProvider
        mock_response = MagicMock()
        mock_response.json.return_value = {
            "candidates": [{"content": {"parts": [{"text": "gemini response"}]}}]
        }
        mock_response.raise_for_status = MagicMock()

        cfg = ProviderConfig(api_key="key")
        provider = GoogleGeminiProvider(cfg)

        with patch("ansible_aisnippet.providers.google_provider.requests.post") as mock_post:
            mock_post.return_value = mock_response
            result = provider.generate("sys", "user")

        assert result == "gemini response"


class TestMistralProvider:
    def test_api_key_from_env(self):
        from ansible_aisnippet.providers.mistral_provider import MistralProvider
        with patch.dict(os.environ, {"MISTRAL_API_KEY": "mist-key"}):
            provider = MistralProvider(ProviderConfig())
        assert provider.api_key == "mist-key"

    def test_generate_calls_api(self):
        from ansible_aisnippet.providers.mistral_provider import MistralProvider
        mock_response = MagicMock()
        mock_response.json.return_value = {
            "choices": [{"message": {"content": "mistral response"}}]
        }
        mock_response.raise_for_status = MagicMock()

        cfg = ProviderConfig(api_key="key")
        provider = MistralProvider(cfg)

        with patch("ansible_aisnippet.providers.mistral_provider.requests.post") as mock_post:
            mock_post.return_value = mock_response
            result = provider.generate("sys", "user")

        assert result == "mistral response"


class TestCohereProvider:
    def test_generate_calls_api(self):
        from ansible_aisnippet.providers.cohere_provider import CohereProvider
        mock_response = MagicMock()
        mock_response.json.return_value = {"text": "cohere response"}
        mock_response.raise_for_status = MagicMock()

        cfg = ProviderConfig(api_key="key")
        provider = CohereProvider(cfg)

        with patch("ansible_aisnippet.providers.cohere_provider.requests.post") as mock_post:
            mock_post.return_value = mock_response
            result = provider.generate("sys", "user")

        assert result == "cohere response"


class TestOllamaProvider:
    def test_default_base_url(self):
        from ansible_aisnippet.providers.ollama_provider import OllamaProvider
        with patch.dict(os.environ, {}, clear=True):
            provider = OllamaProvider(ProviderConfig())
        assert provider.base_url == OllamaProvider.DEFAULT_BASE_URL

    def test_base_url_from_env(self):
        from ansible_aisnippet.providers.ollama_provider import OllamaProvider
        with patch.dict(os.environ, {"OLLAMA_BASE_URL": "http://remote:11434"}):
            provider = OllamaProvider(ProviderConfig())
        assert provider.base_url == "http://remote:11434"

    def test_generate_calls_api(self):
        from ansible_aisnippet.providers.ollama_provider import OllamaProvider
        mock_response = MagicMock()
        mock_response.json.return_value = {"message": {"content": "ollama response"}}
        mock_response.raise_for_status = MagicMock()

        provider = OllamaProvider(ProviderConfig())

        with patch("ansible_aisnippet.providers.ollama_provider.requests.post") as mock_post:
            mock_post.return_value = mock_response
            result = provider.generate("sys", "user")

        assert result == "ollama response"


class TestLMStudioProvider:
    def test_generate_calls_api(self):
        from ansible_aisnippet.providers.lmstudio_provider import LMStudioProvider
        mock_response = MagicMock()
        mock_response.json.return_value = {
            "choices": [{"message": {"content": "lmstudio response"}}]
        }
        mock_response.raise_for_status = MagicMock()

        provider = LMStudioProvider(ProviderConfig())

        with patch("ansible_aisnippet.providers.lmstudio_provider.requests.post") as mock_post:
            mock_post.return_value = mock_response
            result = provider.generate("sys", "user")

        assert result == "lmstudio response"


class TestMetaLlamaProvider:
    def test_default_model(self):
        from ansible_aisnippet.providers.llama_provider import MetaLlamaProvider
        provider = MetaLlamaProvider(ProviderConfig())
        assert provider.model == MetaLlamaProvider.DEFAULT_MODEL

    def test_generate_calls_api(self):
        from ansible_aisnippet.providers.llama_provider import MetaLlamaProvider
        mock_response = MagicMock()
        mock_response.json.return_value = {"message": {"content": "llama response"}}
        mock_response.raise_for_status = MagicMock()

        provider = MetaLlamaProvider(ProviderConfig())

        with patch("ansible_aisnippet.providers.llama_provider.requests.post") as mock_post:
            mock_post.return_value = mock_response
            result = provider.generate("sys", "user")

        assert result == "llama response"


class TestHuggingFaceProvider:
    def test_api_key_from_env(self):
        from ansible_aisnippet.providers.huggingface_provider import HuggingFaceProvider
        with patch.dict(os.environ, {"HF_API_TOKEN": "hf-key"}):
            provider = HuggingFaceProvider(ProviderConfig())
        assert provider.api_key == "hf-key"

    def test_generate_list_response(self):
        from ansible_aisnippet.providers.huggingface_provider import HuggingFaceProvider
        mock_response = MagicMock()
        mock_response.json.return_value = [{"generated_text": "hf response"}]
        mock_response.raise_for_status = MagicMock()

        cfg = ProviderConfig(api_key="key")
        provider = HuggingFaceProvider(cfg)

        with patch("ansible_aisnippet.providers.huggingface_provider.requests.post") as mock_post:
            mock_post.return_value = mock_response
            result = provider.generate("sys", "user")

        assert result == "hf response"


class TestOpenRouterProvider:
    def test_api_key_from_config(self):
        from ansible_aisnippet.providers.openrouter_provider import OpenRouterProvider
        cfg = ProviderConfig(api_key="sk-or-key")
        provider = OpenRouterProvider(cfg)
        assert provider.api_key == "sk-or-key"

    def test_api_key_from_env(self):
        from ansible_aisnippet.providers.openrouter_provider import OpenRouterProvider
        with patch.dict(os.environ, {"OPENROUTER_API_KEY": "env-or-key"}):
            provider = OpenRouterProvider(ProviderConfig())
        assert provider.api_key == "env-or-key"

    def test_default_base_url(self):
        from ansible_aisnippet.providers.openrouter_provider import OpenRouterProvider
        with patch.dict(os.environ, {}, clear=True):
            provider = OpenRouterProvider(ProviderConfig())
        assert provider.base_url == OpenRouterProvider.DEFAULT_BASE_URL

    def test_base_url_from_env(self):
        from ansible_aisnippet.providers.openrouter_provider import OpenRouterProvider
        with patch.dict(os.environ, {"OPENROUTER_BASE_URL": "https://custom.openrouter.ai/v1"}):
            provider = OpenRouterProvider(ProviderConfig())
        assert provider.base_url == "https://custom.openrouter.ai/v1"

    def test_default_model(self):
        from ansible_aisnippet.providers.openrouter_provider import OpenRouterProvider
        with patch.dict(os.environ, {}, clear=True):
            provider = OpenRouterProvider(ProviderConfig())
        assert provider.model == OpenRouterProvider.DEFAULT_MODEL

    def test_model_from_config(self):
        from ansible_aisnippet.providers.openrouter_provider import OpenRouterProvider
        cfg = ProviderConfig(model="anthropic/claude-3-haiku")
        provider = OpenRouterProvider(cfg)
        assert provider.model == "anthropic/claude-3-haiku"

    def test_validate_config_with_key(self):
        from ansible_aisnippet.providers.openrouter_provider import OpenRouterProvider
        cfg = ProviderConfig(api_key="key")
        assert OpenRouterProvider(cfg).validate_config() is True

    def test_validate_config_without_key(self):
        from ansible_aisnippet.providers.openrouter_provider import OpenRouterProvider
        with patch.dict(os.environ, {}, clear=True):
            provider = OpenRouterProvider(ProviderConfig())
        assert provider.validate_config() is False

    def test_generate_calls_api(self):
        from ansible_aisnippet.providers.openrouter_provider import OpenRouterProvider
        mock_response = MagicMock()
        mock_response.json.return_value = {
            "choices": [{"message": {"content": "openrouter response"}}]
        }
        mock_response.raise_for_status = MagicMock()

        cfg = ProviderConfig(api_key="key")
        provider = OpenRouterProvider(cfg)

        with patch("ansible_aisnippet.providers.openrouter_provider.requests.post") as mock_post:
            mock_post.return_value = mock_response
            result = provider.generate("sys", "user")

        assert result == "openrouter response"
        mock_post.assert_called_once()

    def test_generate_sends_auth_header(self):
        from ansible_aisnippet.providers.openrouter_provider import OpenRouterProvider
        mock_response = MagicMock()
        mock_response.json.return_value = {
            "choices": [{"message": {"content": "ok"}}]
        }
        mock_response.raise_for_status = MagicMock()

        cfg = ProviderConfig(api_key="my-or-key")
        provider = OpenRouterProvider(cfg)

        with patch("ansible_aisnippet.providers.openrouter_provider.requests.post") as mock_post:
            mock_post.return_value = mock_response
            provider.generate("sys", "user")

        _, kwargs = mock_post.call_args
        assert kwargs["headers"]["Authorization"] == "Bearer my-or-key"


class TestZenProvider:
    def test_api_key_from_config(self):
        from ansible_aisnippet.providers.zen_provider import ZenProvider
        cfg = ProviderConfig(api_key="zen-key")
        provider = ZenProvider(cfg)
        assert provider.api_key == "zen-key"

    def test_api_key_from_env(self):
        from ansible_aisnippet.providers.zen_provider import ZenProvider
        with patch.dict(os.environ, {"ZEN_API_KEY": "env-zen-key"}):
            provider = ZenProvider(ProviderConfig())
        assert provider.api_key == "env-zen-key"

    def test_default_base_url(self):
        from ansible_aisnippet.providers.zen_provider import ZenProvider
        with patch.dict(os.environ, {}, clear=True):
            provider = ZenProvider(ProviderConfig())
        assert provider.base_url == ZenProvider.DEFAULT_BASE_URL

    def test_base_url_from_env(self):
        from ansible_aisnippet.providers.zen_provider import ZenProvider
        with patch.dict(os.environ, {"ZEN_BASE_URL": "http://localhost:8080/v1"}):
            provider = ZenProvider(ProviderConfig())
        assert provider.base_url == "http://localhost:8080/v1"

    def test_default_model(self):
        from ansible_aisnippet.providers.zen_provider import ZenProvider
        with patch.dict(os.environ, {}, clear=True):
            provider = ZenProvider(ProviderConfig())
        assert provider.model == ZenProvider.DEFAULT_MODEL

    def test_model_from_config(self):
        from ansible_aisnippet.providers.zen_provider import ZenProvider
        cfg = ProviderConfig(model="zen-pro")
        provider = ZenProvider(cfg)
        assert provider.model == "zen-pro"

    def test_validate_config_with_key(self):
        from ansible_aisnippet.providers.zen_provider import ZenProvider
        cfg = ProviderConfig(api_key="key")
        assert ZenProvider(cfg).validate_config() is True

    def test_validate_config_without_key(self):
        from ansible_aisnippet.providers.zen_provider import ZenProvider
        with patch.dict(os.environ, {}, clear=True):
            provider = ZenProvider(ProviderConfig())
        assert provider.validate_config() is False

    def test_generate_calls_api(self):
        from ansible_aisnippet.providers.zen_provider import ZenProvider
        mock_response = MagicMock()
        mock_response.json.return_value = {
            "choices": [{"message": {"content": "zen response"}}]
        }
        mock_response.raise_for_status = MagicMock()

        cfg = ProviderConfig(api_key="key")
        provider = ZenProvider(cfg)

        with patch("ansible_aisnippet.providers.zen_provider.requests.post") as mock_post:
            mock_post.return_value = mock_response
            result = provider.generate("sys", "user")

        assert result == "zen response"
        mock_post.assert_called_once()

    def test_generate_sends_auth_header(self):
        from ansible_aisnippet.providers.zen_provider import ZenProvider
        mock_response = MagicMock()
        mock_response.json.return_value = {
            "choices": [{"message": {"content": "ok"}}]
        }
        mock_response.raise_for_status = MagicMock()

        cfg = ProviderConfig(api_key="my-zen-key")
        provider = ZenProvider(cfg)

        with patch("ansible_aisnippet.providers.zen_provider.requests.post") as mock_post:
            mock_post.return_value = mock_response
            provider.generate("sys", "user")

        _, kwargs = mock_post.call_args
        assert kwargs["headers"]["Authorization"] == "Bearer my-zen-key"
