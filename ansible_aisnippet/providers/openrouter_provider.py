"""OpenRouter provider – routes requests to many AI models via a single API."""
import os

import requests

from .base import BaseProvider


class OpenRouterProvider(BaseProvider):
    """OpenRouter provider for accessing many AI models through one API key.

    OpenRouter (openrouter.ai) is an OpenAI-compatible aggregator that proxies
    requests to models from OpenAI, Anthropic, Meta, Mistral, and many others.

    Set OPENROUTER_API_KEY to your key from https://openrouter.ai/keys.
    Set OPENROUTER_MODEL to choose the model (e.g. "openai/gpt-4o",
    "anthropic/claude-3-haiku", "meta-llama/llama-3.1-8b-instruct:free").
    """

    name = "openrouter"
    DEFAULT_BASE_URL = "https://openrouter.ai/api/v1"
    DEFAULT_MODEL = "openai/gpt-3.5-turbo"

    def __init__(self, config):
        super().__init__(config)
        self.api_key = config.api_key or os.getenv("OPENROUTER_API_KEY")
        self.base_url = config.base_url or os.getenv(
            "OPENROUTER_BASE_URL", self.DEFAULT_BASE_URL
        )
        self.model = config.model or os.getenv("OPENROUTER_MODEL", self.DEFAULT_MODEL)

    def validate_config(self) -> bool:
        return bool(self.api_key)

    def generate(self, system_message: str, user_message: str) -> str:
        url = f"{self.base_url}/chat/completions"
        headers = {
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json",
        }
        payload = {
            "model": self.model,
            "messages": [
                {"role": "system", "content": system_message},
                {"role": "user", "content": user_message},
            ],
            "temperature": self.config.temperature,
        }
        response = requests.post(
            url, json=payload, headers=headers, timeout=self.config.timeout
        )
        response.raise_for_status()
        return response.json()["choices"][0]["message"]["content"]
