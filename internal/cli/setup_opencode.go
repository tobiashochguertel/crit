package cli

import (
	"embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed opencode/crit-review.md opencode/crit-code-review.md opencode/crit-plan-review.md
var opencodeContent embed.FS

var opencodeProject bool
var opencodeForce bool
var opencodeSource string

// opencodeCommands lists the commands installed by setup-opencode.
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
	Long: `Installs crit-review, crit-code-review, and crit-plan-review commands.

Default target:  ~/.config/opencode/commands/
With --project:  .opencode/commands/ (current directory)

Source priority:
  1. --source <dir>     local directory containing *.md command files
  2. $CRIT_OPENCODE_DIR same, via environment variable
  3. (embedded)         files bundled inside the binary (default)`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		destDir, scope, err := resolveTargetDir(
			opencodeProject,
			".opencode/commands",
			".config/opencode/commands",
		)
		if err != nil {
			return err
		}
		src, prefix, err := openSourceFS(opencodeSource, "CRIT_OPENCODE_DIR", opencodeContent, "opencode")
		if err != nil {
			return err
		}
		fmt.Printf("Installing opencode commands %s  →  %s\n", scope, destDir)
		if err := installCommands(src, prefix, destDir, opencodeCommands, opencodeForce); err != nil {
			return err
		}
		fmt.Println("\nAvailable commands (type / in opencode to invoke):")
		fmt.Println("  /crit-review         — routes to code or plan review")
		fmt.Println("  /crit-code-review    — multi-file code review")
		fmt.Println("  /crit-plan-review    — single-file document review")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupOpencodeCmd)
	setupOpencodeCmd.Flags().BoolVar(&opencodeProject, "project", false, "install to .opencode/commands/ in current directory instead of globally")
	setupOpencodeCmd.Flags().BoolVar(&opencodeForce, "force", false, "overwrite existing command files")
	setupOpencodeCmd.Flags().StringVar(&opencodeSource, "source", "", "local directory with *.md command files (overrides embedded; also honours $CRIT_OPENCODE_DIR)")
}
