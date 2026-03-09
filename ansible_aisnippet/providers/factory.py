"""Provider Factory – creates AI provider instances by name (Factory Pattern)."""
from __future__ import annotations

from typing import TYPE_CHECKING

from .base import BaseProvider

if TYPE_CHECKING:
    from ..config import ProviderConfig

_REGISTRY: dict[str, type[BaseProvider]] = {}


def _register(provider_class: type[BaseProvider]) -> type[BaseProvider]:
    """Register a provider class under its ``name`` attribute."""
    _REGISTRY[provider_class.name] = provider_class
    return provider_class


def _lazy_register_all() -> None:
    """Import and register all built-in providers (deferred to first use)."""
    if _REGISTRY:
        return

    from .anthropic_provider import AnthropicProvider
    from .azure_provider import AzureOpenAIProvider
    from .cohere_provider import CohereProvider
    from .google_provider import GoogleGeminiProvider
    from .huggingface_provider import HuggingFaceProvider
    from .llama_provider import MetaLlamaProvider
    from .lmstudio_provider import LMStudioProvider
    from .mistral_provider import MistralProvider
    from .ollama_provider import OllamaProvider
    from .openai_provider import OpenAIProvider
    from .openrouter_provider import OpenRouterProvider
    from .zen_provider import ZenProvider

    for cls in [
        OpenAIProvider,
        AnthropicProvider,
        GoogleGeminiProvider,
        AzureOpenAIProvider,
        MistralProvider,
        CohereProvider,
        OllamaProvider,
        LMStudioProvider,
        MetaLlamaProvider,
        HuggingFaceProvider,
        OpenRouterProvider,
        ZenProvider,
    ]:
        _register(cls)


class ProviderFactory:
    """Factory for creating AI provider instances.

    Usage::

        from ansible_aisnippet.providers.factory import ProviderFactory
        from ansible_aisnippet.config import ProviderConfig

        cfg = ProviderConfig(api_key="sk-...", model="gpt-4")
        provider = ProviderFactory.create("openai", cfg)
        response = provider.generate(system_msg, user_msg)
    """

    @staticmethod
    def create(provider_name: str, config: "ProviderConfig") -> BaseProvider:
        """Instantiate a provider by name.

        Args:
            provider_name: One of the supported provider identifiers
                (``openai``, ``anthropic``, ``google``, ``azure``,
                ``mistral``, ``cohere``, ``ollama``, ``lmstudio``,
                ``llama``, ``huggingface``, ``openrouter``, ``zen``).
            config: A :class:`~ansible_aisnippet.config.ProviderConfig`
                containing API key, model, and other settings.

        Returns:
            A :class:`BaseProvider` instance ready to call.

        Raises:
            ValueError: If *provider_name* is not registered.
        """
        _lazy_register_all()
        name_lower = provider_name.lower().strip()
        provider_class = _REGISTRY.get(name_lower)
        if provider_class is None:
            available = ", ".join(sorted(_REGISTRY.keys()))
            raise ValueError(
                f"Unknown provider '{provider_name}'. "
                f"Available providers: {available}"
            )
        return provider_class(config)

    @staticmethod
    def list_providers() -> list[str]:
        """Return a sorted list of all registered provider names."""
        _lazy_register_all()
        return sorted(_REGISTRY.keys())
