"""Mistral AI provider using the chat completions REST API."""
import os

import requests

from .base import BaseProvider


class MistralProvider(BaseProvider):
    """Mistral AI provider (mistral-small, mistral-medium, mistral-large)."""

    name = "mistral"
    DEFAULT_MODEL = "mistral-small"
    API_URL = "https://api.mistral.ai/v1/chat/completions"

    def __init__(self, config):
        super().__init__(config)
        self.api_key = config.api_key or os.getenv("MISTRAL_API_KEY")
        self.model = config.model or self.DEFAULT_MODEL

    def validate_config(self) -> bool:
        return bool(self.api_key)

    def generate(self, system_message: str, user_message: str) -> str:
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
            self.API_URL,
            json=payload,
            headers=headers,
            timeout=self.config.timeout,
        )
        response.raise_for_status()
        return response.json()["choices"][0]["message"]["content"]
