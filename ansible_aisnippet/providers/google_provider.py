"""Google Gemini provider using the Generative Language REST API."""
import os

import requests

from .base import BaseProvider


class GoogleGeminiProvider(BaseProvider):
    """Google Gemini provider (gemini-pro, gemini-1.5-pro, etc.)."""

    name = "google"
    DEFAULT_MODEL = "gemini-pro"
    API_URL = (
        "https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent"
    )

    def __init__(self, config):
        super().__init__(config)
        self.api_key = config.api_key or os.getenv("GOOGLE_API_KEY")
        self.model = config.model or self.DEFAULT_MODEL

    def validate_config(self) -> bool:
        return bool(self.api_key)

    def generate(self, system_message: str, user_message: str) -> str:
        url = self.API_URL.format(model=self.model)
        headers = {"Content-Type": "application/json"}
        payload = {
            "contents": [
                {
                    "role": "user",
                    "parts": [{"text": f"{system_message}\n\n{user_message}"}],
                }
            ],
            "generationConfig": {
                "temperature": self.config.temperature,
                "maxOutputTokens": 1024,
            },
        }
        response = requests.post(
            f"{url}?key={self.api_key}",
            json=payload,
            headers=headers,
            timeout=self.config.timeout,
        )
        response.raise_for_status()
        return response.json()["candidates"][0]["content"]["parts"][0]["text"]
