package providers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marcuscabrera/ansible-aisnippet/internal/config"
)

// azureProvider implements the Azure OpenAI Service using the REST chat completions API.
type azureProvider struct {
	baseProvider
	apiKey     string
	endpoint   string
	deployment string
	apiVersion string
}

const azureDefaultDeployment = "gpt-35-turbo"
const azureDefaultAPIVersion = "2024-02-01"

func newAzure(cfg config.ProviderConfig) Provider {
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("AZURE_OPENAI_KEY")
	}
	endpoint := cfg.BaseURL
	if endpoint == "" {
		endpoint = os.Getenv("AZURE_OPENAI_ENDPOINT")
	}
	deployment := cfg.Model
	if deployment == "" {
		deployment = os.Getenv("AZURE_OPENAI_DEPLOYMENT")
		if deployment == "" {
			deployment = azureDefaultDeployment
		}
	}
	apiVersion := azureDefaultAPIVersion
	if v, ok := cfg.Extra["api_version"]; ok && v != "" {
		apiVersion = v
	}
	return &azureProvider{
		baseProvider: baseProvider{cfg: cfg},
		apiKey:       apiKey,
		endpoint:     endpoint,
		deployment:   deployment,
		apiVersion:   apiVersion,
	}
}

func (p *azureProvider) Name() string { return "azure" }

func (p *azureProvider) ValidateConfig() bool { return p.apiKey != "" && p.endpoint != "" }

func (p *azureProvider) Generate(systemMessage, userMessage string) (string, error) {
	url := fmt.Sprintf(
		"%sopenai/deployments/%s/chat/completions?api-version=%s",
		p.endpoint, p.deployment, p.apiVersion,
	)
	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "system", "content": systemMessage},
			{"role": "user", "content": userMessage},
		},
		"temperature": p.cfg.Temperature,
	}
	headers := map[string]string{
		"api-key": p.apiKey,
	}
	body, err := httpPost(url, headers, payload, p.cfg.Timeout)
	if err != nil {
		return "", fmt.Errorf("azure: %w", err)
	}
	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("azure: parsing response: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("azure: empty choices in response")
	}
	return resp.Choices[0].Message.Content, nil
}
