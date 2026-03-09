# ansible-aisnippet

> **Generate Ansible tasks and playbooks from plain-English descriptions using AI.**

`ansible-aisnippet` is a command-line tool that converts natural-language descriptions into
ready-to-use Ansible task YAML. It supports **multiple AI providers** (OpenAI, Anthropic,
Google Gemini, Mistral, Cohere, Azure OpenAI, Ollama, LM Studio, Llama, HuggingFace,
OpenRouter, and ZenAI), a similarity-search engine powered by a native TF-IDF implementation
to match the best Ansible snippet template, automatic retries with fallback chaining, and a
configurable response cache.

The project ships a **Go CLI binary** (`cmd/ansible-aisnippet`) with no external runtime
dependencies — just download and run. The original Python implementation remains in
`ansible_aisnippet/` for reference.

---

## Table of Contents

- [Features](#features)
- [Technologies & Frameworks](#technologies--frameworks)
- [Installation](#installation)
  - [Go CLI (recommended)](#go-cli-recommended)
  - [Python CLI](#python-cli)
- [Configuration](#configuration)
  - [Environment Variables](#environment-variables)
  - [Configuration File](#configuration-file)
- [Usage](#usage)
  - [CLI Overview](#cli-overview)
  - [Generate a Single Task](#generate-a-single-task)
  - [Generate Multiple Tasks from a File](#generate-multiple-tasks-from-a-file)
  - [Generate a Full Playbook](#generate-a-full-playbook)
  - [List Available Providers](#list-available-providers)
  - [Save Output to a File](#save-output-to-a-file)
- [Go Architecture](#go-architecture)
- [Contributing](#contributing)
- [Roadmap](#roadmap)
- [License](#license)
- [Contact & Support](#contact--support)

---

## Features

- 🤖 **Multi-provider AI support** — OpenAI, Anthropic Claude, Google Gemini, Mistral,
  Cohere, Azure OpenAI, Ollama (local), LM Studio (local), Llama, HuggingFace,
  OpenRouter, and ZenAI.
- 📝 **Single-task generation** — describe a task in plain English and get valid Ansible YAML.
- 📋 **Batch generation from a YAML file** — generate many tasks (including blocks, rescue
  sections, and `register` directives) in one command.
- 📦 **Full playbook generation** — wrap generated tasks in a complete playbook skeleton.
- 💾 **Response caching** — avoid redundant API calls; configurable TTL and cache size.
- 🔁 **Automatic fallback** — configurable provider chain so generation never silently fails.
- ⚙️ **Flexible configuration** — environment variables or a YAML config file.
- 🐚 **Shell completion** — built-in tab completion for Bash, Zsh, and Fish.
- ⚡ **Single static binary** — the Go CLI ships as a single compiled binary with no runtime
  dependencies.

---

## Technologies & Frameworks

### Go CLI (current)

| Category | Technology |
|---|---|
| Language | Go ≥ 1.21 |
| CLI framework | [Cobra](https://github.com/spf13/cobra) |
| YAML parsing | [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) |
| NLP / similarity | Native TF-IDF + Cosine Similarity (no external libraries) |
| HTTP client | Go standard library `net/http` |
| Testing | Go standard library `testing` |
| CI/CD | GitHub Actions |

### Python CLI (reference)

| Category | Technology |
|---|---|
| Language | Python ≥ 3.10 |
| CLI framework | [Typer](https://typer.tiangolo.com/) + [Rich](https://rich.readthedocs.io/) |
| NLP / similarity | [Gensim](https://radimrehurek.com/gensim/) + [Jieba](https://github.com/fxsjy/jieba) |
| YAML parsing | [ruamel.yaml](https://yaml.readthedocs.io/en/latest/) |
| Package manager | [Poetry](https://python-poetry.org/) |

---

## Installation

### Go CLI (recommended)

#### Build from source

```bash
git clone https://github.com/marcuscabrera/ansible-aisnippet.git
cd ansible-aisnippet
go build -o ansible-aisnippet ./cmd/ansible-aisnippet/
./ansible-aisnippet --help
```

Requires Go 1.21 or later. No other dependencies needed.

#### Cross-compile for your platform

```bash
# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o ansible-aisnippet-linux-amd64 ./cmd/ansible-aisnippet/

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o ansible-aisnippet-darwin-arm64 ./cmd/ansible-aisnippet/

# Windows
GOOS=windows GOARCH=amd64 go build -o ansible-aisnippet.exe ./cmd/ansible-aisnippet/
```

### Python CLI

```bash
pip install ansible-aisnippet
# or with pipx
pipx install ansible-aisnippet
# or from source
pip install poetry && poetry install
```

---

## Configuration

### Environment Variables

The simplest way to configure `ansible-aisnippet` is through environment variables.

| Variable | Description | Default |
|---|---|---|
| `AI_PROVIDER` | Active AI provider | `openai` |
| `AI_FALLBACK_PROVIDERS` | Comma-separated fallback provider list | _(none)_ |
| `AI_CACHE_ENABLED` | Enable response cache (`true`/`false`) | `true` |
| `AI_CACHE_TTL` | Cache time-to-live in seconds | `3600` |
| `AI_CACHE_MAX_SIZE` | Maximum cached items | `100` |
| `AI_RATE_LIMIT_ENABLED` | Enable rate limiter (`true`/`false`) | `true` |
| `AI_RATE_LIMIT_RPM` | Requests per minute | `60` |

Provider API keys are read from standard provider-specific variables:

| Provider | Environment Variable |
|---|---|
| OpenAI | `OPENAI_KEY` or `OPENAI_API_KEY` |
| Anthropic | `ANTHROPIC_API_KEY` |
| Google Gemini | `GOOGLE_API_KEY` |
| Azure OpenAI | `AZURE_OPENAI_KEY` + `AZURE_OPENAI_ENDPOINT` |
| Mistral | `MISTRAL_API_KEY` |
| Cohere | `COHERE_API_KEY` |
| HuggingFace | `HF_API_TOKEN` |
| OpenRouter | `OPENROUTER_API_KEY` |
| ZenAI | `ZEN_API_KEY` |
| Ollama | `OLLAMA_BASE_URL` (default: `http://localhost:11434`) |
| LM Studio | `LMSTUDIO_BASE_URL` (default: `http://localhost:1234/v1`) |
| Llama | `LLAMA_BASE_URL` (default: `http://localhost:11434`) |

```bash
# Example: use OpenAI
export AI_PROVIDER=openai
export OPENAI_KEY=sk-...

# Example: use Anthropic with OpenAI as fallback
export AI_PROVIDER=anthropic
export ANTHROPIC_API_KEY=sk-ant-...
export AI_FALLBACK_PROVIDERS=openai
export OPENAI_KEY=sk-...
```

### Configuration File

For more control, pass a YAML configuration file with `--config`:

```bash
ansible-aisnippet --config config.yml generate "install nginx"
```

A ready-to-use starter template is provided at [`config.example.yml`](config.example.yml).

#### Complete configuration reference

The table below lists every supported field. All fields are optional and fall back to
sensible defaults when omitted.

| Field | Type | Default | Description |
|---|---|---|---|
| `provider` | string | `openai` | Primary AI provider |
| `fallback_providers` | list | `[]` | Backup providers tried in order when the primary fails |
| `cache.enabled` | bool | `true` | Enable response caching |
| `cache.ttl` | int | `3600` | Cache time-to-live in seconds |
| `cache.max_size` | int | `100` | Maximum number of cached responses |
| `rate_limit.enabled` | bool | `true` | Enable request rate limiter |
| `rate_limit.requests_per_minute` | int | `60` | Max requests per minute |
| `providers.<name>.api_key` | string | _(env var)_ | Provider API key; falls back to the provider-specific env variable |
| `providers.<name>.model` | string | _(provider default)_ | Model identifier |
| `providers.<name>.base_url` | string | _(provider default)_ | Custom API endpoint (required for Azure, optional for self-hosted) |
| `providers.<name>.temperature` | float | `0.0` | Response randomness (`0.0` = deterministic) |
| `providers.<name>.max_retries` | int | `3` | Retry attempts on failure |
| `providers.<name>.timeout` | int | `30` | Request timeout in seconds |
| `providers.<name>.extra` | map | `{}` | Provider-specific extra fields (e.g. `api_version` for Azure) |

#### Minimal example (single provider)

```yaml
# config.yml
provider: openai

providers:
  openai:
    api_key: sk-...
    model: gpt-4o
```

#### Full example (all 12 providers)

```yaml
# config.yml — complete example showing every provider and field.
# Only configure the providers you intend to use; omit the rest.

# ── Global ────────────────────────────────────────────────────────────────────
provider: openai                     # Primary provider (default: openai)
fallback_providers:                  # Tried in order when the primary fails
  - anthropic
  - ollama

# ── Cache ─────────────────────────────────────────────────────────────────────
cache:
  enabled: true                      # Enable response cache (default: true)
  ttl: 3600                          # Seconds to keep cached responses (default: 3600)
  max_size: 100                      # Max number of cached items (default: 100)

# ── Rate limiter ──────────────────────────────────────────────────────────────
rate_limit:
  enabled: true                      # Enable rate limiter (default: true)
  requests_per_minute: 60            # Max requests per minute (default: 60)

# ── Per-provider settings ─────────────────────────────────────────────────────
providers:

  # OpenAI — https://platform.openai.com/
  # API key env: OPENAI_KEY or OPENAI_API_KEY
  openai:
    api_key: sk-...
    model: gpt-4o                    # Default: gpt-3.5-turbo
    temperature: 0.0                 # Default: 0.0 (deterministic)
    max_retries: 3                   # Default: 3
    timeout: 30                      # Default: 30 seconds

  # Anthropic — https://www.anthropic.com/
  # API key env: ANTHROPIC_API_KEY
  anthropic:
    api_key: sk-ant-...
    model: claude-3-haiku-20240307   # Default: claude-3-haiku-20240307
    temperature: 0.0
    max_retries: 3
    timeout: 60

  # Google Gemini — https://ai.google.dev/
  # API key env: GOOGLE_API_KEY
  google:
    api_key: AIza...
    model: gemini-pro                # Default: gemini-pro
    temperature: 0.0
    max_retries: 3
    timeout: 30

  # Azure OpenAI Service — https://azure.microsoft.com/en-us/products/ai-services/openai-service
  # API key env: AZURE_OPENAI_KEY
  # Endpoint env: AZURE_OPENAI_ENDPOINT
  # Deployment env: AZURE_OPENAI_DEPLOYMENT
  azure:
    api_key: your-azure-api-key
    base_url: https://your-resource.openai.azure.com/
    model: your-deployment-name      # Default: gpt-35-turbo
    temperature: 0.0
    max_retries: 3
    timeout: 30
    extra:
      api_version: "2024-02-01"      # Azure API version (default: 2024-02-01)

  # Mistral AI — https://mistral.ai/
  # API key env: MISTRAL_API_KEY
  mistral:
    api_key: your-mistral-key
    model: mistral-small             # Default: mistral-small
    temperature: 0.0
    max_retries: 3
    timeout: 30

  # Cohere — https://cohere.com/
  # API key env: COHERE_API_KEY
  cohere:
    api_key: your-cohere-key
    model: command                   # Default: command
    temperature: 0.0
    max_retries: 3
    timeout: 30

  # Ollama (local) — https://ollama.com/
  # Base URL env: OLLAMA_BASE_URL
  # Model env:    OLLAMA_MODEL
  ollama:
    base_url: http://localhost:11434  # Default: http://localhost:11434
    model: llama3                     # Default: llama3
    temperature: 0.0
    max_retries: 3
    timeout: 120                      # Longer timeout recommended for local inference

  # LM Studio (local) — https://lmstudio.ai/
  # Base URL env: LMSTUDIO_BASE_URL
  # Model env:    LMSTUDIO_MODEL
  lmstudio:
    base_url: http://localhost:1234/v1  # Default: http://localhost:1234/v1
    model: local-model                  # Default: local-model
    temperature: 0.0
    max_retries: 3
    timeout: 120

  # Meta Llama via Ollama-compatible endpoint (local)
  # Base URL env: LLAMA_BASE_URL
  # Model env:    LLAMA_MODEL
  llama:
    base_url: http://localhost:11434  # Default: http://localhost:11434
    model: llama3                     # Default: llama3
    temperature: 0.0
    max_retries: 3
    timeout: 120

  # HuggingFace Inference API — https://huggingface.co/inference-api
  # API token env: HF_API_TOKEN
  # Model env:     HF_MODEL
  huggingface:
    api_key: hf_...
    model: mistralai/Mistral-7B-Instruct-v0.2  # Default: mistralai/Mistral-7B-Instruct-v0.2
    temperature: 0.1                            # HuggingFace requires temperature > 0
    max_retries: 3
    timeout: 60

  # OpenRouter — https://openrouter.ai/
  # API key env: OPENROUTER_API_KEY
  openrouter:
    api_key: sk-or-...
    model: openai/gpt-3.5-turbo      # Default: openai/gpt-3.5-turbo
    temperature: 0.0
    max_retries: 3
    timeout: 30

  # ZenAI (OpenCode) — https://api.opencode.ai/
  # API key env: ZEN_API_KEY
  zen:
    api_key: your-zen-key
    model: zen                       # Default: zen
    temperature: 0.0
    max_retries: 3
    timeout: 30
```

---

## Usage

### CLI Overview

```
ansible-aisnippet --help

ansible-aisnippet converts natural-language task descriptions into
Ansible tasks and playbooks by querying AI language models.

Usage:
  ansible-aisnippet [command]

Available Commands:
  completion     Generate the autocompletion script for the specified shell
  generate       Generate an Ansible task or playbook from a description
  help           Help about any command
  list-providers List all available AI providers

Flags:
  -h, --help      help for ansible-aisnippet
  -v, --version   version for ansible-aisnippet
```

### Generate a Single Task

```bash
export AI_PROVIDER=openai
export OPENAI_KEY=sk-...

ansible-aisnippet generate "execute command to start /opt/application/start.sh"
```

You can also select a provider directly on the command line:

```bash
ansible-aisnippet generate --provider anthropic "install and enable nginx"
```

### Generate Multiple Tasks from a File

Create a YAML file describing your tasks:

```yaml
# tasks.yml
- task: Install package htop, nginx and net-tools with generic module
- task: Copy file from local file /tmp/toto to remote /tmp/titi set mode 0666 owner bob group www
  register: test
- name: A block
  when: test.rc == 0
  block:
    - task: wait for port 6300 on localhost timeout 25
  rescue:
    - task: Execute command /opt/application/start.sh creates /var/run/test.lock
- task: Download file from https://tmp.io/test/ set mode 0640 and force true
```

Then run:

```bash
ansible-aisnippet generate -f tasks.yml
```

### Generate a Full Playbook

Add the `--playbook` / `-p` flag to wrap all tasks in a complete playbook:

```bash
ansible-aisnippet generate -f tasks.yml -p
```

### List Available Providers

```bash
ansible-aisnippet list-providers
```

Output:

```
Available AI providers:
  • anthropic
  • azure
  • cohere
  • google
  • huggingface
  • llama
  • lmstudio
  • mistral
  • ollama
  • openai
  • openrouter
  • zen
```

### Save Output to a File

```bash
ansible-aisnippet generate -f tasks.yml -p -o playbook.yml
```

---

## Go Architecture

The Go implementation follows the [Go Standard Layout](https://github.com/golang-standards/project-layout):

```
ansible-aisnippet/
├── cmd/
│   └── ansible-aisnippet/
│       └── main.go                 # Entry point
├── internal/
│   ├── cli/                        # Cobra CLI commands (root, generate, list-providers)
│   ├── config/                     # Configuration loading from env/YAML
│   ├── core/                       # Main coordinator (similarity + provider + cache)
│   │   └── data/
│   │       └── snippets.json       # Embedded Ansible snippet templates
│   ├── providers/                  # Provider interface + all 12 implementations
│   │   ├── provider.go             # Provider interface (Strategy Pattern)
│   │   ├── factory.go              # ProviderFactory with registry
│   │   ├── fallback.go             # FallbackManager
│   │   ├── openai.go
│   │   ├── anthropic.go
│   │   ├── google.go
│   │   ├── azure.go
│   │   ├── mistral.go
│   │   ├── cohere.go
│   │   ├── ollama.go
│   │   ├── lmstudio.go
│   │   ├── llama.go
│   │   ├── huggingface.go
│   │   ├── openrouter.go
│   │   └── zen.go
│   ├── similarity/                 # Native TF-IDF engine (replaces Gensim/Jieba)
│   ├── cache/                      # In-memory response cache with TTL/LRU eviction
│   └── ratelimit/                  # Sliding-window rate limiter
├── go.mod
├── go.sum
└── README.md
```

### Key Design Decisions

- **Strategy + Factory Pattern** — `Provider` interface with `ProviderFactory` registry allows
  adding new providers without touching existing code.
- **Native TF-IDF** — A pure-Go Bag-of-Words / Cosine Similarity engine replaces the heavy
  Gensim + Jieba Python dependencies. It is sufficient for the small English-language
  `snippets.json` corpus.
- **Direct HTTP calls** — All provider implementations use the standard `net/http` package
  instead of per-provider SDKs, keeping the dependency list minimal.
- **Embedded snippets** — `snippets.json` is embedded into the binary at compile time via
  `//go:embed`, producing a single self-contained executable.

---

## Contributing

Contributions of all kinds are welcome — bug reports, feature requests, documentation
improvements, and code changes.

### Getting Started (Go)

1. **Fork** the repository and clone your fork.
2. Create a new branch from `main`:
   ```bash
   git checkout -b feat/my-feature
   ```
3. Build and test:
   ```bash
   go build ./...
   go test ./...
   go vet ./...
   ```

### Adding a New AI Provider (Go)

1. Create `internal/providers/<name>.go` implementing the `Provider` interface.
2. Register the constructor in the `init()` function in `internal/providers/factory.go`.
3. Add tests to `internal/providers/providers_test.go`.

### Getting Started (Python)

```bash
pip install poetry
poetry install
poetry run pytest
```

### Pull Request Process

1. Ensure all tests pass and new code has adequate coverage.
2. Update `README.md` and `changelog.md` if your change is user-facing.
3. Open a pull request against `main` with a clear description.

---

## Roadmap

- **v0.2** — Go CLI with all 12 providers, native TF-IDF, single binary distribution.
- **v0.3** — GoReleaser integration for automatic multi-architecture releases (Linux, macOS
  ARM/Intel, Windows) on every tagged release.
- **v0.4** — Interactive mode; Ansible role scaffolding.
- **v0.5** — Plugin support for custom Ansible collections.

See [GO_REWRITE_PLAN.md](GO_REWRITE_PLAN.md) for the detailed technical roadmap and
[MULTI_PROVIDER_ARCHITECTURE.md](MULTI_PROVIDER_ARCHITECTURE.md) for the multi-provider
design rationale.

---

## License

This project is distributed under the **Apache License 2.0**.
See the [LICENSE](LICENSE) file for the full text.

---

## Contact & Support

| Channel | Link |
|---|---|
| 🐛 Bug reports & feature requests | [GitHub Issues](https://github.com/marcuscabrera/ansible-aisnippet/issues) |
| 💬 General questions & discussions | [GitHub Discussions](https://github.com/marcuscabrera/ansible-aisnippet/discussions) |
