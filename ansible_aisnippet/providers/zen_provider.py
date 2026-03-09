"""OpenCode Zen provider – OpenAI-compatible interface for Zen AI models."""
import os

import requests

from .base import BaseProvider


class ZenProvider(BaseProvider):
    """OpenCode Zen provider for Zen AI (opencode.ai).

    Zen exposes an OpenAI-compatible REST API. By default the provider
    connects to the Zen cloud endpoint, but it can also be pointed at a
    self-hosted Zen instance by setting ZEN_BASE_URL.

    Set ZEN_API_KEY to your Zen API key.
    Set ZEN_MODEL to choose the model (default: ``zen``).
    Set ZEN_BASE_URL to override the API endpoint.
    """

    name = "zen"
    DEFAULT_BASE_URL = "https://api.opencode.ai/v1"
    DEFAULT_MODEL = "zen"

    def __init__(self, config):
        super().__init__(config)
        self.api_key = config.api_key or os.getenv("ZEN_API_KEY")
        self.base_url = config.base_url or os.getenv(
            "ZEN_BASE_URL", self.DEFAULT_BASE_URL
        )
        self.model = config.model or os.getenv("ZEN_MODEL", self.DEFAULT_MODEL)

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
