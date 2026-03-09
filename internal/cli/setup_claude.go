package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

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

Source priority (first non-empty wins):
  1. --source <path|url>  local directory or HTTP(S) URL base
  2. $CRIT_SKILLS_DIR     same, via environment variable
  3. skills_url in config  ~/.config/crit/config.yaml
  4. (default URL)         ` + DefaultSkillsURL,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := LoadConfig()
		source := ResolveSource(claudeSource, "CRIT_SKILLS_DIR", cfg.SkillsURL, DefaultSkillsURL)

		destDir, scope, err := resolveTargetDir(claudeProject, ".claude/skills", ".claude/skills")
		if err != nil {
			return err
		}
		fmt.Printf("Installing Claude Code skills %s  →  %s\n", scope, destDir)
		fmt.Printf("Source: %s\n", source)
		if err := installSkills(source, destDir, skillsToInstall, claudeForce); err != nil {
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
	setupClaudeCmd.Flags().StringVar(&claudeSource, "source", "", "local directory or HTTP(S) URL base for skill files")
}
