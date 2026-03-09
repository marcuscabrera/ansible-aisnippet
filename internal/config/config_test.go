package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/marcuscabrera/ansible-aisnippet/internal/config"
)

func TestDefaultProviderConfig(t *testing.T) {
	cfg := config.DefaultProviderConfig()
	if cfg.Temperature != 0.0 {
		t.Errorf("expected Temperature=0.0, got %v", cfg.Temperature)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("expected MaxRetries=3, got %d", cfg.MaxRetries)
	}
	if cfg.Timeout != 30 {
		t.Errorf("expected Timeout=30, got %d", cfg.Timeout)
	}
	if cfg.Extra == nil {
		t.Error("expected non-nil Extra map")
	}
}

func TestDefault(t *testing.T) {
	cfg := config.Default()
	if cfg.Provider != "openai" {
		t.Errorf("expected provider=openai, got %q", cfg.Provider)
	}
	if len(cfg.FallbackProviders) != 0 {
		t.Errorf("expected empty fallback providers, got %v", cfg.FallbackProviders)
	}
	if !cfg.Cache.Enabled {
		t.Error("expected cache enabled by default")
	}
	if cfg.Cache.TTL != 3600 {
		t.Errorf("expected TTL=3600, got %d", cfg.Cache.TTL)
	}
	if cfg.Cache.MaxSize != 100 {
		t.Errorf("expected MaxSize=100, got %d", cfg.Cache.MaxSize)
	}
	if !cfg.RateLimit.Enabled {
		t.Error("expected rate limit enabled by default")
	}
	if cfg.RateLimit.RequestsPerMinute != 60 {
		t.Errorf("expected 60 RPM, got %d", cfg.RateLimit.RequestsPerMinute)
	}
}

func TestFromEnv_Defaults(t *testing.T) {
	// Clear AI_ env vars
	for _, k := range []string{
		"AI_PROVIDER", "AI_FALLBACK_PROVIDERS",
		"AI_CACHE_ENABLED", "AI_CACHE_TTL", "AI_CACHE_MAX_SIZE",
		"AI_RATE_LIMIT_ENABLED", "AI_RATE_LIMIT_RPM",
	} {
		os.Unsetenv(k)
	}

	cfg := config.FromEnv()
	if cfg.Provider != "openai" {
		t.Errorf("expected provider=openai, got %q", cfg.Provider)
	}
	if len(cfg.FallbackProviders) != 0 {
		t.Errorf("expected empty fallback providers, got %v", cfg.FallbackProviders)
	}
}

func TestFromEnv_CustomProvider(t *testing.T) {
	t.Setenv("AI_PROVIDER", "anthropic")
	cfg := config.FromEnv()
	if cfg.Provider != "anthropic" {
		t.Errorf("expected provider=anthropic, got %q", cfg.Provider)
	}
}

func TestFromEnv_FallbackProviders(t *testing.T) {
	t.Setenv("AI_FALLBACK_PROVIDERS", "ollama, lmstudio")
	cfg := config.FromEnv()
	if len(cfg.FallbackProviders) != 2 {
		t.Fatalf("expected 2 fallback providers, got %d", len(cfg.FallbackProviders))
	}
	if cfg.FallbackProviders[0] != "ollama" || cfg.FallbackProviders[1] != "lmstudio" {
		t.Errorf("unexpected fallback providers: %v", cfg.FallbackProviders)
	}
}

func TestFromEnv_CacheDisabled(t *testing.T) {
	t.Setenv("AI_CACHE_ENABLED", "false")
	cfg := config.FromEnv()
	if cfg.Cache.Enabled {
		t.Error("expected cache disabled")
	}
}

func TestFromEnv_RateLimitRPM(t *testing.T) {
	t.Setenv("AI_RATE_LIMIT_RPM", "30")
	cfg := config.FromEnv()
	if cfg.RateLimit.RequestsPerMinute != 30 {
		t.Errorf("expected 30 RPM, got %d", cfg.RateLimit.RequestsPerMinute)
	}
}

func TestGetProviderConfig_Defaults(t *testing.T) {
	cfg := config.Default()
	pcfg := cfg.GetProviderConfig("openai")
	if pcfg.APIKey != "" {
		t.Errorf("expected empty APIKey, got %q", pcfg.APIKey)
	}
	if pcfg.MaxRetries != 3 {
		t.Errorf("expected MaxRetries=3, got %d", pcfg.MaxRetries)
	}
}

func TestGetProviderConfig_FromProvidersMap(t *testing.T) {
	cfg := config.Default()
	cfg.Providers = map[string]config.ProviderConfig{
		"openai": {APIKey: "sk-test", Model: "gpt-4", Temperature: 0.5},
	}
	pcfg := cfg.GetProviderConfig("openai")
	if pcfg.APIKey != "sk-test" {
		t.Errorf("expected APIKey=sk-test, got %q", pcfg.APIKey)
	}
	if pcfg.Model != "gpt-4" {
		t.Errorf("expected Model=gpt-4, got %q", pcfg.Model)
	}
	if pcfg.Temperature != 0.5 {
		t.Errorf("expected Temperature=0.5, got %v", pcfg.Temperature)
	}
}

func TestGetProviderConfig_UsesActiveProviderWhenEmpty(t *testing.T) {
	cfg := config.Default()
	cfg.Provider = "mistral"
	cfg.Providers = map[string]config.ProviderConfig{
		"mistral": {APIKey: "key-m"},
	}
	pcfg := cfg.GetProviderConfig("")
	if pcfg.APIKey != "key-m" {
		t.Errorf("expected APIKey=key-m, got %q", pcfg.APIKey)
	}
}

func TestFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	content := `provider: anthropic
fallback_providers:
  - ollama
cache:
  enabled: false
  ttl: 60
rate_limit:
  requests_per_minute: 10
providers:
  anthropic:
    api_key: ant-key
    model: claude-3-haiku-20240307
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.FromFile(path)
	if err != nil {
		t.Fatalf("FromFile error: %v", err)
	}

	if cfg.Provider != "anthropic" {
		t.Errorf("expected provider=anthropic, got %q", cfg.Provider)
	}
	if len(cfg.FallbackProviders) != 1 || cfg.FallbackProviders[0] != "ollama" {
		t.Errorf("unexpected fallback providers: %v", cfg.FallbackProviders)
	}
	if cfg.Cache.Enabled {
		t.Error("expected cache disabled")
	}
	if cfg.Cache.TTL != 60 {
		t.Errorf("expected TTL=60, got %d", cfg.Cache.TTL)
	}
	if cfg.RateLimit.RequestsPerMinute != 10 {
		t.Errorf("expected 10 RPM, got %d", cfg.RateLimit.RequestsPerMinute)
	}
	pcfg := cfg.GetProviderConfig("anthropic")
	if pcfg.APIKey != "ant-key" {
		t.Errorf("expected APIKey=ant-key, got %q", pcfg.APIKey)
	}
	if pcfg.Model != "claude-3-haiku-20240307" {
		t.Errorf("expected model=claude-3-haiku-20240307, got %q", pcfg.Model)
	}
}

func TestFromFile_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yml")
	if err := os.WriteFile(path, []byte("- item1\n- item2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := config.FromFile(path)
	if err == nil {
		t.Error("expected error for non-mapping YAML, got nil")
	}
}

func TestFromFile_MissingFile(t *testing.T) {
	_, err := config.FromFile("/nonexistent/path/config.yml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
