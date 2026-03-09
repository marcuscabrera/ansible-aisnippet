"""Abstract base class for all AI providers (Strategy Pattern interface)."""
from abc import ABC, abstractmethod


class BaseProvider(ABC):
    """Abstract base class defining the contract for all AI providers.

    Implements the Strategy Pattern, allowing the system to swap providers
    transparently without changing the calling code.
    """

    def __init__(self, config):
        """Initialise the provider with a ProviderConfig instance."""
        self.config = config

    @property
    @abstractmethod
    def name(self) -> str:
        """Unique identifier for this provider."""

    @abstractmethod
    def generate(self, system_message: str, user_message: str) -> str:
        """Send a prompt to the AI provider and return the response text.

        Args:
            system_message: The system/context prompt.
            user_message: The user request.

        Returns:
            The raw text response from the model.

        Raises:
            RuntimeError: If the API call fails after exhausting retries.
        """

    def validate_config(self) -> bool:
        """Return True if the provider configuration is valid.

        Override in subclasses to add provider-specific validation.
        """
        return True
