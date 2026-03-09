package providers

import (
	"fmt"
	"strings"

	"github.com/marcuscabrera/ansible-aisnippet/internal/config"
)

// registry maps provider name → constructor function.
var registry = map[string]func(config.ProviderConfig) Provider{}

// register adds a provider constructor to the registry.
func register(name string, ctor func(config.ProviderConfig) Provider) {
	registry[name] = ctor
}

func init() {
	register("openai", func(cfg config.ProviderConfig) Provider { return newOpenAI(cfg) })
	register("anthropic", func(cfg config.ProviderConfig) Provider { return newAnthropic(cfg) })
	register("google", func(cfg config.ProviderConfig) Provider { return newGoogle(cfg) })
	register("azure", func(cfg config.ProviderConfig) Provider { return newAzure(cfg) })
	register("mistral", func(cfg config.ProviderConfig) Provider { return newMistral(cfg) })
	register("cohere", func(cfg config.ProviderConfig) Provider { return newCohere(cfg) })
	register("ollama", func(cfg config.ProviderConfig) Provider { return newOllama(cfg) })
	register("lmstudio", func(cfg config.ProviderConfig) Provider { return newLMStudio(cfg) })
	register("llama", func(cfg config.ProviderConfig) Provider { return newLlama(cfg) })
	register("huggingface", func(cfg config.ProviderConfig) Provider { return newHuggingFace(cfg) })
	register("openrouter", func(cfg config.ProviderConfig) Provider { return newOpenRouter(cfg) })
	register("zen", func(cfg config.ProviderConfig) Provider { return newZen(cfg) })
}

// Create instantiates a provider by name with the given config.
//
// provider_name must be one of: openai, anthropic, google, azure, mistral,
// cohere, ollama, lmstudio, llama, huggingface, openrouter, zen.
//
// Returns an error if the provider name is not registered.
func Create(providerName string, cfg config.ProviderConfig) (Provider, error) {
	name := strings.ToLower(strings.TrimSpace(providerName))
	ctor, ok := registry[name]
	if !ok {
		available := ListProviders()
		return nil, fmt.Errorf("unknown provider %q; available providers: %s",
			providerName, strings.Join(available, ", "))
	}
	return ctor(cfg), nil
}

// ListProviders returns a sorted list of all registered provider names.
func ListProviders() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	// Simple sort
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[i] > names[j] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	return names
}
