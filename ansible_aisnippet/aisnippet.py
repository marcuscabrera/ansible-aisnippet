from __future__ import annotations

from gensim import corpora, models, similarities
from rich import print

from .cache import ResponseCache
from .config import Config
from .fallback import FallbackManager
from .helpers import convert_to_yaml, escape_json, find_keys
from .providers.factory import ProviderFactory
from .rate_limiter import RateLimiter

import jieba
import json
import os


class aisnippet:
    def __init__(self, **kwargs):
        """Initialise the snippet generator.

        Keyword Args:
            verbose (bool):   Print extra debug information.
            outputfile (str): Path for saving generated output.
            playbook (bool):  Wrap tasks in a playbook structure.
            config (Config):  Pre-built :class:`~ansible_aisnippet.config.Config`
                              instance.  When omitted a default config is built
                              from environment variables.
            provider (str):   Override the active provider name.
        """
        self.verbose = kwargs.get("verbose")
        self.outputfile = kwargs.get("outputfile")
        self.playbook = kwargs.get("playbook")
        self.opts = kwargs

        # Configuration ----------------------------------------------------
        self.config: Config = kwargs.get("config") or Config.from_env()
        if kwargs.get("provider"):
            self.config.provider = kwargs["provider"]

        # TF-IDF template matching -----------------------------------------
        self.dirpath = os.getcwd()
        self.snippets = self.__load_snippets__(
            os.path.join(os.path.dirname(__file__), "snippets.json")
        )
        self.analyzed_snippets = [
            jieba.lcut(snippet.lower()) for snippet in self.snippets
        ]
        self.dictionary = corpora.Dictionary(self.analyzed_snippets)
        self.corpus = [
            self.dictionary.doc2bow(snippet) for snippet in self.analyzed_snippets
        ]
        self.tfidf = models.TfidfModel(self.corpus)
        self.feature_cnt = len(self.dictionary.token2id)

        # Infrastructure ---------------------------------------------------
        cache_cfg = self.config.cache
        self._cache = (
            ResponseCache(ttl=cache_cfg.ttl, max_size=cache_cfg.max_size)
            if cache_cfg.enabled
            else None
        )

        rl_cfg = self.config.rate_limit
        self._rate_limiter = (
            RateLimiter(requests_per_minute=rl_cfg.requests_per_minute)
            if rl_cfg.enabled
            else None
        )

        self._fallback_manager: FallbackManager | None = self.__build_fallback__()

    # ------------------------------------------------------------------
    # Private helpers
    # ------------------------------------------------------------------

    def __load_snippets__(self, file):
        with open(file) as json_file:
            return json.load(json_file)

    def __find_similar__(self, text):
        kw_vector = self.dictionary.doc2bow(jieba.lcut(text))
        index = similarities.SparseMatrixSimilarity(
            self.tfidf[self.corpus], num_features=self.feature_cnt
        )
        sim = list(index[self.tfidf[kw_vector]])
        idx = sim.index(max(sim))
        return self.snippets[list(self.snippets)[idx]]

    def __build_fallback__(self) -> FallbackManager | None:
        """Build a :class:`FallbackManager` from the current config."""
        all_provider_names = [self.config.provider] + (
            self.config.fallback_providers or []
        )
        # Deduplicate while preserving order
        seen: set[str] = set()
        unique_names = [
            n for n in all_provider_names if not (n in seen or seen.add(n))
        ]
        if len(unique_names) <= 1:
            return None  # No fallback needed; use direct provider call

        providers = [
            ProviderFactory.create(name, self.config.get_provider_config(name))
            for name in unique_names
        ]
        return FallbackManager(providers)

    def __call_provider__(self, system_message: str, user_message: str) -> str:
        """Route the request through cache → rate limiter → provider/fallback."""
        provider_name = self.config.provider

        # Check cache
        if self._cache is not None:
            cached = self._cache.get(provider_name, system_message, user_message)
            if cached is not None:
                if self.verbose:
                    print("[cache] Returning cached response.")
                return cached

        # Apply rate limiting
        if self._rate_limiter is not None:
            self._rate_limiter.acquire()

        # Generate response
        if self._fallback_manager is not None:
            response_text, used_provider = self._fallback_manager.generate(
                system_message, user_message
            )
            if self.verbose and used_provider != provider_name:
                print(f"[fallback] Used provider: {used_provider}")
        else:
            provider_cfg = self.config.get_provider_config(provider_name)
            provider = ProviderFactory.create(provider_name, provider_cfg)
            response_text = provider.generate(system_message, user_message)

        # Store in cache
        if self._cache is not None:
            self._cache.set(provider_name, system_message, user_message, response_text)

        return response_text

    # ------------------------------------------------------------------
    # Public API (unchanged signature for backward compatibility)
    # ------------------------------------------------------------------

    def generate_task(self, text):
        """Generate a single Ansible task for the given natural-language *text*."""
        snippet = self.__find_similar__(text)
        if self.verbose:
            print(snippet)

        system_message = (
            "You are an Ansible expert. Use ansible FQCN. No comment. Json:"
        )
        user_message = (
            "You have to generate an ansible task with name %s using all the "
            "options of the provided template #template 1 %s. No comment. json:"
            % (text.capitalize(), snippet)
        )

        raw_response = self.__call_provider__(system_message, user_message)

        result = {}
        parsed = json.loads(escape_json(raw_response))
        if "tasks" in parsed:
            result = parsed["tasks"]
        else:
            result = parsed
        if self.verbose:
            print(convert_to_yaml(result))
        if isinstance(result, list):
            return result[0]
        return result

    def generate_tasks(self, tasks):
        """Recursively generate Ansible tasks from a list of task descriptors."""
        output_tasks = []
        for d in tasks:
            if "task" in d:
                result = self.generate_task(d["task"])
                if "register" in d:
                    result["register"] = d["register"]
                output_tasks.append(result)
            else:
                if "block" in d:
                    block = {}
                    if "name" in d:
                        block["name"] = d["name"]
                    if "when" in d:
                        block["when"] = d["when"]
                    block["block"] = self.generate_tasks(d["block"])
                if "rescue" in d:
                    block["rescue"] = self.generate_tasks(d["rescue"])
                if "always" in d:
                    block["always"] = self.generate_tasks(d["always"])
                output_tasks.append(block)
        return output_tasks
