package providers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marcuscabrera/ansible-aisnippet/internal/config"
)

// cohereProvider implements the Cohere Chat API.
type cohereProvider struct {
	baseProvider
	apiKey string
	model  string
}

const cohereDefaultModel = "command"
const cohereAPIURL = "https://api.cohere.ai/v1/chat"

func newCohere(cfg config.ProviderConfig) Provider {
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("COHERE_API_KEY")
	}
	model := cfg.Model
	if model == "" {
		model = cohereDefaultModel
	}
	return &cohereProvider{baseProvider: baseProvider{cfg: cfg}, apiKey: apiKey, model: model}
}

func (p *cohereProvider) Name() string { return "cohere" }

func (p *cohereProvider) ValidateConfig() bool { return p.apiKey != "" }

func (p *cohereProvider) Generate(systemMessage, userMessage string) (string, error) {
	payload := map[string]interface{}{
		"model":       p.model,
		"preamble":    systemMessage,
		"message":     userMessage,
		"temperature": p.cfg.Temperature,
	}
	headers := map[string]string{
		"Authorization": "Bearer " + p.apiKey,
	}
	body, err := httpPost(cohereAPIURL, headers, payload, p.cfg.Timeout)
	if err != nil {
		return "", fmt.Errorf("cohere: %w", err)
	}
	var resp struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("cohere: parsing response: %w", err)
	}
	return resp.Text, nil
}
