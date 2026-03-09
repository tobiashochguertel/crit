package cli

// FileType controls which file format and installer is used for a given agent.
type FileType int

const (
	// FileTypeSkills installs SKILL.md files nested in named subdirectories.
	FileTypeSkills FileType = iota
	// FileTypeCommands installs flat .md command files directly into the target directory.
	FileTypeCommands
)

// AgentDef describes all installation parameters for a supported AI agent.
// Adding support for a new agent requires only a new entry in the agents map;
// no new .go files are needed.
type AgentDef struct {
	Name            string              // human-readable name, e.g. "Claude Code"
	CmdName         string              // cobra subcommand name, e.g. "setup-claude"
	DefaultURL      string              // default remote source URL base
	EnvVar          string              // environment variable for source override
	ConfigURL       func(Config) string // extracts source URL from config file
	HomeDir         string              // home-relative install path (global mode)
	ProjectDir      string              // CWD-relative install path (--project mode)
	FileType        FileType            // Skills or Commands installer
	DefaultManifest Manifest            // fallback when manifest.yaml cannot be fetched
}

// agentOrder is the canonical registration sequence; determines command listing in help.
var agentOrder = []string{"claude", "copilot", "opencode"}

// agents is the registry of all supported AI agents.
var agents = map[string]AgentDef{
	"claude": {
		Name:            "Claude Code",
		CmdName:         "setup-claude",
		DefaultURL:      DefaultSkillsURL,
		EnvVar:          "CRIT_SKILLS_DIR",
		ConfigURL:       func(c Config) string { return c.SkillsURL },
		HomeDir:         ".claude/skills",
		ProjectDir:      ".claude/skills",
		FileType:        FileTypeSkills,
		DefaultManifest: Manifest{Skills: defaultSkills},
	},
	"copilot": {
		Name:            "GitHub Copilot CLI",
		CmdName:         "setup-copilot",
		DefaultURL:      DefaultSkillsURL,
		EnvVar:          "CRIT_SKILLS_DIR",
		ConfigURL:       func(c Config) string { return c.SkillsURL },
		HomeDir:         ".copilot/skills",
		ProjectDir:      ".copilot/skills",
		FileType:        FileTypeSkills,
		DefaultManifest: Manifest{Skills: defaultSkills},
	},
	"opencode": {
		Name:            "opencode",
		CmdName:         "setup-opencode",
		DefaultURL:      DefaultCommandsURL,
		EnvVar:          "CRIT_OPENCODE_DIR",
		ConfigURL:       func(c Config) string { return c.CommandsURL },
		HomeDir:         ".config/opencode/commands",
		ProjectDir:      ".opencode/commands",
		FileType:        FileTypeCommands,
		DefaultManifest: Manifest{Commands: defaultCommands},
	},
}

// defaultSkills is the hardcoded fallback used when manifest.yaml cannot be fetched.
// Keep this list in sync with plugin/crit/skills/manifest.yaml.
var defaultSkills = []SkillSpec{
	{Dir: "crit-review", Name: "crit-review"},
	{Dir: "crit-plan-review", Name: "crit-plan-review"},
	{Dir: "crit-code-review", Name: "crit-code-review"},
}

// defaultCommands is the hardcoded fallback used when manifest.yaml cannot be fetched.
// Keep this list in sync with plugin/crit/opencode/manifest.yaml.
var defaultCommands = []CommandSpec{
	{File: "crit-review", Name: "crit-review"},
	{File: "crit-code-review", Name: "crit-code-review"},
	{File: "crit-plan-review", Name: "crit-plan-review"},
}
