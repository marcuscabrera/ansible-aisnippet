"""Anthropic Claude provider using the Messages API."""
import os

import requests

from .base import BaseProvider


class AnthropicProvider(BaseProvider):
    """Anthropic Claude provider (claude-3-*, claude-2, etc.)."""

    name = "anthropic"
    DEFAULT_MODEL = "claude-3-haiku-20240307"
    API_URL = "https://api.anthropic.com/v1/messages"
    ANTHROPIC_VERSION = "2023-06-01"

    def __init__(self, config):
        super().__init__(config)
        self.api_key = config.api_key or os.getenv("ANTHROPIC_API_KEY")
        self.model = config.model or self.DEFAULT_MODEL

    def validate_config(self) -> bool:
        return bool(self.api_key)

    def generate(self, system_message: str, user_message: str) -> str:
        headers = {
            "x-api-key": self.api_key,
            "anthropic-version": self.ANTHROPIC_VERSION,
            "content-type": "application/json",
        }
        payload = {
            "model": self.model,
            "system": system_message,
            "messages": [{"role": "user", "content": user_message}],
            "max_tokens": 1024,
            "temperature": self.config.temperature,
        }
        response = requests.post(
            self.API_URL,
            json=payload,
            headers=headers,
            timeout=self.config.timeout,
        )
        response.raise_for_status()
        return response.json()["content"][0]["text"]
