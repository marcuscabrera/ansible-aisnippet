package similarity_test

import (
	"testing"

	"github.com/marcuscabrera/ansible-aisnippet/internal/similarity"
)

func TestFindMostSimilar_ExactMatch(t *testing.T) {
	docs := []string{
		"install package htop",
		"start nginx service",
		"copy file to remote host",
	}
	engine := similarity.New(docs)
	result := engine.FindMostSimilar("install package htop")
	if result != "install package htop" {
		t.Errorf("expected 'install package htop', got %q", result)
	}
}

func TestFindMostSimilar_PartialMatch(t *testing.T) {
	docs := []string{
		"install upgrade remove and list apt packages",
		"start nginx service",
		"copy file to remote host",
	}
	engine := similarity.New(docs)
	// "apt package installation" should match the first doc
	result := engine.FindMostSimilar("install apt package")
	if result != "install upgrade remove and list apt packages" {
		t.Errorf("expected apt packages doc, got %q", result)
	}
}

func TestFindMostSimilar_EmptyCorpus(t *testing.T) {
	engine := similarity.New(nil)
	result := engine.FindMostSimilar("anything")
	if result != "" {
		t.Errorf("expected empty string for empty corpus, got %q", result)
	}
}

func TestFindMostSimilar_SingleDoc(t *testing.T) {
	docs := []string{"only document"}
	engine := similarity.New(docs)
	result := engine.FindMostSimilar("any query")
	if result != "only document" {
		t.Errorf("expected 'only document', got %q", result)
	}
}

func TestFindMostSimilar_CaseInsensitive(t *testing.T) {
	docs := []string{
		"install upgrade remove and list apt packages",
		"copy file to remote",
	}
	engine := similarity.New(docs)
	// Uppercase query should still match
	result := engine.FindMostSimilar("INSTALL APT PACKAGES")
	if result != "install upgrade remove and list apt packages" {
		t.Errorf("unexpected result: %q", result)
	}
}

func TestFindMostSimilar_ReturnsConsistentResults(t *testing.T) {
	docs := []string{
		"add new hosts and groups in inventory",
		"install upgrade remove and list apt packages",
		"execute command on remote node",
	}
	engine := similarity.New(docs)

	// Same query twice must yield same result.
	r1 := engine.FindMostSimilar("apt install package")
	r2 := engine.FindMostSimilar("apt install package")
	if r1 != r2 {
		t.Errorf("inconsistent results: %q vs %q", r1, r2)
	}
}
