"""Azure OpenAI provider using the OpenAI SDK with Azure endpoint."""
import os

from .base import BaseProvider


class AzureOpenAIProvider(BaseProvider):
    """Azure OpenAI Service provider.

    Requires:
        AZURE_OPENAI_KEY        – API key
        AZURE_OPENAI_ENDPOINT   – e.g. https://<resource>.openai.azure.com/
        AZURE_OPENAI_DEPLOYMENT – deployment name (default: gpt-35-turbo)
    """

    name = "azure"
    DEFAULT_DEPLOYMENT = "gpt-35-turbo"
    DEFAULT_API_VERSION = "2024-02-01"

    def __init__(self, config):
        super().__init__(config)
        self.api_key = config.api_key or os.getenv("AZURE_OPENAI_KEY")
        self.endpoint = config.base_url or os.getenv("AZURE_OPENAI_ENDPOINT", "")
        self.deployment = config.model or os.getenv(
            "AZURE_OPENAI_DEPLOYMENT", self.DEFAULT_DEPLOYMENT
        )
        self.api_version = config.extra.get("api_version", self.DEFAULT_API_VERSION)

    def validate_config(self) -> bool:
        return bool(self.api_key) and bool(self.endpoint)

    def generate(self, system_message: str, user_message: str) -> str:
        try:
            import openai as _openai
        except ImportError as exc:  # pragma: no cover
            raise ImportError(
                "openai package is required for Azure OpenAI provider. "
                "Install it with: pip install openai"
            ) from exc

        if hasattr(_openai, "AzureOpenAI"):
            client = _openai.AzureOpenAI(
                api_key=self.api_key,
                azure_endpoint=self.endpoint,
                api_version=self.api_version,
            )
            response = client.chat.completions.create(
                model=self.deployment,
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
            _openai.api_base = self.endpoint
            _openai.api_type = "azure"
            _openai.api_version = self.api_version
            response = _openai.ChatCompletion.create(
                engine=self.deployment,
                messages=[
                    {"role": "system", "content": system_message},
                    {"role": "user", "content": user_message},
                ],
                temperature=self.config.temperature,
            )
            return response["choices"][0]["message"]["content"]
