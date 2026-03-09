package cli

import (
	"fmt"
	"os"
	"path/filepath"
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
// Each entry in skills has a dir (subdirectory name) and a display name.
//
// Fetches:   FetchFile(source, "<skill.dir>/SKILL.md")
// Writes to: filepath.Join(destDir, skill.dir, "SKILL.md")
func installSkills(source, destDir string, skills []struct{ dir, name string }, force bool) error {
	for _, skill := range skills {
		targetDir := filepath.Join(destDir, skill.dir)
		targetPath := filepath.Join(targetDir, "SKILL.md")

		if !force {
			if _, err := os.Stat(targetPath); err == nil {
				fmt.Printf("  skipping %-22s (already exists; use --force to overwrite)\n", skill.name)
				continue
			}
		}

		data, err := FetchFile(source, skill.dir+"/SKILL.md")
		if err != nil {
			return fmt.Errorf("reading skill %s: %w", skill.name, err)
		}

		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("creating directory for skill %s: %w", skill.name, err)
		}

		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			return fmt.Errorf("writing skill %s: %w", skill.name, err)
		}

		fmt.Printf("  ✓ %-24s →  %s\n", skill.name, targetPath)
	}
	return nil
}

// installCommands fetches flat *.md command files from source and writes them into destDir.
//
// source may be a local directory path or a remote HTTP/HTTPS URL base.
// Each entry in commands has a file (base name without .md) and a display name.
//
// Fetches:   FetchFile(source, "<cmd.file>.md")
// Writes to: filepath.Join(destDir, cmd.file+".md")
func installCommands(source, destDir string, commands []struct{ file, name string }, force bool) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("creating commands directory %s: %w", destDir, err)
	}

	for _, cmd := range commands {
		targetPath := filepath.Join(destDir, cmd.file+".md")

		if !force {
			if _, err := os.Stat(targetPath); err == nil {
				fmt.Printf("  skipping %-22s (already exists; use --force to overwrite)\n", cmd.name)
				continue
			}
		}

		data, err := FetchFile(source, cmd.file+".md")
		if err != nil {
			return fmt.Errorf("reading command %s: %w", cmd.name, err)
		}

		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			return fmt.Errorf("writing command %s: %w", cmd.name, err)
		}

		fmt.Printf("  ✓ %-24s →  %s\n", "/"+cmd.name, targetPath)
	}
	return nil
}
