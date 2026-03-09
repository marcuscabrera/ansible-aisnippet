package providers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marcuscabrera/ansible-aisnippet/internal/config"
)

// zenProvider implements the OpenCode Zen API (OpenAI-compatible).
type zenProvider struct {
	baseProvider
	apiKey  string
	baseURL string
	model   string
}

const zenDefaultBaseURL = "https://api.opencode.ai/v1"
const zenDefaultModel = "zen"

func newZen(cfg config.ProviderConfig) Provider {
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("ZEN_API_KEY")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = os.Getenv("ZEN_BASE_URL")
		if baseURL == "" {
			baseURL = zenDefaultBaseURL
		}
	}
	model := cfg.Model
	if model == "" {
		model = os.Getenv("ZEN_MODEL")
		if model == "" {
			model = zenDefaultModel
		}
	}
	return &zenProvider{
		baseProvider: baseProvider{cfg: cfg},
		apiKey:       apiKey,
		baseURL:      baseURL,
		model:        model,
	}
}

func (p *zenProvider) Name() string { return "zen" }

func (p *zenProvider) ValidateConfig() bool { return p.apiKey != "" }

func (p *zenProvider) Generate(systemMessage, userMessage string) (string, error) {
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
		return "", fmt.Errorf("zen: %w", err)
	}
	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("zen: parsing response: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("zen: empty choices in response")
	}
	return resp.Choices[0].Message.Content, nil
}
