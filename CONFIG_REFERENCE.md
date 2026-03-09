# ansible-aisnippet Configuration File Reference

## Complete Configuration Structure

Based on the analysis of `internal/config/config.go`, here is a comprehensive guide to all supported configuration options:

### 1. Top-Level Configuration Keys

```yaml
# The primary AI provider to use
provider: openai

# List of fallback providers (tried in order if primary fails)
fallback_providers:
  - anthropic
  - ollama

# Cache configuration
cache:
  enabled: true
  ttl: 3600          # Time-to-live in seconds
  max_size: 100      # Maximum number of cached responses

# Rate limiting configuration
rate_limit:
  enabled: true
  requests_per_minute: 60

# Per-provider configuration
providers:
  openai:
    api_key: sk-...
    model: gpt-4o
    temperature: 0.0
    max_retries: 3
    timeout: 30
    base_url: https://api.openai.com/v1  # Optional
    extra:                                # Optional provider-specific fields
      custom_field: value
```

---

## Detailed Configuration Reference

### Global Settings

#### `provider` (string)
- **Default:** `openai`
- **Description:** The primary AI provider to use for generation
- **Supported Values:** 
  - `openai`, `anthropic`, `google`, `azure`, `mistral`, `cohere`
  - `ollama`, `lmstudio`, `llama`, `huggingface`, `openrouter`, `zen`
- **Environment Variable:** `AI_PROVIDER`

#### `fallback_providers` (list of strings)
- **Default:** `[]` (empty)
- **Description:** List of backup providers tried in order if the primary provider fails
- **Environment Variable:** `AI_FALLBACK_PROVIDERS` (comma-separated: `anthropic,ollama`)
- **Example:**
  ```yaml
  fallback_providers:
    - anthropic
    - ollama
    - lmstudio
  ```

---

### Cache Configuration

#### `cache.enabled` (boolean)
- **Default:** `true`
- **Description:** Enable/disable response caching to reduce redundant API calls
- **Environment Variable:** `AI_CACHE_ENABLED` (`"true"` or `"false"`)

#### `cache.ttl` (integer)
- **Default:** `3600` (seconds)
- **Description:** Time-to-live for cached responses in seconds (1 hour default)
- **Environment Variable:** `AI_CACHE_TTL`
- **Recommended Range:** 300-86400 (5 minutes to 24 hours)

#### `cache.max_size` (integer)
- **Default:** `100`
- **Description:** Maximum number of responses to keep in cache
- **Environment Variable:** `AI_CACHE_MAX_SIZE`
- **Recommended Range:** 10-1000 depending on memory constraints

---

### Rate Limiting Configuration

#### `rate_limit.enabled` (boolean)
- **Default:** `true`
- **Description:** Enable/disable request rate limiting
- **Environment Variable:** `AI_RATE_LIMIT_ENABLED`

#### `rate_limit.requests_per_minute` (integer)
- **Default:** `60`
- **Description:** Maximum requests per minute allowed
- **Environment Variable:** `AI_RATE_LIMIT_RPM`
- **Note:** Set according to your AI provider's rate limits
- **Common Values:**
  - OpenAI: 60 (free tier) to 3500+ (paid)
  - Anthropic: 50+ depending on tier
  - Ollama/LM Studio: 1000+ (local, no limit)

---

### Per-Provider Configuration

Each provider under `providers.<provider_name>:` supports:

#### `api_key` (string)
- **Description:** API authentication key for the provider
- **Required For:** Cloud providers (OpenAI, Anthropic, Google, Azure, Mistral, Cohere, HuggingFace, OpenRouter, ZenAI)
- **Optional For:** Local providers (Ollama, LM Studio, Llama)
- **Fallback:** If not specified, loads from provider-specific environment variable
- **Environment Variable Examples:**
  - OpenAI: `OPENAI_KEY` or `OPENAI_API_KEY`
  - Anthropic: `ANTHROPIC_API_KEY`
  - Google: `GOOGLE_API_KEY`
  - Mistral: `MISTRAL_API_KEY`
  - Cohere: `COHERE_API_KEY`
  - Azure: `AZURE_OPENAI_KEY`
  - HuggingFace: `HF_API_TOKEN`
  - OpenRouter: `OPENROUTER_API_KEY`
  - ZenAI: `ZEN_API_KEY`

#### `model` (string)
- **Description:** Specific model identifier to use within the provider
- **Examples:**
  - OpenAI: `gpt-4o`, `gpt-4-turbo`, `gpt-3.5-turbo`
  - Anthropic: `claude-3-opus-20240229`, `claude-3-sonnet-20240229`, `claude-3-haiku-20240307`
  - Google: `gemini-pro`, `gemini-1.5-pro`
  - Mistral: `mistral-large`, `mistral-medium`, `mistral-small`
  - Ollama: `llama3`, `mistral`, `neural-chat`
  - Azure: `deployment-name` (from your Azure instance)
  - Cohere: `command-r-plus`, `command-r`
- **Default Behavior:** 
  - If not specified, each provider uses a sensible default (usually the latest or most stable model)

#### `temperature` (float, 0.0-1.0+)
- **Default:** `0.0` (deterministic)
- **Description:** Controls response randomness/creativity
  - `0.0`: Most deterministic, consistent responses (good for code generation)
  - `0.5-0.7`: Balanced creativity and consistency
  - `1.0+`: More creative/random responses
- **Note:** Most providers support 0.0-2.0, but 0.0-1.0 is standard
- **Recommended:** `0.0` for Ansible task generation (deterministic output)

#### `max_retries` (integer)
- **Default:** `3`
- **Description:** Number of times to retry a request if it fails
- **Recommended Range:** 1-5
- **Use Cases:**
  - Higher values for unreliable networks
  - Lower values for fast-fail scenarios

#### `timeout` (integer)
- **Default:** `30` (seconds)
- **Description:** Request timeout in seconds
- **Recommended Range:**
  - 10-20 for cloud APIs (OpenAI, Anthropic, etc.)
  - 30-60 for local providers (Ollama, LM Studio)
  - 60+ for very large models or slow networks

#### `base_url` (string, optional)
- **Description:** Custom API endpoint URL for the provider
- **Required For:** Self-hosted or proxy scenarios
- **Examples:**
  - Ollama: `http://localhost:11434`
  - LM Studio: `http://localhost:1234/v1`
  - Azure: `https://<your-resource>.openai.azure.com/`
  - Local OpenAI proxy: `http://localhost:8000/v1`
- **Note:** Not typically needed if using official APIs

#### `extra` (map[string]string, optional)
- **Description:** Provider-specific additional configuration fields
- **Advanced Usage:** For custom headers, metadata, or provider-specific options
- **Example:**
  ```yaml
  providers:
    openai:
      api_key: sk-...
      extra:
        organization: org-123
        custom_header: value
  ```

---

## Complete Configuration Examples

### Example 1: Multi-Provider Setup with Fallback Chain

```yaml
provider: openai
fallback_providers:
  - anthropic
  - ollama

cache:
  enabled: true
  ttl: 7200      # 2 hours
  max_size: 200

rate_limit:
  enabled: true
  requests_per_minute: 100

providers:
  openai:
    api_key: sk-proj-...
    model: gpt-4o
    temperature: 0.0
    max_retries: 3
    timeout: 30

  anthropic:
    api_key: sk-ant-...
    model: claude-3-sonnet-20240229
    temperature: 0.0
    timeout: 60

  ollama:
    base_url: http://localhost:11434
    model: llama3
    timeout: 120
```

### Example 2: Cost-Optimized (Fast, Cheap Models)

```yaml
provider: openai

cache:
  enabled: true
  ttl: 3600
  max_size: 150

rate_limit:
  enabled: true
  requests_per_minute: 200

providers:
  openai:
    api_key: sk-...
    model: gpt-3.5-turbo
    temperature: 0.0
    max_retries: 2
    timeout: 20
```

### Example 3: Offline/Self-Hosted Setup

```yaml
provider: ollama
fallback_providers:
  - lmstudio

cache:
  enabled: true
  ttl: 1800

rate_limit:
  enabled: false  # Local providers typically don't need rate limiting

providers:
  ollama:
    base_url: http://localhost:11434
    model: mistral
    timeout: 120

  lmstudio:
    base_url: http://localhost:1234/v1
    model: local-model
    timeout: 120
```

### Example 4: High-Reliability Enterprise Setup

```yaml
provider: anthropic
fallback_providers:
  - openai
  - anthropic  # Secondary Anthropic instance

cache:
  enabled: true
  ttl: 10800    # 3 hours
  max_size: 500

rate_limit:
  enabled: true
  requests_per_minute: 50

providers:
  anthropic:
    api_key: sk-ant-...
    model: claude-3-opus-20240229
    temperature: 0.0
    max_retries: 5
    timeout: 60

  openai:
    api_key: sk-...
    model: gpt-4-turbo
    temperature: 0.0
    max_retries: 5
    timeout: 60
```

### Example 5: Azure OpenAI Deployment

```yaml
provider: azure

providers:
  azure:
    api_key: your-azure-api-key
    model: your-deployment-name
    base_url: https://your-resource.openai.azure.com/
    temperature: 0.0
    timeout: 30
```

---

## Configuration Loading Priority

When using both environment variables and config files:

1. **Explicit config file values** (if `--config` flag is used) take highest priority
2. **Environment variables** override defaults
3. **Default values** are used if nothing is specified

**Example:** If config file sets `provider: anthropic` and env var `AI_PROVIDER=openai`, the config file wins.

---

## Loading Configuration

### Via CLI Flag

```bash
ansible-aisnippet --config /path/to/config.yml generate "install nginx"
ansible-aisnippet -c ~/.ansible-aisnippet/config.yml generate "setup database"
```

### Default Locations (Not yet implemented, but useful suggestion)

Consider storing config at:
- `~/.ansible-aisnippet/config.yml` (user home)
- `./.ansible-aisnippet.yml` (project root)
- `$XDG_CONFIG_HOME/ansible-aisnippet/config.yml` (Linux/Mac)

For now, always use the `--config` flag explicitly.

---

## Environment Variable Reference

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `AI_PROVIDER` | string | `openai` | Active provider |
| `AI_FALLBACK_PROVIDERS` | string (comma-separated) | _(none)_ | Backup providers |
| `AI_CACHE_ENABLED` | string (`"true"`/`"false"`) | `true` | Enable caching |
| `AI_CACHE_TTL` | integer | `3600` | Cache TTL in seconds |
| `AI_CACHE_MAX_SIZE` | integer | `100` | Max cached items |
| `AI_RATE_LIMIT_ENABLED` | string (`"true"`/`"false"`) | `true` | Enable rate limiting |
| `AI_RATE_LIMIT_RPM` | integer | `60` | Requests per minute |

Provider API keys follow standard conventions (e.g., `OPENAI_KEY`, `ANTHROPIC_API_KEY`).

---

## Validation Notes

- **api_key:** Validated when provider is first used (not at config load time)
- **model:** Not validated; provider will error at runtime if model doesn't exist
- **timeout:** Must be > 0; recommended minimum is 10 seconds
- **temperature:** Must be >= 0.0; maximum varies by provider
- **max_retries:** Typical range 0-10
- **cache.ttl:** Typical range 300-86400 seconds
- **cache.max_size:** Typical range 10-1000
- **requests_per_minute:** Should respect provider rate limits

---

## Tips & Best Practices

1. **Sensitive Data:**
   - Never commit `api_key` values to version control
   - Use environment variables or `.gitignore`-d config files for local development
   - Use secrets management systems in production

2. **Multi-Provider Strategy:**
   - Put expensive/powerful models as primary (e.g., GPT-4)
   - Add cheaper models as fallbacks (e.g., GPT-3.5-turbo)
   - Always include a local option (Ollama) for ultimate fallback

3. **Performance:**
   - Enable caching to reduce API costs and latency
   - Adjust `cache.ttl` based on how fresh you need generated tasks to be
   - Use high `timeout` values for local/slow providers

4. **Reliability:**
   - Set `max_retries: 3-5` for production workloads
   - Use `fallback_providers` for critical automation
   - Test all configured providers in dry-run mode first

5. **Cost Optimization:**
   - Use cheaper models (gpt-3.5-turbo, Mistral Small) as primaries
   - Enable caching to avoid re-generation
   - Use local providers (Ollama) when possible

---

## File Format

- YAML 1.2 standard
- Must be a mapping (object), not a list or scalar
- Comments starting with `#` are supported
- Quotes optional for simple strings, required for special characters

