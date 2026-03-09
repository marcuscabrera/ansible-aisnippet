// Package providers defines the Provider interface (Strategy Pattern) and
// all AI provider implementations for ansible-aisnippet.
package providers

import "github.com/marcuscabrera/ansible-aisnippet/internal/config"

// Provider defines the contract for all AI providers.
// It implements the Strategy Pattern, allowing the system to swap providers
// transparently without changing the calling code.
type Provider interface {
	// Name returns the unique identifier for this provider.
	Name() string

	// Generate sends a prompt to the AI provider and returns the response text.
	Generate(systemMessage, userMessage string) (string, error)

	// ValidateConfig returns true if the provider configuration is valid.
	ValidateConfig() bool
}

// baseProvider holds the shared ProviderConfig used by all provider implementations.
type baseProvider struct {
	cfg config.ProviderConfig
}

func (b *baseProvider) ValidateConfig() bool {
	return true
}
