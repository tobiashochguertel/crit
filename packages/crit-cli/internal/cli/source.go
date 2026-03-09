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

// Compile-time fallback values for repo coordinates.
// Override at runtime with CRIT_GITHUB_OWNER, CRIT_GITHUB_REPO, CRIT_GITHUB_BRANCH.
const (
	DefaultRepoOwner  = "tobiashochguertel"
	DefaultRepoName   = "crit"
	DefaultRepoBranch = "main"
)

// envOrDefault returns the value of the named environment variable, or fallback
// if the variable is unset or empty.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// defaultBase returns the GitHub raw-content base URL computed from env vars
// (CRIT_GITHUB_OWNER, CRIT_GITHUB_REPO, CRIT_GITHUB_BRANCH), falling back to
// DefaultRepoOwner, DefaultRepoName, DefaultRepoBranch.
func defaultBase() string {
	owner := envOrDefault("CRIT_GITHUB_OWNER", DefaultRepoOwner)
	repo := envOrDefault("CRIT_GITHUB_REPO", DefaultRepoName)
	branch := envOrDefault("CRIT_GITHUB_BRANCH", DefaultRepoBranch)
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", owner, repo, branch)
}

// DefaultSkillsURL is the effective base URL for skill files (SKILL.md, manifest.yaml).
// Computed once at startup from CRIT_GITHUB_OWNER / CRIT_GITHUB_REPO / CRIT_GITHUB_BRANCH
// env vars with compile-time fallbacks.  To override the URL entirely (e.g. a local
// path or a different service), set CRIT_SKILLS_DIR instead.
//
//nolint:gochecknoglobals
var DefaultSkillsURL = defaultBase() + "/packages/claude-code/skills"

// DefaultCommandsURL is the effective base URL for opencode command files.
// Computed once at startup from the same env vars as DefaultSkillsURL.
// Override entirely with CRIT_OPENCODE_DIR.
//
//nolint:gochecknoglobals
var DefaultCommandsURL = defaultBase() + "/packages/opencode/commands"

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
