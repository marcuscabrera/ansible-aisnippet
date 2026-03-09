package providers_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/marcuscabrera/ansible-aisnippet/internal/config"
	"github.com/marcuscabrera/ansible-aisnippet/internal/providers"
)

// --- Factory tests ---

func TestListProviders_ContainsAllTwelve(t *testing.T) {
	expected := []string{
		"openai", "anthropic", "google", "azure", "mistral",
		"cohere", "ollama", "lmstudio", "llama", "huggingface",
		"openrouter", "zen",
	}
	all := providers.ListProviders()
	set := make(map[string]bool, len(all))
	for _, name := range all {
		set[name] = true
	}
	for _, want := range expected {
		if !set[want] {
			t.Errorf("provider %q missing from ListProviders()", want)
		}
	}
}

func TestCreate_KnownProviders(t *testing.T) {
	names := []string{
		"openai", "anthropic", "google", "azure", "mistral",
		"cohere", "ollama", "lmstudio", "llama", "huggingface",
		"openrouter", "zen",
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			cfg := config.DefaultProviderConfig()
			p, err := providers.Create(name, cfg)
			if err != nil {
				t.Fatalf("Create(%q): unexpected error: %v", name, err)
			}
			if p.Name() != name {
				t.Errorf("expected Name()=%q, got %q", name, p.Name())
			}
		})
	}
}

func TestCreate_CaseInsensitive(t *testing.T) {
	cfg := config.DefaultProviderConfig()
	p, err := providers.Create("OpenAI", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "openai" {
		t.Errorf("expected openai, got %q", p.Name())
	}
}

func TestCreate_UnknownProviderReturnsError(t *testing.T) {
	cfg := config.DefaultProviderConfig()
	_, err := providers.Create("nonexistent", cfg)
	if err == nil {
		t.Error("expected error for unknown provider, got nil")
	}
}

// --- FallbackManager tests ---

// fakeProvider is a test-double that either returns a response or an error.
type fakeProvider struct {
	name     string
	response string
	err      error
}

func (f *fakeProvider) Name() string { return f.name }
func (f *fakeProvider) ValidateConfig() bool { return true }
func (f *fakeProvider) Generate(_, _ string) (string, error) {
	return f.response, f.err
}

func TestFallbackManager_EmptyProviders(t *testing.T) {
	_, err := providers.NewFallbackManager([]providers.Provider{})
	if err == nil {
		t.Error("expected error for empty providers slice")
	}
}

func TestFallbackManager_ReturnsFirstProviderResponse(t *testing.T) {
	p1 := &fakeProvider{name: "p1", response: "hello from p1"}
	p2 := &fakeProvider{name: "p2", response: "hello from p2"}
	mgr, err := providers.NewFallbackManager([]providers.Provider{p1, p2})
	if err != nil {
		t.Fatal(err)
	}
	result, used, err := mgr.Generate("sys", "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello from p1" {
		t.Errorf("expected 'hello from p1', got %q", result)
	}
	if used != "p1" {
		t.Errorf("expected used=p1, got %q", used)
	}
}

func TestFallbackManager_FallsBackOnError(t *testing.T) {
	p1 := &fakeProvider{name: "p1", err: errors.New("API error")}
	p2 := &fakeProvider{name: "p2", response: "hello from p2"}
	mgr, err := providers.NewFallbackManager([]providers.Provider{p1, p2})
	if err != nil {
		t.Fatal(err)
	}
	result, used, err := mgr.Generate("sys", "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello from p2" {
		t.Errorf("expected 'hello from p2', got %q", result)
	}
	if used != "p2" {
		t.Errorf("expected used=p2, got %q", used)
	}
}

func TestFallbackManager_AllFail(t *testing.T) {
	p1 := &fakeProvider{name: "p1", err: errors.New("err1")}
	p2 := &fakeProvider{name: "p2", err: errors.New("err2")}
	mgr, err := providers.NewFallbackManager([]providers.Provider{p1, p2})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = mgr.Generate("sys", "user")
	if err == nil {
		t.Error("expected error when all providers fail")
	}
}

func TestFallbackManager_ErrorIncludesAllProviderNames(t *testing.T) {
	p1 := &fakeProvider{name: "p1", err: errors.New("boom1")}
	p2 := &fakeProvider{name: "p2", err: errors.New("boom2")}
	mgr, _ := providers.NewFallbackManager([]providers.Provider{p1, p2})
	_, _, err := mgr.Generate("sys", "user")
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "p1") || !strings.Contains(msg, "p2") {
		t.Errorf("error message should mention p1 and p2, got: %s", msg)
	}
}

func TestFallbackManager_SingleProviderSuccess(t *testing.T) {
	p1 := &fakeProvider{name: "p1", response: "solo response"}
	mgr, _ := providers.NewFallbackManager([]providers.Provider{p1})
	result, used, err := mgr.Generate("sys", "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "solo response" {
		t.Errorf("expected 'solo response', got %q", result)
	}
	if used != "p1" {
		t.Errorf("expected used=p1, got %q", used)
	}
}
