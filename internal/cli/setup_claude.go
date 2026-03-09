package cli

import (
	"embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed skill/crit-review/SKILL.md skill/crit-plan-review/SKILL.md skill/crit-code-review/SKILL.md
var skillContent embed.FS

// skillsToInstall is shared with setup_copilot.go (same package, same SKILL.md format).
var skillsToInstall = []struct {
	dir  string
	name string
}{
	{"crit-review", "crit-review"},
	{"crit-plan-review", "crit-plan-review"},
	{"crit-code-review", "crit-code-review"},
}

var claudeProject bool
var claudeForce bool
var claudeSource string

var setupClaudeCmd = &cobra.Command{
	Use:   "setup-claude",
	Short: "Install Claude Code skills for crit review workflow",
	Long: `Installs /crit-review, /crit-plan-review, and /crit-code-review skills.

Default target:  ~/.claude/skills/
With --project:  .claude/skills/ (current directory)

Source priority:
  1. --source <dir>     local directory containing skill subdirectories
  2. $CRIT_SKILLS_DIR   same, via environment variable
  3. (embedded)         files bundled inside the binary (default)`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		destDir, scope, err := resolveTargetDir(claudeProject, ".claude/skills", ".claude/skills")
		if err != nil {
			return err
		}
		src, prefix, err := openSourceFS(claudeSource, "CRIT_SKILLS_DIR", skillContent, "skill")
		if err != nil {
			return err
		}
		fmt.Printf("Installing Claude Code skills %s  →  %s\n", scope, destDir)
		if err := installSkills(src, prefix, destDir, skillsToInstall, claudeForce); err != nil {
			return err
		}
		fmt.Println("\nAvailable skills:")
		fmt.Println("  /crit-review         — routes to code or plan review")
		fmt.Println("  /crit-code-review    — multi-file code review")
		fmt.Println("  /crit-plan-review    — single-file document review")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupClaudeCmd)
	setupClaudeCmd.Flags().BoolVar(&claudeProject, "project", false, "install to .claude/skills/ in current directory instead of globally")
	setupClaudeCmd.Flags().BoolVar(&claudeForce, "force", false, "overwrite existing skill files")
	setupClaudeCmd.Flags().StringVar(&claudeSource, "source", "", "local directory with skill subdirs (overrides embedded; also honours $CRIT_SKILLS_DIR)")
}
