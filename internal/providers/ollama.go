package providers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marcuscabrera/ansible-aisnippet/internal/config"
)

// ollamaProvider implements the Ollama local inference API.
type ollamaProvider struct {
	baseProvider
	baseURL string
	model   string
}

const ollamaDefaultModel = "llama3"
const ollamaDefaultBaseURL = "http://localhost:11434"

func newOllama(cfg config.ProviderConfig) Provider {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = os.Getenv("OLLAMA_BASE_URL")
		if baseURL == "" {
			baseURL = ollamaDefaultBaseURL
		}
	}
	model := cfg.Model
	if model == "" {
		model = os.Getenv("OLLAMA_MODEL")
		if model == "" {
			model = ollamaDefaultModel
		}
	}
	return &ollamaProvider{baseProvider: baseProvider{cfg: cfg}, baseURL: baseURL, model: model}
}

func (p *ollamaProvider) Name() string { return "ollama" }

func (p *ollamaProvider) ValidateConfig() bool { return p.baseURL != "" }

func (p *ollamaProvider) Generate(systemMessage, userMessage string) (string, error) {
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
		return "", fmt.Errorf("ollama: %w", err)
	}
	var resp struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("ollama: parsing response: %w", err)
	}
	return resp.Message.Content, nil
}
