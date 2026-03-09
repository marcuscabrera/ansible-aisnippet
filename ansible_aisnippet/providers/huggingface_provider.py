"""HuggingFace Inference API provider."""
import os

import requests

from .base import BaseProvider


class HuggingFaceProvider(BaseProvider):
    """HuggingFace Inference API provider.

    Supports Inference API (serverless) endpoints.
    Set HF_API_TOKEN and optionally HF_MODEL to choose the model.
    """

    name = "huggingface"
    DEFAULT_MODEL = "mistralai/Mistral-7B-Instruct-v0.2"
    API_URL = "https://api-inference.huggingface.co/models/{model}"

    def __init__(self, config):
        super().__init__(config)
        self.api_key = config.api_key or os.getenv("HF_API_TOKEN")
        self.model = config.model or os.getenv("HF_MODEL", self.DEFAULT_MODEL)

    def validate_config(self) -> bool:
        return bool(self.api_key)

    def generate(self, system_message: str, user_message: str) -> str:
        url = self.API_URL.format(model=self.model)
        headers = {
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json",
        }
        # Instruction-following format compatible with most chat-tuned models
        prompt = f"[INST] {system_message}\n\n{user_message} [/INST]"
        payload = {
            "inputs": prompt,
            "parameters": {
                # HuggingFace API rejects temperature=0; use a small positive value
                "temperature": max(0.01, self.config.temperature),
                "max_new_tokens": 1024,
                "return_full_text": False,
            },
        }
        response = requests.post(
            url, json=payload, headers=headers, timeout=self.config.timeout
        )
        response.raise_for_status()
        data = response.json()
        # Response is a list of generated texts
        if isinstance(data, list):
            return data[0].get("generated_text", "")
        return data.get("generated_text", "")
