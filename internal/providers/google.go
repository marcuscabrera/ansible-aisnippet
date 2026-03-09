package providers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marcuscabrera/ansible-aisnippet/internal/config"
)

// googleProvider implements the Google Generative Language REST API (Gemini).
type googleProvider struct {
	baseProvider
	apiKey string
	model  string
}

const googleDefaultModel = "gemini-pro"
const googleAPIURLTemplate = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent"

func newGoogle(cfg config.ProviderConfig) Provider {
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	model := cfg.Model
	if model == "" {
		model = googleDefaultModel
	}
	return &googleProvider{baseProvider: baseProvider{cfg: cfg}, apiKey: apiKey, model: model}
}

func (p *googleProvider) Name() string { return "google" }

func (p *googleProvider) ValidateConfig() bool { return p.apiKey != "" }

func (p *googleProvider) Generate(systemMessage, userMessage string) (string, error) {
	url := fmt.Sprintf(googleAPIURLTemplate+"?key=%s", p.model, p.apiKey)
	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": systemMessage + "\n\n" + userMessage},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     p.cfg.Temperature,
			"maxOutputTokens": 1024,
		},
	}
	body, err := httpPost(url, nil, payload, p.cfg.Timeout)
	if err != nil {
		return "", fmt.Errorf("google: %w", err)
	}
	var resp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("google: parsing response: %w", err)
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("google: empty candidates in response")
	}
	return resp.Candidates[0].Content.Parts[0].Text, nil
}
