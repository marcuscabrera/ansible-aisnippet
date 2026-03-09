// Package config provides centralized configuration for ansible-aisnippet.
//
// It supports loading from environment variables or a YAML file.
//
// Environment variable hierarchy (first match wins):
//   - Variables explicitly set by the caller
//   - OS environment variables
//   - Default values
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ProviderConfig holds configuration for a single AI provider.
type ProviderConfig struct {
	APIKey      string            `yaml:"api_key"`
	Model       string            `yaml:"model"`
	BaseURL     string            `yaml:"base_url"`
	Temperature float64           `yaml:"temperature"`
	MaxRetries  int               `yaml:"max_retries"`
	Timeout     int               `yaml:"timeout"`
	Extra       map[string]string `yaml:"extra"`
}

// DefaultProviderConfig returns a ProviderConfig populated with sensible defaults.
func DefaultProviderConfig() ProviderConfig {
	return ProviderConfig{
		Temperature: 0.0,
		MaxRetries:  3,
		Timeout:     30,
		Extra:       make(map[string]string),
	}
}

// CacheConfig holds configuration for the response cache.
type CacheConfig struct {
	Enabled bool `yaml:"enabled"`
	TTL     int  `yaml:"ttl"`
	MaxSize int  `yaml:"max_size"`
}

// DefaultCacheConfig returns CacheConfig with sensible defaults.
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Enabled: true,
		TTL:     3600,
		MaxSize: 100,
	}
}

// RateLimitConfig holds configuration for the rate limiter.
type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerMinute int  `yaml:"requests_per_minute"`
}

// DefaultRateLimitConfig returns RateLimitConfig with sensible defaults.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 60,
	}
}

// Config is the top-level configuration object for ansible-aisnippet.
type Config struct {
	Provider          string                    `yaml:"provider"`
	FallbackProviders []string                  `yaml:"fallback_providers"`
	Providers         map[string]ProviderConfig `yaml:"providers"`
	Cache             CacheConfig               `yaml:"cache"`
	RateLimit         RateLimitConfig           `yaml:"rate_limit"`
}

// Default returns a Config populated with sensible defaults.
func Default() *Config {
	return &Config{
		Provider:          "openai",
		FallbackProviders: []string{},
		Providers:         make(map[string]ProviderConfig),
		Cache:             DefaultCacheConfig(),
		RateLimit:         DefaultRateLimitConfig(),
	}
}

// GetProviderConfig returns a ProviderConfig for the given provider name.
// Falls back to the active provider if name is empty. Provider-specific settings
// from Config.Providers are merged with defaults.
func (c *Config) GetProviderConfig(name string) ProviderConfig {
	if name == "" {
		name = c.Provider
	}
	base := DefaultProviderConfig()
	if overrides, ok := c.Providers[name]; ok {
		if overrides.APIKey != "" {
			base.APIKey = overrides.APIKey
		}
		if overrides.Model != "" {
			base.Model = overrides.Model
		}
		if overrides.BaseURL != "" {
			base.BaseURL = overrides.BaseURL
		}
		if overrides.Temperature != 0 {
			base.Temperature = overrides.Temperature
		}
		if overrides.MaxRetries != 0 {
			base.MaxRetries = overrides.MaxRetries
		}
		if overrides.Timeout != 0 {
			base.Timeout = overrides.Timeout
		}
		if len(overrides.Extra) > 0 {
			base.Extra = overrides.Extra
		}
	}
	return base
}

// FromEnv builds a Config from OS environment variables.
//
// Supported variables:
//
//	AI_PROVIDER               Active provider (default: openai)
//	AI_FALLBACK_PROVIDERS     Comma-separated fallback list
//	AI_CACHE_ENABLED          "true" / "false"
//	AI_CACHE_TTL              Seconds (int)
//	AI_CACHE_MAX_SIZE         Items (int)
//	AI_RATE_LIMIT_ENABLED     "true" / "false"
//	AI_RATE_LIMIT_RPM         Requests per minute (int)
func FromEnv() *Config {
	cfg := Default()

	if v := os.Getenv("AI_PROVIDER"); v != "" {
		cfg.Provider = v
	}

	if v := os.Getenv("AI_FALLBACK_PROVIDERS"); v != "" {
		parts := strings.Split(v, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				cfg.FallbackProviders = append(cfg.FallbackProviders, p)
			}
		}
	}

	if v := os.Getenv("AI_CACHE_ENABLED"); v != "" {
		cfg.Cache.Enabled = strings.ToLower(v) != "false"
	}
	if v := os.Getenv("AI_CACHE_TTL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Cache.TTL = n
		}
	}
	if v := os.Getenv("AI_CACHE_MAX_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Cache.MaxSize = n
		}
	}

	if v := os.Getenv("AI_RATE_LIMIT_ENABLED"); v != "" {
		cfg.RateLimit.Enabled = strings.ToLower(v) != "false"
	}
	if v := os.Getenv("AI_RATE_LIMIT_RPM"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.RateLimit.RequestsPerMinute = n
		}
	}

	return cfg
}

// fileConfig mirrors the YAML layout for unmarshalling from a config file.
type fileConfig struct {
	Provider          string                    `yaml:"provider"`
	FallbackProviders []string                  `yaml:"fallback_providers"`
	Cache             *CacheConfig              `yaml:"cache"`
	RateLimit         *RateLimitConfig          `yaml:"rate_limit"`
	Providers         map[string]ProviderConfig `yaml:"providers"`
}

// FromFile loads a Config from a YAML file.
//
// Expected YAML layout:
//
//	provider: openai
//	fallback_providers:
//	  - anthropic
//	  - ollama
//	cache:
//	  enabled: true
//	  ttl: 3600
//	  max_size: 100
//	rate_limit:
//	  enabled: true
//	  requests_per_minute: 60
//	providers:
//	  openai:
//	    api_key: sk-...
//	    model: gpt-4
func FromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	// First check if it's a mapping (not a list/scalar)
	var raw interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}
	if _, ok := raw.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("configuration file %q must contain a YAML mapping", path)
	}

	var fc fileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("unmarshalling config file %q: %w", path, err)
	}

	cfg := Default()
	if fc.Provider != "" {
		cfg.Provider = fc.Provider
	}
	if len(fc.FallbackProviders) > 0 {
		cfg.FallbackProviders = fc.FallbackProviders
	}
	if fc.Cache != nil {
		cfg.Cache = *fc.Cache
	}
	if fc.RateLimit != nil {
		cfg.RateLimit = *fc.RateLimit
	}
	if fc.Providers != nil {
		cfg.Providers = fc.Providers
	}

	return cfg, nil
}
