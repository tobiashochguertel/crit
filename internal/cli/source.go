package cli

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultSkillsURL is the base raw-content URL for the canonical skill files
// (SKILL.md files used by Claude Code and Copilot CLI).
// Individual files are fetched by appending "/<skill-dir>/SKILL.md".
const DefaultSkillsURL = "https://raw.githubusercontent.com/tobiashochguertel/crit/feature/multi-agent-plugin-support/plugin/crit/skills"

// DefaultCommandsURL is the base raw-content URL for the canonical opencode
// command files.  Individual files are fetched by appending "/<cmd>.md".
const DefaultCommandsURL = "https://raw.githubusercontent.com/tobiashochguertel/crit/feature/multi-agent-plugin-support/plugin/crit/opencode"

// Config holds crit's persistent configuration, stored in
// ~/.config/crit/config.yaml (XDG) or ~/.crit.yaml (fallback).
type Config struct {
	// SkillsURL overrides DefaultSkillsURL for Claude Code / Copilot skill installs.
	SkillsURL string `yaml:"skills_url,omitempty"`
	// CommandsURL overrides DefaultCommandsURL for opencode command installs.
	CommandsURL string `yaml:"commands_url,omitempty"`
}

// configPaths returns candidate config file locations in priority order.
func configPaths() []string {
	var paths []string

	// 1. $XDG_CONFIG_HOME/crit/config.yaml
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		paths = append(paths, filepath.Join(xdg, "crit", "config.yaml"))
	}

	if home, err := os.UserHomeDir(); err == nil {
		// 2. ~/.config/crit/config.yaml
		paths = append(paths, filepath.Join(home, ".config", "crit", "config.yaml"))
		// 3. ~/.crit.yaml
		paths = append(paths, filepath.Join(home, ".crit.yaml"))
	}

	return paths
}

// LoadConfig reads the first config file found in configPaths().
// Returns a zero-value Config if no file exists or if parsing fails.
func LoadConfig() Config {
	for _, p := range configPaths() {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var cfg Config
		if err := yaml.Unmarshal(data, &cfg); err == nil {
			return cfg
		}
	}
	return Config{}
}

// Manifest describes the files available at a given source location.
// It is fetched from manifest.yaml at the source base URL or local path,
// allowing new skills/commands to be discovered at runtime without a CLI release.
type Manifest struct {
	Version  int           `yaml:"version"`
	Skills   []SkillSpec   `yaml:"skills,omitempty"`
	Commands []CommandSpec `yaml:"commands,omitempty"`
}

// SkillSpec identifies a single SKILL.md-based skill subdirectory.
type SkillSpec struct {
	Dir  string `yaml:"dir"`  // subdirectory name (and skill name)
	Name string `yaml:"name"` // display name
}

// CommandSpec identifies a single flat .md command file.
type CommandSpec struct {
	File string `yaml:"file"` // filename without .md extension
	Name string `yaml:"name"` // display name
}

// FetchManifest attempts to load manifest.yaml from source.
// Returns defaultManifest unchanged if the file is absent, unreachable, or malformed.
func FetchManifest(source string, defaultManifest Manifest) Manifest {
	data, err := FetchFile(source, "manifest.yaml")
	if err != nil {
		return defaultManifest
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return defaultManifest
	}
	if len(m.Skills)+len(m.Commands) == 0 {
		return defaultManifest
	}
	return m
}

// ResolveSource returns the effective source (local path or URL) for an asset
// type.  The first non-empty value in priority order wins:
//
//  1. flagValue  – passed via a --source CLI flag
//  2. envVar     – environment variable (e.g. "CRIT_SKILLS_DIR")
//  3. cfgValue   – value from the config file
//  4. defaultURL – hardcoded fallback URL
func ResolveSource(flagValue, envVar, cfgValue, defaultURL string) string {
	if flagValue != "" {
		return flagValue
	}
	if v := os.Getenv(envVar); v != "" {
		return v
	}
	if cfgValue != "" {
		return cfgValue
	}
	return defaultURL
}

// IsURL returns true if s looks like an HTTP or HTTPS URL.
func IsURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// FetchFile reads a single file from either a local path or a remote URL base.
//
// When source is a URL:
//
//	url = strings.TrimRight(source, "/") + "/" + relPath
//	An HTTP GET is performed and the response body is returned.
//
// When source is a local path:
//
//	The file is read from filepath.Join(source, filepath.FromSlash(relPath)).
func FetchFile(source, relPath string) ([]byte, error) {
	if IsURL(source) {
		url := strings.TrimRight(source, "/") + "/" + relPath
		resp, err := http.Get(url) //nolint:noctx
		if err != nil {
			return nil, fmt.Errorf("fetching %s: %w", url, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("fetching %s: HTTP %d %s", url, resp.StatusCode, resp.Status)
		}
		return io.ReadAll(resp.Body)
	}

	// Expand ~ shorthand for home directory.
	if strings.HasPrefix(source, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			source = filepath.Join(home, source[2:])
		}
	}

	return os.ReadFile(filepath.Join(source, filepath.FromSlash(relPath)))
}
