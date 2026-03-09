---
title: "Repository Structure"
description: "Annotated map of every directory and file in the crit repository, explaining why apparent duplicates exist and which layer each piece belongs to."
last_updated: "2025-03-09"
---

# Repository Structure

The crit repository looks confusing at first because similar-looking directories appear
in multiple places (e.g. `commands/` vs `plugin/crit/commands/`).  They are **not**
duplicates — each set belongs to a different **installation layer**.  This document
explains every directory and file.

## Annotated Tree

```tree
.
├── .claude-plugin/                    # [Layer A] Makes this repo a Claude Code MARKETPLACE
│   ├── marketplace.json               #   Declares this repo as a marketplace; lists plugin sources
│   └── plugin.json                    #   Fallback: lets the repo be installed as a plugin directly
│
├── .github/
│   └── workflows/
│       └── release.yml                # GoReleaser GitHub Actions workflow for tagged releases
│
├── .gitignore                         # Ignores built binaries (/crit, dist/), .task/, .claude/
│
├── .goreleaser.yaml                   # GoReleaser config: cross-platform build + GitHub release
│
├── .mise.toml                         # Pins Go version (1.24.2) for mise version manager
│
├── .opencode/                         # [Layer E] opencode commands for developing crit ITSELF
│   └── commands/
│       ├── crit-review.md             #   Route between code-review and plan-review modes
│       ├── crit-code-review.md        #   Run a code review with crit (for crit contributors)
│       └── crit-plan-review.md        #   Run a plan/doc review with crit (for crit contributors)
│
├── assets/
│   └── crit_logo.png                  # Logo used in README
│
├── cmd/
│   └── crit/
│       └── main.go                    # Binary entry point; calls cli.Execute()
│
├── commands/                          # [Layer A] Claude Code slash commands for the ROOT plugin
│   ├── review.md                      #   /review  — routes to code or plan review
│   ├── code-review.md                 #   /code-review — multi-file code review
│   └── plan-review.md                 #   /plan-review  — document/plan review
│
├── CHANGELOG.md                       # Version history
│
├── demo/
│   ├── demo.gif                       # Animated demo (plan review)
│   ├── demo.tape                      # VHS tape source for demo.gif
│   ├── code-review.gif                # Animated demo (code review)
│   ├── code-review.tape               # VHS tape source for code-review.gif
│   └── plan.md                        # Sample plan used in demo recording
│
├── docs/
│   ├── ai-agent-plugins/              # Documentation for multi-agent plugin support
│   │   ├── README.md                  #   Overview of supported AI agents
│   │   ├── copilot-cli.md             #   GitHub Copilot CLI plugin setup guide
│   │   └── opencode.md                #   opencode command setup guide
│   └── spec/
│       └── repository-structure.md    # THIS FILE — annotated repo map
│
├── go.mod                             # Go module definition (module path: github.com/kevindutra/crit)
├── go.sum                             # Dependency checksums (auto-managed by go toolchain)
│
├── internal/                          # All Go application code (not exported)
│   ├── cli/
│   │   ├── opencode/                  # [Layer D] opencode command files EMBEDDED in the binary
│   │   │   ├── crit-review.md         #   Installed to ~/.config/opencode/commands/ by setup-opencode
│   │   │   ├── crit-code-review.md    #   (global) or .opencode/commands/ (--project)
│   │   │   └── crit-plan-review.md
│   │   │
│   │   ├── skill/                     # [Layer C] SKILL.md files EMBEDDED in the binary
│   │   │   ├── crit-review/
│   │   │   │   └── SKILL.md           #   Installed to ~/.claude/skills/ by setup-claude
│   │   │   ├── crit-code-review/
│   │   │   │   └── SKILL.md           #   (global) or .claude/skills/ (--project)
│   │   │   └── crit-plan-review/
│   │   │       └── SKILL.md           #   Also used by setup-copilot for GitHub Copilot CLI
│   │   │
│   │   ├── comment.go                 # Comment data model and serialization
│   │   ├── review.go                  # `crit review` command implementation
│   │   ├── review_test.go
│   │   ├── root.go                    # Root cobra command + persistent flags
│   │   ├── setup_claude.go            # `crit setup-claude` — installs skills for Claude Code
│   │   ├── setup_copilot.go           # `crit setup-copilot` — installs skills for GitHub Copilot CLI
│   │   ├── setup_opencode.go          # `crit setup-opencode` — installs commands for opencode
│   │   └── status.go                  # `crit status` command — prints pending review comments
│   │
│   ├── document/
│   │   ├── document.go                # Markdown document parser and line index
│   │   ├── document_test.go
│   │   └── paths.go                   # Path resolution helpers
│   │
│   ├── git/
│   │   └── diff.go                    # Git diff parsing (changed files, hunks)
│   │
│   ├── review/
│   │   ├── session.go                 # Review session lifecycle
│   │   ├── store.go                   # Comment persistence (JSON file store)
│   │   ├── store_test.go
│   │   └── types.go                   # Shared types: Comment, ReviewSession, etc.
│   │
│   └── tui/
│       ├── app.go                     # Bubbletea root model (main TUI app)
│       ├── app_test.go
│       ├── filetab.go                 # File tab component (multi-file review)
│       ├── highlight.go               # Syntax highlighting via chroma
│       ├── keys.go                    # Keybinding definitions
│       ├── messages.go                # Bubbletea messages (events between components)
│       └── styles.go                  # Lipgloss style definitions
│
├── plugin/
│   └── crit/                          # [Layer B] The INSTALLABLE plugin package
│       ├── .claude-plugin/
│       │   └── plugin.json            #   Plugin manifest for the marketplace-installed plugin
│       ├── commands/                  #   Slash commands installed to ~/.claude/plugins/crit/commands/
│       │   ├── review.md
│       │   ├── code-review.md
│       │   └── plan-review.md
│       └── skills/                    # [Layer B+] Copilot CLI skills inside the plugin
│           ├── crit-review/
│           │   └── SKILL.md           #   Discoverable by GitHub Copilot CLI via plugin.json "skills"
│           ├── crit-code-review/
│           │   └── SKILL.md
│           └── crit-plan-review/
│               └── SKILL.md
│
├── README.md                          # Project documentation and install instructions
└── Taskfile.yml                       # Dev task runner: build, test, lint, format, tidy, clean, all
```

## The Five Installation Layers

Understanding the structure requires knowing **who installs what, where, and for whom**.

### Layer A — Root plugin (repo-as-plugin / marketplace host)

| Directory | Purpose |
|-----------|---------|
| `.claude-plugin/marketplace.json` | Declares this repo a **Claude Code marketplace**. When a user runs `/plugin marketplace add tobiashochguertel/crit`, Claude Code reads this file to discover available plugins. It points to `./plugin/crit` as the plugin source. |
| `.claude-plugin/plugin.json` | Allows the repo to also be installed **directly** as a plugin (without going through a marketplace). The slash commands in `commands/` are used in this mode. |
| `commands/` | Slash commands (`/review`, `/code-review`, `/plan-review`) available when the repo is used as a direct plugin. Content mirrors `plugin/crit/commands/` but serves this alternate install path. |

### Layer B — Marketplace-installed plugin (`plugin/crit/`)

When a user installs via `/plugin install crit`, Claude Code copies `plugin/crit/` to
`~/.claude/plugins/crit/`.  Everything inside this directory is what the user receives:

| Path | Purpose |
|------|---------|
| `plugin/crit/.claude-plugin/plugin.json` | Plugin manifest consumed by Claude Code after install. Includes `"skills": "skills/"` to expose Copilot CLI skills. |
| `plugin/crit/commands/` | Slash commands available after plugin install (`/crit:review`, etc.). |
| `plugin/crit/skills/` | GitHub Copilot CLI skills. Discovered by Copilot CLI via the `"skills"` field in `plugin.json`. |

### Layer C — Standalone Claude / Copilot skill install (`internal/cli/skill/`)

`crit setup-claude` and `crit setup-copilot` read these files from the embedded
filesystem inside the binary and write them to:

- `~/.claude/skills/` (global) or `.claude/skills/` (`--project`) for **Claude Code**
- `~/.copilot/skills/` (global) or `.copilot/skills/` (`--project`) for **GitHub Copilot CLI**

This layer is for users who do **not** use the plugin marketplace but want the skills
available as standalone `/crit-review`, `/crit-code-review`, `/crit-plan-review` commands.

### Layer D — opencode command install (`internal/cli/opencode/`)

`crit setup-opencode` reads these flat `.md` files (embedded in the binary) and copies
them to:

- `~/.config/opencode/commands/` (global) or `.opencode/commands/` (`--project`)

opencode commands are flat files with YAML frontmatter (`description`, `agent`, `model`).
They do not support `allowed-tools` or `argument-hint`.

### Layer E — Developer environment (`.opencode/commands/`)

When a **contributor to crit** runs opencode inside this repository, these commands
are available.  Mirrors the role of root `commands/` but for opencode instead of
Claude Code.

## Why `commands/` and `plugin/crit/commands/` Have the Same Content

They are intentionally identical in content but exist for different reasons:

| File set | Used when |
|----------|-----------|
| `commands/*.md` | Repo is added as a plugin directly (Layer A path) |
| `plugin/crit/commands/*.md` | Plugin is installed via marketplace (Layer B path) |

The Claude Code plugin system resolves slash commands from wherever the plugin was
installed.  Keeping both sets ensures the commands work regardless of how the user
installed crit.

## Why `internal/cli/skill/` and `plugin/crit/skills/` Both Have SKILL.md Files

| File set | Format | Used by |
|----------|--------|---------|
| `internal/cli/skill/*/SKILL.md` | Embedded → installed by `crit setup-claude` / `setup-copilot` | Users who run the CLI setup commands |
| `plugin/crit/skills/*/SKILL.md` | Checked into `plugin/` tree, copied at plugin install | GitHub Copilot CLI users who install via the plugin marketplace |

The content is equivalent, but the delivery mechanism is different.

## Files That Are **Not** Tracked (`.gitignore`)

| Path | Reason |
|------|--------|
| `/crit` | Root-level built binary from `go build` |
| `dist/` | GoReleaser output directory with cross-platform binaries |
| `.task/` | Taskfile internal checksum cache (build freshness tracking) |
| `.claude/` | Local Claude Code session state (not part of the plugin) |
