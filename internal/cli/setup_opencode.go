package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

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

Source priority (first non-empty wins):
  1. --source <path|url>     local directory or HTTP(S) URL base
  2. $CRIT_OPENCODE_DIR      same, via environment variable
  3. commands_url in config  ~/.config/crit/config.yaml
  4. (default URL)           ` + DefaultCommandsURL,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := LoadConfig()
		source := ResolveSource(opencodeSource, "CRIT_OPENCODE_DIR", cfg.CommandsURL, DefaultCommandsURL)

		destDir, scope, err := resolveTargetDir(
			opencodeProject,
			".opencode/commands",
			".config/opencode/commands",
		)
		if err != nil {
			return err
		}
		fmt.Printf("Installing opencode commands %s  →  %s\n", scope, destDir)
		fmt.Printf("Source: %s\n", source)
		if err := installCommands(source, destDir, opencodeCommands, opencodeForce); err != nil {
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
	setupOpencodeCmd.Flags().StringVar(&opencodeSource, "source", "", "local directory or HTTP(S) URL base for command files")
}
