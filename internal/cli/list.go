package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/marcuscabrera/ansible-aisnippet/internal/providers"
)

var listProvidersCmd = &cobra.Command{
	Use:   "list-providers",
	Short: "List all available AI providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Available AI providers:")
		for _, name := range providers.ListProviders() {
			fmt.Printf("  • %s\n", name)
		}
		return nil
	},
}
