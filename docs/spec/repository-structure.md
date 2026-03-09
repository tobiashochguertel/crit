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
│   └── plugin.json                    #   [Layer A:claude-code] Fallback: direct plugin install
│
├── .github/
│   └── workflows/
│       └── release.yml                # GoReleaser GitHub Actions workflow for tagged releases
│
├── .gitignore                         # Ignores /crit, dist/, .task/, .claude/, .opencode/
│
├── .goreleaser.yaml                   # GoReleaser config: cross-platform build + GitHub release
│
├── .mise.toml                         # Pins Go version (1.24.2) for mise version manager
│
├── assets/
│   └── crit_logo.png                  # Logo used in README
│
├── cmd/
│   └── crit/
│       └── main.go                    # Binary entry point; calls cli.Execute()
│
├── commands/                          # [Layer A:claude-code] Slash commands for the ROOT plugin
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
│   │   ├── claude-code.md             #   Claude Code plugin setup guide
│   │   ├── copilot-cli.md             #   GitHub Copilot CLI plugin setup guide
│   │   └── opencode.md                #   opencode command setup guide
│   └── spec/
│       ├── repository-structure.md              # THIS FILE — annotated repo map
│       ├── repository-structure-alt1-monorepo.md    # Alternative: monorepo layout
│       └── repository-structure-alt2-multi-repo.md  # Alternative: split repositories
│
├── go.mod                             # Go module definition (module path: github.com/tobiashochguertel/crit)
├── go.sum                             # Dependency checksums (auto-managed by go toolchain)
│
├── internal/                          # All Go application code (not exported)
│   ├── cli/
│   │   ├── opencode/                  # [Layer C:opencode] opencode command files EMBEDDED in binary
│   │   │   ├── crit-review.md         #   Installed to ~/.config/opencode/commands/ by setup-opencode
│   │   │   ├── crit-code-review.md    #   (global) or .opencode/commands/ (--project)
│   │   │   └── crit-plan-review.md
│   │   │
│   │   ├── skill/                     # [Layer C:claude-code] [Layer C:copilot] SKILL.md files EMBEDDED in binary
│   │   │   ├── crit-review/
│   │   │   │   └── SKILL.md           #   Installed to ~/.claude/skills/ by setup-claude
│   │   │   ├── crit-code-review/      #   (global) or .claude/skills/ (--project)
│   │   │   │   └── SKILL.md           #   Also used by setup-copilot → ~/.copilot/skills/
│   │   │   └── crit-plan-review/
│   │   │       └── SKILL.md
│   │   │
│   │   ├── comment.go                 # Comment data model and serialization
│   │   ├── review.go                  # `crit review` command implementation
│   │   ├── review_test.go
│   │   ├── root.go                    # Root cobra command + persistent flags
│   │   ├── setup.go                   # Shared installer helpers (resolveTargetDir, installSkills, …)
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
│   └── crit/                          # [Layer B:claude-code] The INSTALLABLE plugin package
│       ├── .claude-plugin/
│       │   └── plugin.json            #   Plugin manifest consumed by Claude Code after install
│       ├── commands/                  #   Slash commands installed to ~/.claude/plugins/crit/commands/
│       │   ├── review.md
│       │   ├── code-review.md
│       │   └── plan-review.md
│       └── skills/                    # [Layer B:copilot] Copilot CLI skills inside the plugin
│           ├── crit-review/
│           │   └── SKILL.md           #   Discoverable by GitHub Copilot CLI via plugin.json "skills"
│           ├── crit-code-review/
│           │   └── SKILL.md
│           └── crit-plan-review/
│               └── SKILL.md
│
├── README.md                          # Project documentation and install instructions
└── Taskfile.yml                       # Dev tasks: build, test, lint, format, tidy, clean, all
                                       #            init-claude, init-copilot, init-opencode
```

> **Note — `.opencode/` is git-ignored.**  Run `task init-opencode` to create
> `.opencode/commands/` locally.  These project-local opencode commands
> are labelled **[Layer D:opencode]** in the design documents; they are not
> committed because each developer may use a different AI agent tool.

## The Installation Layers

Understanding the structure requires knowing **who installs what, where, and for whom**.
The layer labels follow the pattern `[Layer X:agent]` — the same letter means the same
_kind_ of artifact; the agent suffix says _which tool_ it targets.

### Layer A — Marketplace host and direct plugin entry point

| Path | Label | Purpose |
|------|-------|---------|
| `.claude-plugin/marketplace.json` | `[Layer A]` | Declares this repo a **Claude Code marketplace**. Running `/plugin marketplace add tobiashochguertel/crit` makes Claude Code read this file to find plugins. |
| `.claude-plugin/plugin.json` | `[Layer A:claude-code]` | Allows the repo to be installed **directly** as a plugin (without a marketplace). |
| `commands/` | `[Layer A:claude-code]` | Slash commands available when the repo is used as a direct plugin (`/review`, `/code-review`, `/plan-review`). |

### Layer B — Marketplace-installed package

| Path | Label | Purpose |
|------|-------|---------|
| `plugin/crit/` | `[Layer B:claude-code]` | The installable plugin directory. Claude Code copies this to `~/.claude/plugins/crit/` on `/plugin install crit`. |
| `plugin/crit/commands/` | `[Layer B:claude-code]` | Slash commands available after plugin install. |
| `plugin/crit/skills/` | `[Layer B:copilot]` | GitHub Copilot CLI skills. Discovered via the `"skills"` field in `plugin/crit/.claude-plugin/plugin.json`. |

### Layer C — Embedded binary → CLI-installed assets

`crit setup-*` commands read these files from the **embedded filesystem** inside the binary
and copy them to the user's machine.  The same skill files serve both Claude Code and Copilot CLI.

| Path | Label | Installed to (global / --project) |
|------|-------|------------------------------------|
| `internal/cli/skill/*/SKILL.md` | `[Layer C:claude-code]` | `~/.claude/skills/` / `.claude/skills/` |
| `internal/cli/skill/*/SKILL.md` | `[Layer C:copilot]` | `~/.copilot/skills/` / `.copilot/skills/` |
| `internal/cli/opencode/*.md` | `[Layer C:opencode]` | `~/.config/opencode/commands/` / `.opencode/commands/` |

Source resolution for all three commands (priority order):
1. `--source <dir>` flag
2. `$CRIT_SKILLS_DIR` / `$CRIT_OPENCODE_DIR` environment variable
3. Embedded files bundled in the binary (default)

### Layer D — Project-local developer tools (git-ignored)

| Path | Label | Created by |
|------|-------|-----------|
| `.opencode/commands/` | `[Layer D:opencode]` | `task init-opencode` (runs `crit setup-opencode --project --force`) |

These files are **not committed**.  They are for contributors who use opencode
to work on crit itself.  Run `task init-claude`, `task init-copilot`, or
`task init-opencode` to populate the corresponding local config.

## Why `commands/` and `plugin/crit/commands/` Have the Same Content

They are intentionally identical in content but exist for different reasons:

| File set | Label | Used when |
|----------|-------|-----------|
| `commands/*.md` | `[Layer A:claude-code]` | Repo added as a direct plugin |
| `plugin/crit/commands/*.md` | `[Layer B:claude-code]` | Plugin installed via marketplace |

## Why `internal/cli/skill/` and `plugin/crit/skills/` Both Have SKILL.md Files

| File set | Label | Delivery mechanism |
|----------|-------|--------------------|
| `internal/cli/skill/*/SKILL.md` | `[Layer C:claude-code]` / `[Layer C:copilot]` | Embedded in binary; installed by `crit setup-claude` / `setup-copilot` |
| `plugin/crit/skills/*/SKILL.md` | `[Layer B:copilot]` | Committed in `plugin/` tree; copied at plugin install |

## Files That Are **Not** Tracked (`.gitignore`)

| Path | Reason |
|------|--------|
| `/crit` | Root-level built binary from `go build` |
| `dist/` | GoReleaser output with cross-platform binaries |
| `.task/` | Taskfile internal checksum cache |
| `.claude/` | Local Claude Code session state |
| `.opencode/` | `[Layer D:opencode]` project-local opencode commands; created by `task init-opencode` |
