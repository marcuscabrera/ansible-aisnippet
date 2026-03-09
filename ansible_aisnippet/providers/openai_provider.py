"""OpenAI provider (supports gpt-3.5-turbo, gpt-4, etc.)."""
import os

from .base import BaseProvider


class OpenAIProvider(BaseProvider):
    """OpenAI ChatGPT provider using the openai Python SDK."""

    name = "openai"
    DEFAULT_MODEL = "gpt-3.5-turbo"

    def __init__(self, config):
        super().__init__(config)
        self.api_key = config.api_key or os.getenv("OPENAI_KEY") or os.getenv("OPENAI_API_KEY")
        self.model = config.model or self.DEFAULT_MODEL

    def validate_config(self) -> bool:
        return bool(self.api_key)

    def generate(self, system_message: str, user_message: str) -> str:
        try:
            import openai as _openai
        except ImportError as exc:  # pragma: no cover
            raise ImportError(
                "openai package is required for OpenAI provider. "
                "Install it with: pip install openai"
            ) from exc

        # Support both openai >= 1.0 (new) and < 1.0 (legacy) APIs
        if hasattr(_openai, "OpenAI"):
            client = _openai.OpenAI(api_key=self.api_key)
            response = client.chat.completions.create(
                model=self.model,
                messages=[
                    {"role": "system", "content": system_message},
                    {"role": "user", "content": user_message},
                ],
                temperature=self.config.temperature,
            )
            return response.choices[0].message.content
        else:
            # Legacy openai < 1.0
            _openai.api_key = self.api_key
            response = _openai.ChatCompletion.create(
                model=self.model,
                messages=[
                    {"role": "system", "content": system_message},
                    {"role": "user", "content": user_message},
                ],
                temperature=self.config.temperature,
            )
            return response["choices"][0]["message"]["content"]
