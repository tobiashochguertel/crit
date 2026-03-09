package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var copilotProject bool
var copilotForce bool
var copilotSource string

var setupCopilotCmd = &cobra.Command{
	Use:   "setup-copilot",
	Short: "Install GitHub Copilot CLI skills for crit review workflow",
	Long: `Installs crit-review, crit-plan-review, and crit-code-review skills.

Default target:  ~/.copilot/skills/
With --project:  .copilot/skills/ (current directory)

The skill files are format-compatible with Claude Code and share the same
default source URL.  The same $CRIT_SKILLS_DIR environment variable applies.

Source priority (first non-empty wins):
  1. --source <path|url>  local directory or HTTP(S) URL base
  2. $CRIT_SKILLS_DIR     same, via environment variable
  3. skills_url in config  ~/.config/crit/config.yaml
  4. (default URL)         ` + DefaultSkillsURL,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := LoadConfig()
		source := ResolveSource(copilotSource, "CRIT_SKILLS_DIR", cfg.SkillsURL, DefaultSkillsURL)

		destDir, scope, err := resolveTargetDir(copilotProject, ".copilot/skills", ".copilot/skills")
		if err != nil {
			return err
		}
		fmt.Printf("Installing GitHub Copilot CLI skills %s  →  %s\n", scope, destDir)
		fmt.Printf("Source: %s\n", source)
		if err := installSkills(source, destDir, skillsToInstall, copilotForce); err != nil {
			return err
		}
		fmt.Println("\nAvailable skills (reference by name in your Copilot CLI session):")
		fmt.Println("  crit-review         — routes to code or plan review")
		fmt.Println("  crit-code-review    — multi-file code review")
		fmt.Println("  crit-plan-review    — single-file document review")
		fmt.Println("\nExample:")
		fmt.Println("  ghcs 'use the crit-code-review skill to review my changes'")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCopilotCmd)
	setupCopilotCmd.Flags().BoolVar(&copilotProject, "project", false, "install to .copilot/skills/ in current directory instead of globally")
	setupCopilotCmd.Flags().BoolVar(&copilotForce, "force", false, "overwrite existing skill files")
	setupCopilotCmd.Flags().StringVar(&copilotSource, "source", "", "local directory or HTTP(S) URL base for skill files")
}
