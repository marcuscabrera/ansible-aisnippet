"""Ollama provider for locally-running open-source models."""
import os

import requests

from .base import BaseProvider


class OllamaProvider(BaseProvider):
    """Ollama provider for self-hosted LLMs (llama3, mistral, phi3, etc.).

    Requires Ollama to be running locally or at a custom URL.
    Set OLLAMA_BASE_URL to point at a remote Ollama instance.
    """

    name = "ollama"
    DEFAULT_MODEL = "llama3"
    DEFAULT_BASE_URL = "http://localhost:11434"

    def __init__(self, config):
        super().__init__(config)
        self.base_url = config.base_url or os.getenv(
            "OLLAMA_BASE_URL", self.DEFAULT_BASE_URL
        )
        self.model = config.model or os.getenv("OLLAMA_MODEL", self.DEFAULT_MODEL)

    def validate_config(self) -> bool:
        return bool(self.base_url)

    def generate(self, system_message: str, user_message: str) -> str:
        url = f"{self.base_url}/api/chat"
        payload = {
            "model": self.model,
            "messages": [
                {"role": "system", "content": system_message},
                {"role": "user", "content": user_message},
            ],
            "stream": False,
            "options": {"temperature": self.config.temperature},
        }
        response = requests.post(url, json=payload, timeout=self.config.timeout)
        response.raise_for_status()
        return response.json()["message"]["content"]
