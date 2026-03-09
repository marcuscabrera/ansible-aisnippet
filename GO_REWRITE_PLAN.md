# Análise e Planejamento para Reescrita em Go: Ansible AISnippet

## 1. Sumário Executivo da Análise Atual

O projeto **Ansible AISnippet** é uma ferramenta de linha de comando (CLI) escrita em Python, projetada para converter descrições em linguagem natural em tarefas e playbooks do Ansible usando Inteligência Artificial.

**Componentes Principais:**
- **CLI:** Construída utilizando o framework `Typer`, que fornece uma interface amigável e suporte a autocomplete.
- **Motor de Similaridade:** Utiliza `Gensim` e `Jieba` para calcular TF-IDF e encontrar o snippet de template mais próximo à descrição do usuário, minimizando chamadas redundantes de API.
- **Integração com LLMs (Providers):** Implementa um padrão *Factory* (`ProviderFactory`) para gerenciar múltiplos provedores (OpenAI, Anthropic, Gemini, etc.). Possui um `FallbackManager` para encadeamento de tentativas em caso de falha de um provedor.
- **Gerenciamento de Estado:** Inclui cache local (`ResponseCache`) e limitação de taxa (`RateLimiter`).

**Avaliação de Qualidade e Arquitetura:**
- **Padrões de Projeto:** O uso de Factory para os provedores é uma excelente escolha, promovendo alta extensibilidade. A injeção de dependências e configurações também é bem delimitada.
- **SOLID:** O código segue bem os princípios, embora o módulo central `aisnippet.py` acumule algumas responsabilidades (motor de TF-IDF, coordenação de cache, limitação de taxa e chamadas à API) que poderiam ser mais desacopladas (violando parcialmente o *Single Responsibility Principle*).
- **Tratamento de Erros e Testes:** Utiliza `pytest` para testes e depende do tratamento de erros específico de cada SDK de provedor e do mecanismo de fallback para robustez.
- **Distribuição:** Sendo em Python, requer o uso de `pip`, `pipx` ou `poetry`, além da instalação de múltiplas dependências como Gensim, que são pesadas.

## 2. Arquitetura Proposta para a Versão em Go

A reescrita em Go focará em três grandes vantagens: **distribuição de um binário único e leve**, **alta performance/inicialização rápida**, e **concorrência**.

**Design Proposto:**
- **CLI Framework:** Substituir o Typer pelo **Cobra** (`github.com/spf13/cobra`), padrão na comunidade Go para ferramentas de linha de comando.
- **Configuração:** Utilizar o **Viper** (`github.com/spf13/viper`) para leitura de variáveis de ambiente e arquivos YAML.
- **Motor de Similaridade:** Em vez de depender de bibliotecas pesadas de Data Science como Gensim, implementar um motor de TF-IDF nativo e leve em Go, adequado para o tamanho reduzido do `snippets.json`.
- **Provedores (LLMs):** Manter o padrão Factory utilizando interfaces Go. Evitar SDKs pesados quando possível, preferindo chamadas diretas via `net/http` e `encoding/json` ou SDKs mantidos oficialmente pela comunidade (ex: `github.com/sashabaranov/go-openai`).
- **Concorrência:** Utilizar *Goroutines* e `sync.WaitGroup` para processar chamadas de API em paralelo ao ler arquivos de múltiplas tarefas (`tasks.yml`), mantendo o respeito ao `RateLimiter`.
- **Manipulação de YAML/JSON:** Usar `gopkg.in/yaml.v3` para a leitura e escrita confiável de arquivos YAML do Ansible.

## 3. Estrutura de Diretórios e Arquivos

Adotando o Go Standard Layout:

```text
ansible-aisnippet/
├── cmd/
│   └── ansible-aisnippet/
│       └── main.go                 # Ponto de entrada da aplicação
├── internal/
│   ├── cli/                        # Comandos do Cobra (root, generate, list)
│   ├── config/                     # Lógica de carregamento via Viper
│   ├── core/                       # Coordenador principal (equivalente ao aisnippet.py)
│   ├── providers/                  # Interface Provider e implementações (openai, anthropic...)
│   │   ├── factory.go
│   │   ├── fallback.go             # FallbackManager
│   │   └── openai.go
│   ├── similarity/                 # Implementação nativa de TF-IDF e tokenização
│   ├── cache/                      # Caching em memória / disco
│   └── ratelimit/                  # Lógica de Rate Limiter e Backoff
├── snippets.json                   # Banco de snippets de templates
├── go.mod
├── go.sum
└── README.md
```

## 4. Cronograma de Implementação (Roadmap)

**Tempo Estimado Total: 6 Semanas**

* **Fase 1: Setup e Fundações (1 Semana)**
  * Configuração do projeto (`go mod init`).
  * Implementação da CLI básica com Cobra.
  * Carregamento de configuração usando Viper e setup das estruturas de dados.
* **Fase 2: Motor de Similaridade e Core (1.5 Semanas)**
  * Implementação da leitura de `snippets.json`.
  * Criação de um algoritmo customizado de TF-IDF e tokenização leve em Go.
  * Lógica do Rate Limiter e Cache em memória.
* **Fase 3: Implementação de Provedores e Fallback (2 Semanas)**
  * Definição da interface `Provider`.
  * Implementação da *Factory* e do *FallbackManager*.
  * Integração com os 3 provedores principais (OpenAI, Anthropic, Gemini).
  * Adição dos provedores locais (Ollama, LM Studio).
* **Fase 4: Processamento de Playbooks, Concorrência e Testes (1.5 Semanas)**
  * Implementação de leitura e parse do `tasks.yml`.
  * Adição de concorrência com Goroutines para geração de múltiplas tarefas.
  * Escrita de Testes Unitários e de Integração.
  * Setup de CI/CD (GitHub Actions) para geração de binários multi-arquitetura com `GoReleaser`.

## 5. Melhorias e Otimizações

- **Binário Estático:** A distribuição será um único arquivo binário compilado. Não haverá mais "Dependency Hell" ou a necessidade de gerenciar ambientes virtuais Python nos servidores ou máquinas dos DevOps.
- **Processamento Concorrente:** Ao invés de iterar sequencialmente sobre um arquivo `tasks.yml`, o Go permite disparar *Goroutines* para gerar tarefas independentes em paralelo, acelerando drasticamente a geração de playbooks grandes.
- **Tempo de Inicialização (Cold Start):** Ferramentas CLI em Go inicializam quase instantaneamente em comparação com Python, o que melhora a experiência de uso no terminal.
- **Segurança de Tipos (Type Safety):** O sistema de tipagem forte do Go eliminará potenciais bugs em tempo de execução relacionados ao processamento de dicionários e listas vindos do JSON/YAML.

## 6. Considerações Finais: Desafios e Mitigações

**Desafio 1: Substituição do Gensim e Jieba**
- *Problema:* O ecossistema NLP em Go não é tão maduro quanto o do Python.
- *Mitigação:* Como o espaço de busca (os templates Ansible) é pequeno e focado, uma implementação simples de Bag-of-Words e Cosine Similarity com TF-IDF construída puramente em Go será suficiente e muito mais leve do que importar bibliotecas complexas.

**Desafio 2: Manutenção de Compatibilidade de Templates**
- *Problema:* O projeto Python usa Jinja2 e formatação flexível de dicionários via Typer/ruamel.yaml.
- *Mitigação:* Utilizar o pacote nativo `text/template` do Go acoplado com `gopkg.in/yaml.v3`. Testes de regressão rigorosos devem ser mapeados desde o início para garantir que as saídas do Go sejam idênticas às saídas da versão em Python.

**Desafio 3: Integração com múltiplos SDKs de IA**
- *Problema:* Manter SDKs de diversos provedores atualizados pode ser complexo.
- *Mitigação:* Focar nas APIs REST subjacentes (via pacote nativo `net/http` e `encoding/json`) em vez de adotar um SDK de terceiros para cada provedor, simplificando as dependências do projeto e centralizando a lógica de chamadas HTTP (retry, timeout) no próprio projeto.

**Recomendações para a Manutenção em Go:**
- Utilizar `golangci-lint` para manter um alto padrão de qualidade do código.
- Configurar o `GoReleaser` no GitHub Actions para automatizar a publicação de binários para Linux, macOS (ARM/Intel) e Windows, maximizando o alcance da ferramenta na comunidade DevOps.
