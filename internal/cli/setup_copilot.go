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
embedded source. The same $CRIT_SKILLS_DIR environment variable applies.

Source priority:
  1. --source <dir>     local directory containing skill subdirectories
  2. $CRIT_SKILLS_DIR   same, via environment variable
  3. (embedded)         files bundled inside the binary (default)`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		destDir, scope, err := resolveTargetDir(copilotProject, ".copilot/skills", ".copilot/skills")
		if err != nil {
			return err
		}
		src, prefix, err := openSourceFS(copilotSource, "CRIT_SKILLS_DIR", skillContent, "skill")
		if err != nil {
			return err
		}
		fmt.Printf("Installing GitHub Copilot CLI skills %s  →  %s\n", scope, destDir)
		if err := installSkills(src, prefix, destDir, skillsToInstall, copilotForce); err != nil {
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
	setupCopilotCmd.Flags().StringVar(&copilotSource, "source", "", "local directory with skill subdirs (overrides embedded; also honours $CRIT_SKILLS_DIR)")
}
