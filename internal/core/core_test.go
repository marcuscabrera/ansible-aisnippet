package core_test

import (
	"testing"

	"github.com/marcuscabrera/ansible-aisnippet/internal/config"
	"github.com/marcuscabrera/ansible-aisnippet/internal/core"
)

func testConfig() *config.Config {
	// Minimal config with cache/rate-limit disabled so tests don't block.
	cfg := config.Default()
	cfg.Cache.Enabled = false
	cfg.RateLimit.Enabled = false
	return cfg
}

func TestNew_SucceedsWithDefaultConfig(t *testing.T) {
	cfg := testConfig()
	a, err := core.New(core.Options{Config: cfg})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == nil {
		t.Error("expected non-nil AISnippet")
	}
}

func TestNew_OverridesProvider(t *testing.T) {
	cfg := testConfig()
	_, err := core.New(core.Options{Config: cfg, Provider: "ollama"})
	if err != nil {
		t.Fatalf("unexpected error when overriding provider to ollama: %v", err)
	}
}

func TestNew_InvalidProviderReturnsError(t *testing.T) {
	cfg := testConfig()
	// Set up a config that has fallback providers with an invalid name.
	cfg.FallbackProviders = []string{"nonexistent_provider_xyz"}
	_, err := core.New(core.Options{Config: cfg})
	if err == nil {
		t.Error("expected error for invalid fallback provider, got nil")
	}
}

func TestGenerateTasks_EmptyList(t *testing.T) {
	cfg := testConfig()
	a, err := core.New(core.Options{Config: cfg})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tasks, err := a.GenerateTasks(nil)
	if err != nil {
		t.Fatalf("unexpected error for empty task list: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected empty output, got %v", tasks)
	}
}

func TestGenerateTasks_WithBlockDescriptor(t *testing.T) {
	cfg := testConfig()
	a, err := core.New(core.Options{Config: cfg})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// A block descriptor with no "task" field should produce an entry with "block" key.
	// The nested block is empty so no provider calls are made.
	descriptors := []core.TaskDescriptor{
		{
			Name: "My block",
			When: "ansible_os_family == 'Debian'",
		},
	}
	tasks, err := a.GenerateTasks(descriptors)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0]["name"] != "My block" {
		t.Errorf("expected name='My block', got %v", tasks[0]["name"])
	}
	if tasks[0]["when"] != "ansible_os_family == 'Debian'" {
		t.Errorf("expected when condition, got %v", tasks[0]["when"])
	}
}
