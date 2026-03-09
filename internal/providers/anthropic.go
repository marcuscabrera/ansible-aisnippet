package providers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marcuscabrera/ansible-aisnippet/internal/config"
)

// anthropicProvider implements the Anthropic Messages API.
type anthropicProvider struct {
	baseProvider
	apiKey string
	model  string
}

const anthropicDefaultModel = "claude-3-haiku-20240307"
const anthropicAPIURL = "https://api.anthropic.com/v1/messages"
const anthropicVersion = "2023-06-01"

func newAnthropic(cfg config.ProviderConfig) Provider {
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	model := cfg.Model
	if model == "" {
		model = anthropicDefaultModel
	}
	return &anthropicProvider{baseProvider: baseProvider{cfg: cfg}, apiKey: apiKey, model: model}
}

func (p *anthropicProvider) Name() string { return "anthropic" }

func (p *anthropicProvider) ValidateConfig() bool { return p.apiKey != "" }

func (p *anthropicProvider) Generate(systemMessage, userMessage string) (string, error) {
	payload := map[string]interface{}{
		"model":  p.model,
		"system": systemMessage,
		"messages": []map[string]string{
			{"role": "user", "content": userMessage},
		},
		"max_tokens":  1024,
		"temperature": p.cfg.Temperature,
	}
	headers := map[string]string{
		"x-api-key":         p.apiKey,
		"anthropic-version": anthropicVersion,
	}
	body, err := httpPost(anthropicAPIURL, headers, payload, p.cfg.Timeout)
	if err != nil {
		return "", fmt.Errorf("anthropic: %w", err)
	}
	var resp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("anthropic: parsing response: %w", err)
	}
	if len(resp.Content) == 0 {
		return "", fmt.Errorf("anthropic: empty content in response")
	}
	return resp.Content[0].Text, nil
}
