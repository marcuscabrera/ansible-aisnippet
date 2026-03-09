package providers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marcuscabrera/ansible-aisnippet/internal/config"
)

// lmstudioProvider implements the LM Studio OpenAI-compatible local server API.
type lmstudioProvider struct {
	baseProvider
	baseURL string
	model   string
}

const lmstudioDefaultBaseURL = "http://localhost:1234/v1"
const lmstudioDefaultModel = "local-model"

func newLMStudio(cfg config.ProviderConfig) Provider {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = os.Getenv("LMSTUDIO_BASE_URL")
		if baseURL == "" {
			baseURL = lmstudioDefaultBaseURL
		}
	}
	model := cfg.Model
	if model == "" {
		model = os.Getenv("LMSTUDIO_MODEL")
		if model == "" {
			model = lmstudioDefaultModel
		}
	}
	return &lmstudioProvider{baseProvider: baseProvider{cfg: cfg}, baseURL: baseURL, model: model}
}

func (p *lmstudioProvider) Name() string { return "lmstudio" }

func (p *lmstudioProvider) ValidateConfig() bool { return p.baseURL != "" }

func (p *lmstudioProvider) Generate(systemMessage, userMessage string) (string, error) {
	url := p.baseURL + "/chat/completions"
	payload := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemMessage},
			{"role": "user", "content": userMessage},
		},
		"temperature": p.cfg.Temperature,
	}
	body, err := httpPost(url, nil, payload, p.cfg.Timeout)
	if err != nil {
		return "", fmt.Errorf("lmstudio: %w", err)
	}
	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("lmstudio: parsing response: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("lmstudio: empty choices in response")
	}
	return resp.Choices[0].Message.Content, nil
}
