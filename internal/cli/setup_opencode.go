package cli

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

//go:embed opencode/crit-review.md opencode/crit-code-review.md opencode/crit-plan-review.md
var opencodeContent embed.FS

var opencodeProject bool
var opencodeForce bool

// opencode commands to install: filename (without .md) -> display name
var opencodeCommands = []struct {
	file string
	name string
}{
	{"crit-review", "crit-review"},
	{"crit-code-review", "crit-code-review"},
	{"crit-plan-review", "crit-plan-review"},
}

var setupOpencodeCmd = &cobra.Command{
	Use:   "setup-opencode",
	Short: "Install opencode commands for crit review workflow",
	Long:  "Installs crit-review, crit-code-review, and crit-plan-review commands to ~/.config/opencode/commands/ (or .opencode/commands/ with --project). These commands integrate crit with the opencode AI assistant.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var baseDir string

		if opencodeProject {
			baseDir = filepath.Join(".opencode", "commands")
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("could not determine home directory: %w", err)
			}
			baseDir = filepath.Join(home, ".config", "opencode", "commands")
		}

		scope := "globally"
		if opencodeProject {
			scope = "for this project"
		}

		for _, command := range opencodeCommands {
			targetPath := filepath.Join(baseDir, command.file+".md")

			if !opencodeForce {
				if _, err := os.Stat(targetPath); err == nil {
					fmt.Printf("Skipping %s (already exists, use --force to overwrite)\n", command.name)
					continue
				}
			}

			content, err := opencodeContent.ReadFile(filepath.Join("opencode", command.file+".md"))
			if err != nil {
				return fmt.Errorf("reading embedded command %s: %w", command.name, err)
			}

			if err := os.MkdirAll(baseDir, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", baseDir, err)
			}

			if err := os.WriteFile(targetPath, content, 0644); err != nil {
				return fmt.Errorf("writing command file %s: %w", command.name, err)
			}

			fmt.Printf("Installed /%s command %s to %s\n", command.name, scope, targetPath)
		}

		fmt.Println("\nAvailable commands (type / in opencode to invoke):")
		fmt.Println("  /crit-review         — Routes to code or plan review")
		fmt.Println("  /crit-code-review    — Multi-file code review")
		fmt.Println("  /crit-plan-review    — Single-file document review")
		fmt.Println("\nOr run from the repo root with:")
		if opencodeProject {
			fmt.Println("  /crit-code-review    (project command)")
		} else {
			fmt.Println("  /crit-code-review    (global command)")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupOpencodeCmd)
	setupOpencodeCmd.Flags().BoolVar(&opencodeProject, "project", false, "install to .opencode/commands/ in the current directory instead of globally")
	setupOpencodeCmd.Flags().BoolVar(&opencodeForce, "force", false, "overwrite existing command files")
}
