// Package cli provides the Cobra CLI commands for ansible-aisnippet.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "0.2.0"

var rootCmd = &cobra.Command{
	Use:   "ansible-aisnippet",
	Short: "Generate Ansible tasks from natural-language descriptions using AI",
	Long: `ansible-aisnippet converts natural-language task descriptions into
Ansible tasks and playbooks by querying AI language models.

Set AI_PROVIDER and the corresponding API key environment variable to choose
your AI backend (openai, anthropic, google, azure, mistral, cohere, ollama,
lmstudio, llama, huggingface, openrouter, zen).`,
}

// Execute is the entry point called by main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = version
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(listProvidersCmd)
}
