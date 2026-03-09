"""Fallback manager – tries providers in order until one succeeds."""
from __future__ import annotations

from typing import List, Tuple

from .providers.base import BaseProvider


class FallbackManager:
    """Tries a list of providers in order and returns the first successful response.

    This implements a simple failover strategy: if the primary provider raises
    an exception the next provider in the list is tried, and so on.  All errors
    are collected and surfaced in a single :class:`RuntimeError` only when every
    provider has been exhausted.

    Args:
        providers: Ordered list of :class:`~ansible_aisnippet.providers.base.BaseProvider`
                   instances.  The first element is treated as the primary provider.

    Example::

        manager = FallbackManager([openai_provider, anthropic_provider, ollama_provider])
        response, used_provider = manager.generate(system_msg, user_msg)
    """

    def __init__(self, providers: List[BaseProvider]) -> None:
        if not providers:
            raise ValueError("FallbackManager requires at least one provider.")
        self.providers = providers

    def generate(
        self, system_message: str, user_message: str
    ) -> Tuple[str, str]:
        """Generate a response, falling back on provider errors.

        Returns:
            A tuple of ``(response_text, provider_name)`` where
            *provider_name* is the name of the provider that succeeded.

        Raises:
            RuntimeError: When all providers have failed.
        """
        errors: List[str] = []
        for provider in self.providers:
            try:
                result = provider.generate(system_message, user_message)
                return result, provider.name
            except Exception as exc:
                errors.append(f"{provider.name}: {exc}")

        raise RuntimeError(
            "All providers failed to generate a response. Errors:\n"
            + "\n".join(f"  - {e}" for e in errors)
        )
