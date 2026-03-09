// Package core is the main coordinator for ansible-aisnippet.
// It orchestrates snippet similarity matching, rate limiting, caching, and
// provider calls to generate Ansible tasks from natural-language descriptions.
package core

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/marcuscabrera/ansible-aisnippet/internal/cache"
	"github.com/marcuscabrera/ansible-aisnippet/internal/config"
	"github.com/marcuscabrera/ansible-aisnippet/internal/providers"
	"github.com/marcuscabrera/ansible-aisnippet/internal/ratelimit"
	"github.com/marcuscabrera/ansible-aisnippet/internal/similarity"
)

//go:embed data/snippets.json
var snippetsJSON []byte

// TaskDescriptor describes a single task entry in a tasks YAML file.
type TaskDescriptor struct {
	Task    string           `yaml:"task"`
	Name    string           `yaml:"name"`
	When    string           `yaml:"when"`
	Register string          `yaml:"register"`
	Block   []TaskDescriptor `yaml:"block"`
	Rescue  []TaskDescriptor `yaml:"rescue"`
	Always  []TaskDescriptor `yaml:"always"`
}

// AISnippet is the main coordinator that handles snippet matching, caching,
// rate limiting, and provider calls.
type AISnippet struct {
	cfg         *config.Config
	verbose     bool
	snippets    map[string]interface{}
	snippetKeys []string
	simEngine   *similarity.Engine
	cache       *cache.ResponseCache
	rateLimiter *ratelimit.RateLimiter
	fallback    *providers.FallbackManager
}

// Options configures an AISnippet instance.
type Options struct {
	Verbose  bool
	Config   *config.Config
	Provider string // overrides Config.Provider when non-empty
}

// New creates an AISnippet instance from the given options.
func New(opts Options) (*AISnippet, error) {
	cfg := opts.Config
	if cfg == nil {
		cfg = config.FromEnv()
	}
	if opts.Provider != "" {
		cfg.Provider = opts.Provider
	}

	// Load snippets.
	var raw map[string]interface{}
	if err := json.Unmarshal(snippetsJSON, &raw); err != nil {
		return nil, fmt.Errorf("loading snippets: %w", err)
	}
	keys := make([]string, 0, len(raw))
	for k := range raw {
		keys = append(keys, k)
	}

	// Build similarity engine.
	simEngine := similarity.New(keys)

	// Build cache.
	var responseCache *cache.ResponseCache
	if cfg.Cache.Enabled {
		responseCache = cache.New(cfg.Cache.TTL, cfg.Cache.MaxSize)
	}

	// Build rate limiter.
	var rateLimiter *ratelimit.RateLimiter
	if cfg.RateLimit.Enabled {
		rateLimiter = ratelimit.New(cfg.RateLimit.RequestsPerMinute)
	}

	// Build fallback manager.
	var fallbackMgr *providers.FallbackManager
	allProviderNames := deduplicateOrder(append([]string{cfg.Provider}, cfg.FallbackProviders...))
	if len(allProviderNames) > 1 {
		provs := make([]providers.Provider, 0, len(allProviderNames))
		for _, name := range allProviderNames {
			pcfg := cfg.GetProviderConfig(name)
			p, err := providers.Create(name, pcfg)
			if err != nil {
				return nil, fmt.Errorf("building fallback provider %q: %w", name, err)
			}
			provs = append(provs, p)
		}
		mgr, err := providers.NewFallbackManager(provs)
		if err != nil {
			return nil, err
		}
		fallbackMgr = mgr
	}

	return &AISnippet{
		cfg:         cfg,
		verbose:     opts.Verbose,
		snippets:    raw,
		snippetKeys: keys,
		simEngine:   simEngine,
		cache:       responseCache,
		rateLimiter: rateLimiter,
		fallback:    fallbackMgr,
	}, nil
}

// findSimilarSnippet returns the snippet template value most similar to text.
func (a *AISnippet) findSimilarSnippet(text string) interface{} {
	key := a.simEngine.FindMostSimilar(text)
	if a.verbose {
		fmt.Printf("[similarity] matched snippet key: %q\n", key)
	}
	return a.snippets[key]
}

// callProvider routes the request through cache → rate limiter → provider/fallback.
func (a *AISnippet) callProvider(systemMessage, userMessage string) (string, error) {
	providerName := a.cfg.Provider

	// Check cache.
	if a.cache != nil {
		if cached, ok := a.cache.Get(providerName, systemMessage, userMessage); ok {
			if a.verbose {
				fmt.Println("[cache] Returning cached response.")
			}
			return cached, nil
		}
	}

	// Apply rate limiting.
	if a.rateLimiter != nil {
		a.rateLimiter.Acquire()
	}

	// Generate response.
	var responseText string
	if a.fallback != nil {
		result, usedProvider, err := a.fallback.Generate(systemMessage, userMessage)
		if err != nil {
			return "", err
		}
		if a.verbose && usedProvider != providerName {
			fmt.Printf("[fallback] Used provider: %s\n", usedProvider)
		}
		responseText = result
	} else {
		pcfg := a.cfg.GetProviderConfig(providerName)
		p, err := providers.Create(providerName, pcfg)
		if err != nil {
			return "", err
		}
		result, err := p.Generate(systemMessage, userMessage)
		if err != nil {
			return "", err
		}
		responseText = result
	}

	// Store in cache.
	if a.cache != nil {
		a.cache.Set(providerName, systemMessage, userMessage, responseText)
	}

	return responseText, nil
}

// GenerateTask generates a single Ansible task for the given natural-language text.
// It returns the task as a map ready for YAML marshalling.
func (a *AISnippet) GenerateTask(text string) (map[string]interface{}, error) {
	snippet := a.findSimilarSnippet(text)
	snippetJSON, err := json.Marshal(snippet)
	if err != nil {
		return nil, fmt.Errorf("marshalling snippet: %w", err)
	}

	systemMessage := "You are an Ansible expert. Use ansible FQCN. No comment. Json:"
	userMessage := fmt.Sprintf(
		"You have to generate an ansible task with name %s using all the "+
			"options of the provided template #template 1 %s. No comment. json:",
		capitalize(text), string(snippetJSON),
	)

	rawResponse, err := a.callProvider(systemMessage, userMessage)
	if err != nil {
		return nil, err
	}

	return parseTaskResponse(rawResponse)
}

// GenerateTasks recursively generates Ansible tasks from a list of TaskDescriptors.
func (a *AISnippet) GenerateTasks(tasks []TaskDescriptor) ([]map[string]interface{}, error) {
	var output []map[string]interface{}
	for _, d := range tasks {
		if d.Task != "" {
			result, err := a.GenerateTask(d.Task)
			if err != nil {
				return nil, err
			}
			if d.Register != "" {
				result["register"] = d.Register
			}
			output = append(output, result)
		} else {
			block := make(map[string]interface{})
			if d.Name != "" {
				block["name"] = d.Name
			}
			if d.When != "" {
				block["when"] = d.When
			}
			if len(d.Block) > 0 {
				inner, err := a.GenerateTasks(d.Block)
				if err != nil {
					return nil, err
				}
				block["block"] = inner
			}
			if len(d.Rescue) > 0 {
				rescue, err := a.GenerateTasks(d.Rescue)
				if err != nil {
					return nil, err
				}
				block["rescue"] = rescue
			}
			if len(d.Always) > 0 {
				always, err := a.GenerateTasks(d.Always)
				if err != nil {
					return nil, err
				}
				block["always"] = always
			}
			output = append(output, block)
		}
	}
	return output, nil
}

// parseTaskResponse parses the raw JSON response from an AI provider into a
// task map. It handles both {"tasks": [...]} and a bare object/array.
func parseTaskResponse(raw string) (map[string]interface{}, error) {
	cleaned := escapeJSON(raw)
	var parsed interface{}
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		return nil, fmt.Errorf("parsing provider response as JSON: %w", err)
	}

	switch v := parsed.(type) {
	case map[string]interface{}:
		if tasks, ok := v["tasks"]; ok {
			if taskList, ok := tasks.([]interface{}); ok && len(taskList) > 0 {
				if task, ok := taskList[0].(map[string]interface{}); ok {
					return task, nil
				}
			}
		}
		return v, nil
	case []interface{}:
		if len(v) > 0 {
			if task, ok := v[0].(map[string]interface{}); ok {
				return task, nil
			}
		}
		return nil, fmt.Errorf("unexpected empty array in provider response")
	default:
		return nil, fmt.Errorf("unexpected provider response type: %T", parsed)
	}
}

// escapeJSON mirrors the Python escape_json helper: replaces spaces inside
// Jinja2 double-brace expressions with underscores to avoid JSON parse errors.
func escapeJSON(text string) string {
	if !strings.Contains(text, "{{") {
		return text
	}
	// Replace spaces within {{ ... }} to avoid breaking JSON parsing.
	var out strings.Builder
	inBraces := 0
	for i := 0; i < len(text); i++ {
		c := text[i]
		if i+1 < len(text) && c == '{' && text[i+1] == '{' {
			inBraces++
			out.WriteByte(c)
			continue
		}
		if i+1 < len(text) && c == '}' && text[i+1] == '}' {
			if inBraces > 0 {
				inBraces--
			}
			out.WriteByte(c)
			continue
		}
		if inBraces > 0 && c == ' ' {
			out.WriteByte('_')
		} else {
			out.WriteByte(c)
		}
	}
	return out.String()
}

// capitalize returns text with its first letter uppercased.
func capitalize(text string) string {
	if text == "" {
		return text
	}
	return strings.ToUpper(text[:1]) + text[1:]
}

// deduplicateOrder returns a slice with duplicates removed, preserving order.
func deduplicateOrder(items []string) []string {
	seen := make(map[string]bool)
	result := items[:0]
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}
