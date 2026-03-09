package providers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marcuscabrera/ansible-aisnippet/internal/config"
)

// mistralProvider implements the Mistral AI chat completions REST API.
type mistralProvider struct {
	baseProvider
	apiKey string
	model  string
}

const mistralDefaultModel = "mistral-small"
const mistralAPIURL = "https://api.mistral.ai/v1/chat/completions"

func newMistral(cfg config.ProviderConfig) Provider {
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("MISTRAL_API_KEY")
	}
	model := cfg.Model
	if model == "" {
		model = mistralDefaultModel
	}
	return &mistralProvider{baseProvider: baseProvider{cfg: cfg}, apiKey: apiKey, model: model}
}

func (p *mistralProvider) Name() string { return "mistral" }

func (p *mistralProvider) ValidateConfig() bool { return p.apiKey != "" }

func (p *mistralProvider) Generate(systemMessage, userMessage string) (string, error) {
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
	body, err := httpPost(mistralAPIURL, headers, payload, p.cfg.Timeout)
	if err != nil {
		return "", fmt.Errorf("mistral: %w", err)
	}
	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("mistral: parsing response: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("mistral: empty choices in response")
	}
	return resp.Choices[0].Message.Content, nil
}
