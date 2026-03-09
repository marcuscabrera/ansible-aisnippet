# ansible-aisnippet

> **Generate Ansible tasks and playbooks from plain-English descriptions using AI.**

`ansible-aisnippet` is a command-line tool that converts natural-language descriptions into
ready-to-use Ansible task YAML. It supports **multiple AI providers** (OpenAI, Anthropic,
Google Gemini, Mistral, Cohere, Azure OpenAI, Ollama, LM Studio, Llama, and HuggingFace),
a similarity-search cache powered by Gensim/Jieba to avoid redundant API calls, automatic
retries with exponential back-off, and a configurable fallback chain so generation never
silently fails.

---

## Table of Contents

- [Features](#features)
- [Technologies & Frameworks](#technologies--frameworks)
- [Installation](#installation)
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
- [Contributing](#contributing)
- [Roadmap](#roadmap)
- [License](#license)
- [Contact & Support](#contact--support)

---

## Features

- 🤖 **Multi-provider AI support** — OpenAI, Anthropic Claude, Google Gemini, Mistral,
  Cohere, Azure OpenAI, Ollama (local), LM Studio (local), Llama, and HuggingFace.
- 📝 **Single-task generation** — describe a task in plain English and get valid Ansible YAML.
- 📋 **Batch generation from a YAML file** — generate many tasks (including blocks, rescue
  sections, and `register` directives) in one command.
- 📦 **Full playbook generation** — wrap generated tasks in a complete playbook skeleton.
- 💾 **Response caching** — avoid redundant API calls; configurable TTL and cache size.
- 🔁 **Automatic retries & fallback** — configurable retry count with exponential back-off
  and a fallback provider chain.
- ⚙️ **Flexible configuration** — environment variables or a YAML config file.
- 🐚 **Shell completion** — built-in tab completion for Bash, Zsh, and Fish.

---

## Technologies & Frameworks

| Category | Technology |
|---|---|
| Language | Python ≥ 3.10 |
| CLI framework | [Typer](https://typer.tiangolo.com/) + [Rich](https://rich.readthedocs.io/) |
| AI / LLM SDKs | `openai`, `anthropic`, `google-generativeai`, `mistralai`, `cohere` |
| NLP / similarity | [Gensim](https://radimrehurek.com/gensim/) + [Jieba](https://github.com/fxsjy/jieba) |
| YAML parsing | [ruamel.yaml](https://yaml.readthedocs.io/en/latest/) |
| Templating | [Jinja2](https://jinja.palletsprojects.com/) |
| Package manager | [Poetry](https://python-poetry.org/) |
| Testing | [pytest](https://pytest.org/) |
| CI/CD | GitHub Actions |

---

## Installation

### Using pip (recommended)

```bash
pip install ansible-aisnippet
```

### Using pipx (isolated environment)

```bash
pipx install ansible-aisnippet
```

### From source (development)

```bash
git clone https://github.com/marcuscabrera/ansible-aisnippet.git
cd ansible-aisnippet
pip install poetry
poetry install
```

### Dependencies

All runtime dependencies are managed by Poetry and installed automatically. The key ones are:

- `python ^3.10`
- `typer[all] ^0.7`
- `Jinja2 ^3.1`
- `ruamel.yaml ^0.17`
- `rich ^12.6`
- `gensim ^4.3`
- `jieba ^0.42`
- `openai >=0.27`
- `requests ^2.28`

Optional provider SDKs (install as needed):

```bash
pip install anthropic              # Anthropic Claude
pip install google-generativeai    # Google Gemini
pip install mistralai              # Mistral AI
pip install cohere                 # Cohere
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
| OpenAI | `OPENAI_KEY` |
| Anthropic | `ANTHROPIC_API_KEY` |
| Google Gemini | `GOOGLE_API_KEY` |
| Mistral | `MISTRAL_API_KEY` |
| Cohere | `COHERE_API_KEY` |
| Azure OpenAI | `AZURE_OPENAI_API_KEY` |
| HuggingFace | `HUGGINGFACE_API_KEY` |
| OpenRouter | `OPENROUTER_API_KEY` |

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

```yaml
# config.yml
provider: openai
fallback_providers:
  - anthropic
  - ollama

cache:
  enabled: true
  ttl: 3600
  max_size: 100

rate_limit:
  enabled: true
  requests_per_minute: 60

providers:
  openai:
    api_key: sk-...
    model: gpt-4o
    temperature: 0.0
    max_retries: 3
    timeout: 30
  anthropic:
    api_key: sk-ant-...
    model: claude-3-haiku-20240307
  ollama:
    base_url: http://localhost:11434
    model: llama3
```

```bash
ansible-aisnippet --config config.yml generate "install nginx"
```

---

## Usage

### CLI Overview

```
ansible-aisnippet --help

 Usage: ansible-aisnippet [OPTIONS] COMMAND [ARGS]...

╭─ Options ─────────────────────────────────────────────────────╮
│ --version             -v        Show the application's        │
│                                 version and exit.             │
│ --install-completion            Install completion for the    │
│                                 current shell.                │
│ --show-completion               Show completion for the       │
│                                 current shell, to copy it or  │
│                                 customize the installation.   │
│ --help                          Show this message and exit.   │
╰───────────────────────────────────────────────────────────────╯
╭─ Commands ────────────────────────────────────────────────────╮
│ generate        Ask an AI provider to write an Ansible task   │
│                 using a template                              │
│ list-providers  List all available AI providers               │
╰───────────────────────────────────────────────────────────────╯
```

### Generate a Single Task

```bash
export AI_PROVIDER=openai
export OPENAI_KEY=sk-...

ansible-aisnippet generate "execute command to start /opt/application/start.sh create /var/run/test.lock"
```

Example output:

```yaml
name: Execute command to start /opt/application/start.sh create /var/run/test.lock
ansible.builtin.command:
  chdir: /opt/application
  cmd: ./start.sh && touch /var/run/test.lock
  creates: /var/run/test.lock
  removes: ''
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

Example output:

```yaml
- name: Playbook generated with AI
  hosts: all
  gather_facts: true
  tasks:
  - name: Install package htop, nginx and net-tools
    ansible.builtin.yum:
      name:
      - htop
      - nginx
      - net-tools
      state: present
  - name: Copy file from local file /tmp/toto to remote /tmp/titi
    ansible.builtin.copy:
      src: /tmp/toto
      dest: /tmp/titi
      mode: '0666'
      owner: bob
      group: www
    register: test
  - name: A block
    when: test.rc == 0
    block:
    - name: Wait for port 6300 on localhost timeout 25
      ansible.builtin.wait_for:
        host: 127.0.0.1
        port: '6300'
        timeout: '25'
    rescue:
    - name: Execute command /opt/application/start.sh creates /var/run/test.lock
      ansible.builtin.command:
        chdir: /tmp/test
        cmd: /opt/application/start.sh
        creates: /var/run/test.lock
  - name: Download file from https://tmp.io/test/
    ansible.builtin.get_url:
      backup: false
      decompress: true
      dest: /tmp/test
      force: true
      group: root
      mode: '0640'
      owner: root
      timeout: '10'
      tmp_dest: /tmp/test
      url: https://tmp.io/test/
      validate_certs: true
```

### List Available Providers

```bash
ansible-aisnippet list-providers
```

### Save Output to a File

```bash
ansible-aisnippet generate -f tasks.yml -p -o playbook.yml
```

---

## Contributing

Contributions of all kinds are welcome — bug reports, feature requests, documentation
improvements, and code changes.

### Getting Started

1. **Fork** the repository and clone your fork.
2. Create a new branch from `main`:
   ```bash
   git checkout -b feat/my-feature
   ```
3. Install development dependencies:
   ```bash
   pip install poetry
   poetry install
   ```
4. Make your changes, add tests, and verify everything passes:
   ```bash
   poetry run pytest
   ```

### Code Style

- Follow [PEP 8](https://pep8.org/) for Python style.
- Format code with [Black](https://black.readthedocs.io/) before committing:
  ```bash
  poetry run black ansible_aisnippet tests
  ```
- Keep lines ≤ 88 characters (Black default).
- Add docstrings to all public classes and methods.
- Use type hints throughout.

### Adding a New AI Provider

1. Create `ansible_aisnippet/providers/<name>_provider.py` implementing `BaseProvider`.
2. Register the provider in `ansible_aisnippet/providers/factory.py`.
3. Add the provider name to the `--provider` help text in `main.py`.
4. Add a test in `tests/test_providers.py`.

### Pull Request Process

1. Ensure all existing tests pass and new code has adequate test coverage.
2. Update the `README.md` and `changelog.md` if your change affects user-facing behaviour.
3. Open a pull request against `main` with a clear description of the problem and
   the solution, referencing any related issues.
4. Address any review comments promptly.

### Reporting Bugs

Please open a [GitHub Issue](https://github.com/marcuscabrera/ansible-aisnippet/issues)
with steps to reproduce, expected behaviour, actual behaviour, and your environment
details (OS, Python version, `ansible-aisnippet` version).

---

## Roadmap

The following features and improvements are planned for upcoming releases:

- **v0.2** — Extended provider support (OpenRouter, ZenAI), improved prompt engineering
  per provider, and full test coverage for all providers.
- **v0.3** — Interactive mode that lets users review and refine generated tasks before
  saving; richer verbose output showing token usage.
- **v0.4** — Ansible role scaffolding — generate a complete role directory structure
  (tasks, handlers, defaults, meta) from a description.
- **v0.5** — Plugin support for custom Ansible collections; export to AWX/AAP Job
  Templates.
- **Ongoing** — Performance improvements to the similarity-search cache, documentation
  translations, and community-contributed provider adapters.

See [MULTI_PROVIDER_ARCHITECTURE.md](MULTI_PROVIDER_ARCHITECTURE.md) for the detailed
technical roadmap and design rationale.

---

## License

This project is distributed under the **Apache License 2.0**.
See the [LICENSE](LICENSE) file for the full text.

```
Copyright 2024 ansible-aisnippet contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

---

## Contact & Support

| Channel | Link |
|---|---|
| 🐛 Bug reports & feature requests | [GitHub Issues](https://github.com/marcuscabrera/ansible-aisnippet/issues) |
| 💬 General questions & discussions | [GitHub Discussions](https://github.com/marcuscabrera/ansible-aisnippet/discussions) |
| 📖 Author's blog | [blog.stephane-robert.info](https://blog.stephane-robert.info/) |
| 📧 Author e-mail | robert.stephane.28@gmail.com |
