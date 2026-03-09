package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// resolveTargetDir returns the installation directory and a human-readable scope string.
//
// When project is true, localRelPath is returned as-is (relative to CWD).
// Otherwise homeRelPath is appended to the user's home directory.
func resolveTargetDir(project bool, localRelPath, homeRelPath string) (dir, scope string, err error) {
	if project {
		return localRelPath, "for this project", nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, homeRelPath), "globally", nil
}

// installSkills fetches SKILL.md files from source and writes them under destDir.
//
// source may be a local directory path or a remote HTTP/HTTPS URL base.
// Each SkillSpec carries a Dir (subdirectory name) and a display Name.
//
// Fetches:   FetchFile(source, "<spec.Dir>/SKILL.md")
// Writes to: filepath.Join(destDir, spec.Dir, "SKILL.md")
func installSkills(source, destDir string, skills []SkillSpec, force bool) error {
	for _, skill := range skills {
		targetDir := filepath.Join(destDir, skill.Dir)
		targetPath := filepath.Join(targetDir, "SKILL.md")

		if !force {
			if _, err := os.Stat(targetPath); err == nil {
				fmt.Printf("  skipping %-22s (already exists; use --force to overwrite)\n", skill.Name)
				continue
			}
		}

		data, err := FetchFile(source, skill.Dir+"/SKILL.md")
		if err != nil {
			return fmt.Errorf("reading skill %s: %w", skill.Name, err)
		}

		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("creating directory for skill %s: %w", skill.Name, err)
		}

		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			return fmt.Errorf("writing skill %s: %w", skill.Name, err)
		}

		fmt.Printf("  ✓ %-24s →  %s\n", skill.Name, targetPath)
	}
	return nil
}

// installCommands fetches flat *.md command files from source and writes them into destDir.
//
// source may be a local directory path or a remote HTTP/HTTPS URL base.
// Each CommandSpec carries a File (base name without .md) and a display Name.
//
// Fetches:   FetchFile(source, "<spec.File>.md")
// Writes to: filepath.Join(destDir, spec.File+".md")
func installCommands(source, destDir string, commands []CommandSpec, force bool) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("creating commands directory %s: %w", destDir, err)
	}

	for _, cmd := range commands {
		targetPath := filepath.Join(destDir, cmd.File+".md")

		if !force {
			if _, err := os.Stat(targetPath); err == nil {
				fmt.Printf("  skipping %-22s (already exists; use --force to overwrite)\n", cmd.Name)
				continue
			}
		}

		data, err := FetchFile(source, cmd.File+".md")
		if err != nil {
			return fmt.Errorf("reading command %s: %w", cmd.Name, err)
		}

		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			return fmt.Errorf("writing command %s: %w", cmd.Name, err)
		}

		fmt.Printf("  ✓ %-24s →  %s\n", "/"+cmd.Name, targetPath)
	}
	return nil
}

// makeSetupCmd builds a cobra.Command from an AgentDef.
//
// At runtime the command:
//  1. Resolves the source (flag → env var → config → default URL).
//  2. Fetches manifest.yaml from the source to discover available files.
//  3. Falls back to def.DefaultManifest when the manifest is absent or malformed.
//  4. Installs the files using the appropriate installer (skills or commands).
func makeSetupCmd(def AgentDef) *cobra.Command {
	var project, force bool
	var source string

	configKey := "skills_url"
	if def.FileType == FileTypeCommands {
		configKey = "commands_url"
	}

	long := fmt.Sprintf(`Installs files for the %s integration with crit.

Default target:  ~/%s/
With --project:  %s/ (current directory)

The file list is fetched from manifest.yaml at the source URL at runtime.
If the manifest cannot be retrieved, a built-in default file list is used.

Source priority (first non-empty wins):
  1. --source <path|url>     local directory or HTTP(S) URL base
  2. $%s          same, via environment variable
  3. %s in config  ~/.config/crit/config.yaml
  4. (default URL)           %s`,
		def.Name, def.HomeDir, def.ProjectDir,
		def.EnvVar, configKey, def.DefaultURL,
	)

	cmd := &cobra.Command{
		Use:   def.CmdName,
		Short: fmt.Sprintf("Install %s files for the crit review workflow", def.Name),
		Long:  long,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := LoadConfig()
			src := ResolveSource(source, def.EnvVar, def.ConfigURL(cfg), def.DefaultURL)

			destDir, scope, err := resolveTargetDir(project, def.ProjectDir, def.HomeDir)
			if err != nil {
				return err
			}

			fmt.Printf("Installing %s files %s\n", def.Name, scope)
			fmt.Printf("Source:  %s\n\n", src)

			manifest := FetchManifest(src, def.DefaultManifest)

			switch def.FileType {
			case FileTypeSkills:
				if err := installSkills(src, destDir, manifest.Skills, force); err != nil {
					return err
				}
				fmt.Println("\nAvailable skills (use in Claude Code / Copilot CLI):")
				for _, s := range manifest.Skills {
					fmt.Printf("  /%s\n", s.Name)
				}
			case FileTypeCommands:
				if err := installCommands(src, destDir, manifest.Commands, force); err != nil {
					return err
				}
				fmt.Println("\nAvailable commands (type / in opencode to invoke):")
				for _, c := range manifest.Commands {
					fmt.Printf("  /%s\n", c.Name)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&project, "project", false,
		"install into the current project directory instead of the home directory")
	cmd.Flags().BoolVar(&force, "force", false,
		"overwrite existing files")
	cmd.Flags().StringVar(&source, "source", "",
		"local directory path or HTTP(S) URL base to fetch files from")

	return cmd
}
