# Arquitetura Multi-Provedor para Ansible AI Snippet

Este documento apresenta uma análise detalhada e uma proposta de arquitetura para suportar múltiplos provedores de Inteligência Artificial no projeto `ansible-aisnippet`.

## 1️⃣ Análise da Arquitetura Atual

### Estrutura do Projeto e Componentes
O projeto possui uma estrutura simples e direta:
- `main.py`: Ponto de entrada CLI (usando `typer`). Lida com argumentos e orquestra as chamadas para geração de tarefas ou playbooks.
- `aisnippet.py`: Contém a classe principal `aisnippet`, responsável por carregar os templates (`snippets.json`), realizar buscas de similaridade (usando `gensim` e `jieba`) e comunicar-se com a API de IA.
- `helpers.py`: Funções utilitárias para manipulação de JSON e YAML (`ruamel.yaml`).
- `snippets.json`: Base de dados de templates de tarefas do Ansible.

### Provedores Atualmente Suportados
Atualmente, o sistema suporta **exclusivamente a OpenAI**, com integração direta no método `generate_task` da classe `aisnippet`. O modelo está hardcoded para `gpt-3.5-turbo-0301`.

### Padrões de Integração Existentes
A integração com a OpenAI é feita de forma síncrona, instanciando diretamente o cliente `openai.ChatCompletion.create`. A chave de API é lida da variável de ambiente `OPENAI_KEY`.

### Pontos de Extensibilidade
A arquitetura atual possui alto acoplamento com a biblioteca `openai`. O método `generate_task` concentra a lógica de IA, tornando a adição de novos provedores difícil sem modificar o núcleo da classe. O principal ponto de extensibilidade seria abstrair a chamada à API de IA para fora do fluxo principal de geração.

---

## 2️⃣ Avaliação de Qualidade do Código

### Princípios SOLID
- **S (Responsabilidade Única):** A classe `aisnippet` viola este princípio, pois mistura a lógica de busca de similaridade (NLP com Gensim/Jieba) com a comunicação da API do modelo de linguagem (OpenAI).
- **O (Aberto/Fechado):** O código atual não está aberto para extensão no que diz respeito aos provedores de IA. Para adicionar um novo provedor, seria necessário alterar o código da classe `aisnippet`.
- **D (Inversão de Dependência):** A classe de alto nível `aisnippet` depende de um módulo de baixo nível específico (`openai`).

### Reutilização e Duplicação
Há boa reutilização nos helpers, mas a lógica de chamada de API e formatação de prompts está rigidamente contida dentro do método de geração, impossibilitando seu reuso.

### Tratamento de Erros
O tratamento de erros é incipiente. Não há captura de exceções específicas (como falhas de rede, limites de taxa ou tokens inválidos da OpenAI) durante a chamada à API no método `generate_task`.

### Cobertura de Testes
Embora exista uma pasta `tests/`, não foram encontrados testes unitários implementados no repositório analisado.

### Documentação
A documentação no `README.md` explica bem o uso da ferramenta CLI, mas não detalha a estrutura interna ou como estender o código.

---

## 3️⃣ Provedores Recomendados

Para oferecer flexibilidade e opções de custo-benefício aos usuários, recomendamos o suporte aos seguintes 10 provedores/modelos:

1. **OpenAI** (Modelos GPT-3.5, GPT-4, GPT-4o) - Melhor qualidade geral.
2. **Anthropic Claude** (Modelos Opus, Sonnet, Haiku) - Excelentes para tarefas de codificação.
3. **Google Gemini** (Gemini Pro, Gemini Flash) - Integração nativa no ecossistema Google.
4. **Meta Llama 3** (via Groq ou APIs gerenciadas) - Modelos abertos de alta velocidade e qualidade.
5. **Mistral AI** (Mistral Large, Mixtral) - Modelos eficientes e competitivos.
6. **Cohere** (Command R+) - Focados em tarefas empresariais e RAG.
7. **Azure OpenAI** - Para clientes corporativos com requisitos estritos de conformidade.
8. **Ollama** (Local) - Permite rodar modelos (como Llama 3, Phi-3) localmente sem custo de API e com privacidade total.
9. **LM Studio** (Local) - Outra opção popular para execução local via servidor compativel com a API da OpenAI.
10. **HuggingFace Inference API** - Acesso a milhares de modelos open-source hospedados.

---

## 4️⃣ Padrões de Design Propostos

Para acomodar múltiplos provedores, propomos a adoção dos seguintes padrões de design:

### Provider Interface
Criar uma interface abstrata (ex: `BaseLLMProvider`) que defina o contrato que todos os provedores devem seguir (ex: um método `generate(system_prompt, user_prompt)`).

### Factory Pattern
Implementar uma `LLMProviderFactory` responsável por instanciar a classe concreta correta do provedor com base na configuração do usuário (ex: se o usuário escolher `--provider anthropic`, a factory retorna uma instância de `AnthropicProvider`).

### Strategy Pattern
A escolha do provedor na classe principal (`aisnippet`) será feita injetando a estratégia (provedor) através da interface, permitindo que a classe execute a geração sem conhecer a implementação específica da API.

### Adapter Pattern
Alguns provedores possuem APIs únicas. Adaptadores serão criados para traduzir a interface comum do nosso sistema para as chamadas específicas de cada SDK (ex: `AnthropicAdapter`, `GoogleGeminiAdapter`).

---

## 5️⃣ Melhorias Técnicas

### Configuração Centralizada
Substituir o uso direto de variáveis de ambiente (`os.getenv("OPENAI_KEY")`) em várias partes do código por uma classe de configuração centralizada (ex: usando `pydantic-settings`), que suporte arquivos `.env` e arquivos de configuração yaml/json.

### Rate Limiting e Tratamento de Erros
Implementar tratamento adequado para HTTP 429 (Too Many Requests), utilizando bibliotecas de retry com *exponential backoff* (ex: `tenacity`).

### Caching de Respostas
Implementar cache local (ex: usando `sqlite` ou em memória para a mesma execução) para evitar chamar a API repetidamente com o mesmo prompt, economizando tempo e custo.

### Fallback Strategy
Permitir a configuração de um provedor de contingência. Se a OpenAI falhar, o sistema pode automaticamente tentar o Anthropic ou um modelo local.

### Monitoramento e Logs
Adicionar logs estruturados (ex: biblioteca `logging` do Python) com níveis configuráveis (DEBUG, INFO, ERROR) para rastrear o tempo de resposta e erros de cada provedor, substituindo prints dispersos.

### Segurança de Credenciais
Garantir que as chaves de API não sejam armazenadas ou exibidas em logs sob nenhuma circunstância.

---

## 6️⃣ Exemplos de Implementação

### Criar Novo Provedor (Exemplo: Ollama Local)

```python
from abc import ABC, abstractmethod
import requests

class BaseLLMProvider(ABC):
    @abstractmethod
    def generate(self, system_message: str, user_message: str) -> str:
        pass

class OllamaProvider(BaseLLMProvider):
    def __init__(self, model_name="llama3", base_url="http://localhost:11434"):
        self.model_name = model_name
        self.base_url = base_url

    def generate(self, system_message: str, user_message: str) -> str:
        url = f"{self.base_url}/api/chat"
        payload = {
            "model": self.model_name,
            "messages": [
                {"role": "system", "content": system_message},
                {"role": "user", "content": user_message}
            ],
            "stream": False
        }
        response = requests.post(url, json=payload)
        response.raise_for_status()
        return response.json()["message"]["content"]
```

### Integrar no Sistema (Factory)

```python
class LLMProviderFactory:
    @staticmethod
    def get_provider(provider_name: str, **kwargs) -> BaseLLMProvider:
        if provider_name == "openai":
            return OpenAIProvider(**kwargs)
        elif provider_name == "ollama":
            return OllamaProvider(**kwargs)
        else:
            raise ValueError(f"Provedor {provider_name} não suportado.")
```

### Configurar Fallback e Rate Limiting

```python
from tenacity import retry, stop_after_attempt, wait_exponential

class ResilientLLMService:
    def __init__(self, primary: BaseLLMProvider, fallback: BaseLLMProvider = None):
        self.primary = primary
        self.fallback = fallback

    @retry(stop=stop_after_attempt(3), wait=wait_exponential(multiplier=1, min=2, max=10))
    def _call_primary(self, sys_msg, usr_msg):
        return self.primary.generate(sys_msg, usr_msg)

    def generate(self, sys_msg, usr_msg):
        try:
            return self._call_primary(sys_msg, usr_msg)
        except Exception as e:
            if self.fallback:
                # Log a falha e tenta o fallback
                return self.fallback.generate(sys_msg, usr_msg)
            raise e
```

---

## 7️⃣ Roadmap de Implementação

### Fase 1: Refatoração do Núcleo (Semana 1)
- Extrair a lógica de chamadas da API da classe `aisnippet`.
- Criar a interface `BaseLLMProvider`.
- Implementar o provedor atual (`OpenAIProvider`) seguindo a nova interface.
- Implementar a camada de Configuração Centralizada.

### Fase 2: Integração de Provedores Principais (Semana 2)
- Adicionar suporte a **Anthropic Claude**.
- Adicionar suporte a **Google Gemini**.
- Criar a `LLMProviderFactory` e atualizar o CLI (parâmetros `--provider` e `--model`).

### Fase 3: Suporte Local e Aberto (Semana 3)
- Adicionar suporte ao **Ollama** e **LM Studio** para execução local sem custo.
- Adicionar suporte à API de inferência da **HuggingFace** e/ou **Groq** (para Llama 3).

### Fase 4: Resiliência e Tratamento de Erros (Semana 4)
- Implementar as políticas de retry (ex: Tenacity) para todos os provedores HTTP.
- Implementar o Fallback Strategy.
- Implementar sistema de logs adequado.

### Fase 5: Otimização e Cache (Semana 5)
- Implementar o sistema de Caching de Respostas.
- Ajustar prompts para os diferentes modelos (modelos abertos podem requerer ajustes sutis em relação ao GPT).

### Fase 6: Testes, Documentação e Lançamento (Semana 6)
- Escrever testes unitários para a Factory, Adapters e lógicas de Fallback.
- Atualizar o `README.md` com instruções detalhadas para múltiplos provedores e execução local.
- Lançamento de versão (ex: v0.2.0).

---

## 8️⃣ Considerações de Custo e Performance

### Comparação de Custos
- **OpenAI / Anthropic (Tier Alto - GPT-4o / Opus):** Maior custo, mas com a mais alta precisão para tarefas complexas de automação e Ansible avançado.
- **OpenAI / Anthropic (Tier Base - GPT-3.5 / Haiku / Flash):** Custo muito baixo e geralmente suficientes para tarefas simples.
- **Groq / Llama 3:** Custo extremamente baixo (ou gratuito no tier de desenvolvimento) e altíssima velocidade.
- **Ollama / LM Studio:** Custo **zero**. Utiliza processamento local (requer CPU/GPU razoável no ambiente do usuário).

### Latência e Qualidade
- Provedores como **Groq** fornecem a menor latência possível.
- Modelos maiores via API (GPT-4, Claude Opus) podem demorar alguns segundos, o que é aceitável em processos assíncronos de CLI, mas não em pipelines contínuos severos.
- Provedores locais têm latência dependente do hardware da máquina.

### Escalabilidade
A adoção da arquitetura por interfaces permite escalar não apenas no número de provedores, mas também facilita a integração futura em pipelines de CI/CD, onde limites de taxa rigorosos exigem proxies, fallbacks ou modelos locais dedicados para não interromper a esteira de deployment.
