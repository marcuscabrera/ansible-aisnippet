package providers

import (
	"fmt"
	"strings"
)

// FallbackManager tries a list of providers in order and returns the first
// successful response. All errors are collected and surfaced in a single error
// only when every provider has been exhausted.
type FallbackManager struct {
	providers []Provider
}

// NewFallbackManager creates a FallbackManager from an ordered list of providers.
// The first element is treated as the primary provider.
func NewFallbackManager(providers []Provider) (*FallbackManager, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("FallbackManager requires at least one provider")
	}
	return &FallbackManager{providers: providers}, nil
}

// Generate sends the prompt to each provider in order, returning the first
// successful response and the name of the provider that handled it.
// Returns an error if all providers fail.
func (f *FallbackManager) Generate(systemMessage, userMessage string) (string, string, error) {
	var errs []string
	for _, p := range f.providers {
		result, err := p.Generate(systemMessage, userMessage)
		if err == nil {
			return result, p.Name(), nil
		}
		errs = append(errs, fmt.Sprintf("%s: %v", p.Name(), err))
	}
	return "", "", fmt.Errorf("all providers failed to generate a response. Errors:\n  - %s",
		strings.Join(errs, "\n  - "))
}
