package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var copilotProject bool
var copilotForce bool

var setupCopilotCmd = &cobra.Command{
	Use:   "setup-copilot",
	Short: "Install GitHub Copilot CLI skills for crit review workflow",
	Long:  "Installs crit-review, crit-plan-review, and crit-code-review skills to ~/.copilot/skills/ (or .copilot/skills/ with --project). These skills integrate crit with the GitHub Copilot CLI plugin system.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var baseDir string

		if copilotProject {
			baseDir = filepath.Join(".copilot", "skills")
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("could not determine home directory: %w", err)
			}
			baseDir = filepath.Join(home, ".copilot", "skills")
		}

		scope := "globally"
		if copilotProject {
			scope = "for this project"
		}

		for _, skill := range skillsToInstall {
			targetDir := filepath.Join(baseDir, skill.dir)
			targetPath := filepath.Join(targetDir, "SKILL.md")

			if !copilotForce {
				if _, err := os.Stat(targetPath); err == nil {
					fmt.Printf("Skipping %s (already exists, use --force to overwrite)\n", skill.name)
					continue
				}
			}

			content, err := skillContent.ReadFile(filepath.Join("skill", skill.dir, "SKILL.md"))
			if err != nil {
				return fmt.Errorf("reading embedded skill %s: %w", skill.name, err)
			}

			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", targetDir, err)
			}

			if err := os.WriteFile(targetPath, content, 0644); err != nil {
				return fmt.Errorf("writing skill file %s: %w", skill.name, err)
			}

			fmt.Printf("Installed %s skill %s to %s\n", skill.name, scope, targetPath)
		}

		fmt.Println("\nAvailable skills:")
		fmt.Println("  crit-review         — Routes to code or plan review")
		fmt.Println("  crit-code-review    — Multi-file code review")
		fmt.Println("  crit-plan-review    — Single-file document review")
		fmt.Println("\nTo use: mention a skill by name in your Copilot CLI conversation, e.g.:")
		fmt.Println("  ghcs 'use the crit-code-review skill to review my changes'")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCopilotCmd)
	setupCopilotCmd.Flags().BoolVar(&copilotProject, "project", false, "install to .copilot/skills/ in the current directory instead of globally")
	setupCopilotCmd.Flags().BoolVar(&copilotForce, "force", false, "overwrite existing skill files")
}
