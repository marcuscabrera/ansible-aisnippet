package providers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marcuscabrera/ansible-aisnippet/internal/config"
)

// openRouterProvider implements the OpenRouter aggregator API (OpenAI-compatible).
type openRouterProvider struct {
	baseProvider
	apiKey  string
	baseURL string
	model   string
}

const openRouterDefaultBaseURL = "https://openrouter.ai/api/v1"
const openRouterDefaultModel = "openai/gpt-3.5-turbo"

func newOpenRouter(cfg config.ProviderConfig) Provider {
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENROUTER_API_KEY")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = os.Getenv("OPENROUTER_BASE_URL")
		if baseURL == "" {
			baseURL = openRouterDefaultBaseURL
		}
	}
	model := cfg.Model
	if model == "" {
		model = os.Getenv("OPENROUTER_MODEL")
		if model == "" {
			model = openRouterDefaultModel
		}
	}
	return &openRouterProvider{
		baseProvider: baseProvider{cfg: cfg},
		apiKey:       apiKey,
		baseURL:      baseURL,
		model:        model,
	}
}

func (p *openRouterProvider) Name() string { return "openrouter" }

func (p *openRouterProvider) ValidateConfig() bool { return p.apiKey != "" }

func (p *openRouterProvider) Generate(systemMessage, userMessage string) (string, error) {
	url := p.baseURL + "/chat/completions"
	payload := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemMessage},
			{"role": "user", "content": userMessage},
		},
		"temperature": p.cfg.Temperature,
	}
	headers := map[string]string{
		"Authorization": "Bearer " + p.apiKey,
	}
	body, err := httpPost(url, headers, payload, p.cfg.Timeout)
	if err != nil {
		return "", fmt.Errorf("openrouter: %w", err)
	}
	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("openrouter: parsing response: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openrouter: empty choices in response")
	}
	return resp.Choices[0].Message.Content, nil
}
