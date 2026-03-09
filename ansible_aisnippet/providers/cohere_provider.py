"""Cohere provider using the Chat API."""
import os

import requests

from .base import BaseProvider


class CohereProvider(BaseProvider):
    """Cohere provider (command, command-r, command-r-plus, etc.)."""

    name = "cohere"
    DEFAULT_MODEL = "command"
    API_URL = "https://api.cohere.ai/v1/chat"

    def __init__(self, config):
        super().__init__(config)
        self.api_key = config.api_key or os.getenv("COHERE_API_KEY")
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
            "preamble": system_message,
            "message": user_message,
            "temperature": self.config.temperature,
        }
        response = requests.post(
            self.API_URL,
            json=payload,
            headers=headers,
            timeout=self.config.timeout,
        )
        response.raise_for_status()
        return response.json()["text"]
