package providers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marcuscabrera/ansible-aisnippet/internal/config"
)

// huggingFaceProvider implements the HuggingFace Inference API.
type huggingFaceProvider struct {
	baseProvider
	apiKey string
	model  string
}

const huggingFaceDefaultModel = "mistralai/Mistral-7B-Instruct-v0.2"
const huggingFaceAPIURLTemplate = "https://api-inference.huggingface.co/models/%s"

func newHuggingFace(cfg config.ProviderConfig) Provider {
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("HF_API_TOKEN")
	}
	model := cfg.Model
	if model == "" {
		model = os.Getenv("HF_MODEL")
		if model == "" {
			model = huggingFaceDefaultModel
		}
	}
	return &huggingFaceProvider{baseProvider: baseProvider{cfg: cfg}, apiKey: apiKey, model: model}
}

func (p *huggingFaceProvider) Name() string { return "huggingface" }

func (p *huggingFaceProvider) ValidateConfig() bool { return p.apiKey != "" }

func (p *huggingFaceProvider) Generate(systemMessage, userMessage string) (string, error) {
	url := fmt.Sprintf(huggingFaceAPIURLTemplate, p.model)
	// Instruction-following format compatible with most chat-tuned models.
	temperature := p.cfg.Temperature
	if temperature < 0.01 {
		temperature = 0.01
	}
	payload := map[string]interface{}{
		"inputs": fmt.Sprintf("[INST] %s\n\n%s [/INST]", systemMessage, userMessage),
		"parameters": map[string]interface{}{
			"temperature":      temperature,
			"max_new_tokens":   1024,
			"return_full_text": false,
		},
	}
	headers := map[string]string{
		"Authorization": "Bearer " + p.apiKey,
	}
	body, err := httpPost(url, headers, payload, p.cfg.Timeout)
	if err != nil {
		return "", fmt.Errorf("huggingface: %w", err)
	}
	// Response is a JSON array of generated texts.
	var listResp []struct {
		GeneratedText string `json:"generated_text"`
	}
	if err := json.Unmarshal(body, &listResp); err == nil && len(listResp) > 0 {
		return listResp[0].GeneratedText, nil
	}
	// Fallback: single object.
	var singleResp struct {
		GeneratedText string `json:"generated_text"`
	}
	if err := json.Unmarshal(body, &singleResp); err != nil {
		return "", fmt.Errorf("huggingface: parsing response: %w", err)
	}
	return singleResp.GeneratedText, nil
}
