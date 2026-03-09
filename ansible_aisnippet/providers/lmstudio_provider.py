"""LM Studio provider – uses the OpenAI-compatible local server API."""
import os

import requests

from .base import BaseProvider


class LMStudioProvider(BaseProvider):
    """LM Studio local server provider (OpenAI-compatible REST API).

    Requires LM Studio to be running with the local server enabled.
    Default endpoint: http://localhost:1234/v1
    """

    name = "lmstudio"
    DEFAULT_BASE_URL = "http://localhost:1234/v1"
    DEFAULT_MODEL = "local-model"

    def __init__(self, config):
        super().__init__(config)
        self.base_url = config.base_url or os.getenv(
            "LMSTUDIO_BASE_URL", self.DEFAULT_BASE_URL
        )
        self.model = config.model or os.getenv("LMSTUDIO_MODEL", self.DEFAULT_MODEL)

    def validate_config(self) -> bool:
        return bool(self.base_url)

    def generate(self, system_message: str, user_message: str) -> str:
        url = f"{self.base_url}/chat/completions"
        headers = {"Content-Type": "application/json"}
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
