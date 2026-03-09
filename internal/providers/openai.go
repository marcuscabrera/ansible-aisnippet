package providers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marcuscabrera/ansible-aisnippet/internal/config"
)

// openaiProvider implements the OpenAI ChatCompletion API.
type openaiProvider struct {
	baseProvider
	apiKey string
	model  string
}

const openaiDefaultModel = "gpt-3.5-turbo"
const openaiAPIURL = "https://api.openai.com/v1/chat/completions"

func newOpenAI(cfg config.ProviderConfig) Provider {
	apiKey := cfg.APIKey
	if apiKey == "" {
		if k := os.Getenv("OPENAI_KEY"); k != "" {
			apiKey = k
		} else {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
	}
	model := cfg.Model
	if model == "" {
		model = openaiDefaultModel
	}
	return &openaiProvider{baseProvider: baseProvider{cfg: cfg}, apiKey: apiKey, model: model}
}

func (p *openaiProvider) Name() string { return "openai" }

func (p *openaiProvider) ValidateConfig() bool { return p.apiKey != "" }

func (p *openaiProvider) Generate(systemMessage, userMessage string) (string, error) {
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
	body, err := httpPost(openaiAPIURL, headers, payload, p.cfg.Timeout)
	if err != nil {
		return "", fmt.Errorf("openai: %w", err)
	}
	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("openai: parsing response: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai: empty choices in response")
	}
	return resp.Choices[0].Message.Content, nil
}
