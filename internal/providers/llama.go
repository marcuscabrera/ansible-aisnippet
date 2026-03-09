package providers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marcuscabrera/ansible-aisnippet/internal/config"
)

// llamaProvider implements the Meta Llama provider via Ollama-compatible endpoints.
type llamaProvider struct {
	baseProvider
	baseURL string
	model   string
}

const llamaDefaultModel = "llama3"
const llamaDefaultBaseURL = "http://localhost:11434"

func newLlama(cfg config.ProviderConfig) Provider {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = os.Getenv("LLAMA_BASE_URL")
		if baseURL == "" {
			baseURL = llamaDefaultBaseURL
		}
	}
	model := cfg.Model
	if model == "" {
		model = os.Getenv("LLAMA_MODEL")
		if model == "" {
			model = llamaDefaultModel
		}
	}
	return &llamaProvider{baseProvider: baseProvider{cfg: cfg}, baseURL: baseURL, model: model}
}

func (p *llamaProvider) Name() string { return "llama" }

func (p *llamaProvider) ValidateConfig() bool { return p.baseURL != "" }

func (p *llamaProvider) Generate(systemMessage, userMessage string) (string, error) {
	url := p.baseURL + "/api/chat"
	payload := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemMessage},
			{"role": "user", "content": userMessage},
		},
		"stream": false,
		"options": map[string]interface{}{
			"temperature": p.cfg.Temperature,
		},
	}
	body, err := httpPost(url, nil, payload, p.cfg.Timeout)
	if err != nil {
		return "", fmt.Errorf("llama: %w", err)
	}
	var resp struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("llama: parsing response: %w", err)
	}
	return resp.Message.Content, nil
}
