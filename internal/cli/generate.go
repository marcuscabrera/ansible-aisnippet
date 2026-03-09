package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/marcuscabrera/ansible-aisnippet/internal/config"
	"github.com/marcuscabrera/ansible-aisnippet/internal/core"
)

var (
	flagVerbose    bool
	flagFiletasks  string
	flagOutputfile string
	flagPlaybook   bool
	flagProvider   string
	flagConfig     string
)

var generateCmd = &cobra.Command{
	Use:   "generate [text]",
	Short: "Generate an Ansible task or playbook from a description",
	Long: `Generate one or more Ansible tasks from a natural-language description.

Examples:
  # Generate a single task from a sentence
  ansible-aisnippet generate "Install package htop"

  # Generate tasks from a YAML file
  ansible-aisnippet generate -f tasks.yml

  # Generate a full playbook and save to file
  ansible-aisnippet generate -f tasks.yml --playbook -o playbook.yml`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGenerate,
}

func init() {
	generateCmd.Flags().BoolVarP(&flagVerbose, "verbose", "v", false, "Enable verbose output")
	generateCmd.Flags().StringVarP(&flagFiletasks, "filetasks", "f", "", "Path to a YAML file containing tasks to generate")
	generateCmd.Flags().StringVarP(&flagOutputfile, "outputfile", "o", "", "Path to save the generated YAML output")
	generateCmd.Flags().BoolVarP(&flagPlaybook, "playbook", "p", false, "Wrap generated tasks in a playbook structure")
	generateCmd.Flags().StringVar(&flagProvider, "provider", "", "AI provider to use (overrides AI_PROVIDER env var)")
	generateCmd.Flags().StringVarP(&flagConfig, "config", "c", "", "Path to a YAML configuration file")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Build configuration.
	var cfg *config.Config
	var err error
	if flagConfig != "" {
		cfg, err = config.FromFile(flagConfig)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
	} else {
		cfg = config.FromEnv()
	}
	if flagProvider != "" {
		cfg.Provider = flagProvider
	}

	assistant, err := core.New(core.Options{
		Verbose: flagVerbose,
		Config:  cfg,
	})
	if err != nil {
		return fmt.Errorf("initializing assistant: %w", err)
	}

	if flagFiletasks != "" {
		// Generate tasks from a YAML task file.
		tasksData, err := os.ReadFile(flagFiletasks)
		if err != nil {
			return fmt.Errorf("reading tasks file: %w", err)
		}
		var taskDescs []core.TaskDescriptor
		if err := yaml.Unmarshal(tasksData, &taskDescs); err != nil {
			return fmt.Errorf("parsing tasks file: %w", err)
		}

		outputTasks, err := assistant.GenerateTasks(taskDescs)
		if err != nil {
			return fmt.Errorf("generating tasks: %w", err)
		}

		var output interface{}
		if flagPlaybook {
			output = []map[string]interface{}{
				{
					"name":          "Playbook generated with AI",
					"hosts":         "all",
					"gather_facts":  true,
					"tasks":         outputTasks,
				},
			}
		} else {
			output = outputTasks
		}

		if flagOutputfile != "" {
			if err := writeYAML(flagOutputfile, output); err != nil {
				return fmt.Errorf("writing output file: %w", err)
			}
			fmt.Printf("Output saved to %s\n", flagOutputfile)
		} else {
			yamlBytes, err := yaml.Marshal(output)
			if err != nil {
				return fmt.Errorf("marshalling output: %w", err)
			}
			fmt.Println("Result:")
			fmt.Println(string(yamlBytes))
		}
	} else {
		// Generate a single task from a text argument.
		text := "Install package htop"
		if len(args) > 0 {
			text = args[0]
		}

		task, err := assistant.GenerateTask(text)
		if err != nil {
			return fmt.Errorf("generating task: %w", err)
		}

		yamlBytes, err := yaml.Marshal(task)
		if err != nil {
			return fmt.Errorf("marshalling task: %w", err)
		}
		fmt.Println("Result:")
		fmt.Println(string(yamlBytes))
	}
	return nil
}

// writeYAML marshals data as YAML and writes it to path.
func writeYAML(path string, data interface{}) error {
	out, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}
