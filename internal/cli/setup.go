package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path"
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

// openSourceFS resolves which filesystem to read install files from.
//
// Priority (first non-empty wins):
//  1. flagPath  – passed via a --source CLI flag
//  2. envVar    – environment variable name (e.g. "CRIT_SKILLS_DIR")
//  3. embedded  – the embedded fs.FS (fallback, bundled in the binary)
//
// When a flag or env path is used the returned prefix is "." because files
// are expected directly under that directory root.
// When the embedded fallback is used, embeddedPrefix is returned instead
// (e.g. "skill" or "opencode" to match the embed.FS directory layout).
func openSourceFS(flagPath, envVar string, embedded fs.FS, embeddedPrefix string) (fs.FS, string, error) {
	p := flagPath
	if p == "" {
		p = os.Getenv(envVar)
	}
	if p != "" {
		info, err := os.Stat(p)
		if err != nil {
			return nil, "", fmt.Errorf("source path %q: %w", p, err)
		}
		if !info.IsDir() {
			return nil, "", fmt.Errorf("source path %q is not a directory", p)
		}
		return os.DirFS(p), ".", nil
	}
	return embedded, embeddedPrefix, nil
}

// installSkills reads SKILL.md files from src and writes them under destDir.
//
// Each entry in skills has a dir (subdirectory name) and a display name.
// The source path is  path.Join(prefix, skill.dir, "SKILL.md").
// The destination is  filepath.Join(destDir, skill.dir, "SKILL.md").
func installSkills(src fs.FS, prefix, destDir string, skills []struct{ dir, name string }, force bool) error {
	for _, skill := range skills {
		targetDir := filepath.Join(destDir, skill.dir)
		targetPath := filepath.Join(targetDir, "SKILL.md")

		if !force {
			if _, err := os.Stat(targetPath); err == nil {
				fmt.Printf("  skipping %-22s (already exists; use --force to overwrite)\n", skill.name)
				continue
			}
		}

		srcPath := path.Join(prefix, skill.dir, "SKILL.md")
		data, err := fs.ReadFile(src, srcPath)
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

// installCommands reads flat *.md command files from src and writes them into destDir.
//
// Each entry in commands has a file (base name without .md) and a display name.
// The source path is  path.Join(prefix, cmd.file+".md").
// The destination is  filepath.Join(destDir, cmd.file+".md").
func installCommands(src fs.FS, prefix, destDir string, commands []struct{ file, name string }, force bool) error {
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

		srcPath := path.Join(prefix, cmd.file+".md")
		data, err := fs.ReadFile(src, srcPath)
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
