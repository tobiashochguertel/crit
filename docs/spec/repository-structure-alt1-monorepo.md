---
title: "Alternative Repository Structure — Monorepo"
description: "A proposed refactoring of the crit repository to a clean monorepo where each AI agent integration lives in its own package, eliminating duplication between embedded binary files and plugin distribution files."
last_updated: "2025-03-09"
---

# Alternative Repository Structure — Monorepo

This document proposes refactoring the crit repository into a **monorepo** where every
AI agent integration lives in a dedicated `packages/<agent>/` directory.  The root
remains the Claude Code Plugin **Marketplace host** (Layer A).  The CLI source code
moves to `packages/crit-cli/`.  All documentation and demo assets move into `docs/`.

See [`repository-structure.md`](./repository-structure.md) for the current structure and
the explanation of Layers A–E.

---

## Goals

| Goal | How the monorepo achieves it |
|------|------------------------------|
| Eliminate Layer C/D duplication | One canonical skills/commands source per agent; binary embeds from generated copies only |
| Clear ownership per AI agent | `packages/claude-code/`, `packages/copilot/`, `packages/opencode/` |
| Marketplace uses `git-subdir` | `marketplace.json` references each package as a subdirectory — no full-repo clone needed |
| Flat root | Only marketplace files and top-level metadata live at the root |
| Docs and demos co-located | `docs/` contains all documentation, specs, demos, and assets |

---

## Proposed Tree

```tree
.
├── .claude-plugin/                       # [Layer A] Marketplace host — unchanged
│   ├── marketplace.json                  #   Points to packages/claude-code/ via git-subdir source
│   └── plugin.json                       #   Fallback: direct-install plugin manifest
│
├── .github/
│   └── workflows/
│       └── release.yml                   # GoReleaser CI; only triggers for packages/crit-cli/
│
├── .gitignore                             # Root-level ignores (dist/, .task/, .claude/)
│
├── packages/
│   │
│   ├── crit-cli/                          # [CLI] Go source for the `crit` binary
│   │   ├── .goreleaser.yaml               #   Cross-platform release config
│   │   ├── .mise.toml                     #   Pins Go version (e.g. 1.24.2)
│   │   ├── CHANGELOG.md                   #   Version history (moved from root)
│   │   ├── Taskfile.yml                   #   build, test, lint, format, tidy, clean, sync, all
│   │   ├── go.mod                         #   module github.com/kevindutra/crit (path unchanged)
│   │   ├── go.sum
│   │   │
│   │   ├── .opencode/                     #   [Layer E] opencode cmds for crit DEVELOPMENT
│   │   │   └── commands/
│   │   │       ├── crit-review.md         #     Symlink → ../../opencode/commands/crit-review.md
│   │   │       ├── crit-code-review.md    #     (or Taskfile `sync` step copies them)
│   │   │       └── crit-plan-review.md
│   │   │
│   │   ├── cmd/
│   │   │   └── crit/
│   │   │       └── main.go
│   │   │
│   │   └── internal/
│   │       ├── cli/
│   │       │   ├── embed/                 #   BUILD-TIME GENERATED — not committed to git
│   │       │   │   ├── claude-code/       #     Copied from packages/claude-code/skills/
│   │       │   │   │   ├── crit-review/SKILL.md
│   │       │   │   │   ├── crit-code-review/SKILL.md
│   │       │   │   │   └── crit-plan-review/SKILL.md
│   │       │   │   └── opencode/          #     Copied from packages/opencode/commands/
│   │       │   │       ├── crit-review.md
│   │       │   │       ├── crit-code-review.md
│   │       │   │       └── crit-plan-review.md
│   │       │   │
│   │       │   ├── comment.go
│   │       │   ├── review.go
│   │       │   ├── review_test.go
│   │       │   ├── root.go
│   │       │   ├── setup_claude.go        #   Embeds from ./embed/claude-code/
│   │       │   ├── setup_copilot.go       #   Embeds from ./embed/claude-code/ (same skills)
│   │       │   ├── setup_opencode.go      #   Embeds from ./embed/opencode/
│   │       │   └── status.go
│   │       ├── document/
│   │       ├── git/
│   │       ├── review/
│   │       └── tui/
│   │
│   ├── claude-code/                       # [Layer B] Claude Code plugin package
│   │   ├── .claude-plugin/
│   │   │   └── plugin.json                #   Plugin manifest (includes "skills": "skills/")
│   │   ├── commands/                      #   Slash commands used after /plugin install
│   │   │   ├── review.md
│   │   │   ├── code-review.md
│   │   │   └── plan-review.md
│   │   └── skills/                        #   ★ SINGLE SOURCE OF TRUTH for skills
│   │       ├── crit-review/
│   │       │   └── SKILL.md               #   Used by both Claude Code and Copilot CLI
│   │       ├── crit-code-review/
│   │       │   └── SKILL.md
│   │       └── crit-plan-review/
│   │           └── SKILL.md
│   │
│   ├── copilot/                           # [Layer B'] Copilot CLI plugin package
│   │   └── skills/                        #   ★ Could be symlinks → ../claude-code/skills/
│   │       ├── crit-review/               #     (skills are format-compatible)
│   │       │   └── SKILL.md
│   │       ├── crit-code-review/
│   │       │   └── SKILL.md
│   │       └── crit-plan-review/
│   │           └── SKILL.md
│   │
│   └── opencode/                          # [Layer D] opencode commands package
│       └── commands/                      #   ★ SINGLE SOURCE OF TRUTH for opencode commands
│           ├── crit-review.md
│           ├── crit-code-review.md
│           └── crit-plan-review.md
│
├── docs/
│   ├── ai-agent-plugins/
│   │   ├── README.md
│   │   ├── claude-code.md
│   │   ├── copilot-cli.md
│   │   └── opencode.md
│   ├── spec/
│   │   ├── repository-structure.md        #   Current structure
│   │   ├── repository-structure-alt1-monorepo.md   # THIS FILE
│   │   └── repository-structure-alt2-multi-repo.md
│   ├── assets/                            #   Moved from root assets/
│   │   └── crit_logo.png
│   └── demo/                             #   Moved from root demo/
│       ├── demo.gif
│       ├── demo.tape
│       ├── code-review.gif
│       ├── code-review.tape
│       └── plan.md
│
├── README.md                              # Project README (update asset/demo paths)
└── Taskfile.yml                           # Root orchestrator: delegates to packages/crit-cli/Taskfile.yml
```

---

## Key Changes Explained

### 1. Eliminating Layer C / D Duplication

**The problem today:** The same SKILL.md content appears in two places:

| Location | Role |
|----------|------|
| `internal/cli/skill/*/SKILL.md` | Embedded in the Go binary for `setup-claude` / `setup-copilot` |
| `plugin/crit/skills/*/SKILL.md` | Distributed as part of the marketplace plugin |

If a skill description changes, both copies must be updated manually — and they can drift apart.

**The monorepo fix:** `packages/claude-code/skills/` becomes the **single source of truth**.
The Go binary no longer has its own copy committed to git.  Instead:

1. A Taskfile `sync` step (run automatically before `build`) copies files from
   `packages/claude-code/skills/` and `packages/opencode/commands/` into
   `packages/crit-cli/internal/cli/embed/` (a gitignored generated directory).
2. `setup_claude.go` and `setup_copilot.go` embed from `./embed/claude-code/`.
3. `setup_opencode.go` embeds from `./embed/opencode/`.

```yaml
# In packages/crit-cli/Taskfile.yml
sync:
  desc: "Sync plugin files into embed/ for binary embedding"
  cmds:
    - rm -rf internal/cli/embed
    - mkdir -p internal/cli/embed/claude-code internal/cli/embed/opencode
    - cp -r ../../packages/claude-code/skills/. internal/cli/embed/claude-code/
    - cp ../../packages/opencode/commands/*.md internal/cli/embed/opencode/

build:
  deps: [sync]
  cmds:
    - go build ./cmd/crit
```

`.gitignore` in `packages/crit-cli/`:
```
internal/cli/embed/
```

### 2. Marketplace `marketplace.json` using `git-subdir`

With the new structure, `marketplace.json` can reference the plugin using the
`git-subdir` source type.  Claude Code performs a **sparse clone** of only the
`packages/claude-code/` subtree — minimizing bandwidth for users:

```json
{
  "name": "crit-marketplace",
  "owner": { "name": "Tobias Hochgürtel" },
  "metadata": {
    "pluginRoot": "./packages"
  },
  "plugins": [
    {
      "name": "crit",
      "source": {
        "source": "git-subdir",
        "url": "https://github.com/tobiashochguertel/crit.git",
        "path": "packages/claude-code"
      },
      "description": "Review documents and code with an interactive TUI, then let Claude address feedback automatically.",
      "version": "1.0.2",
      "commands": ["./commands/"],
      "strict": false
    }
  ]
}
```

Alternatively, use a relative path (works when the marketplace and plugin are in the
same repo):

```json
{
  "plugins": [
    {
      "name": "crit",
      "source": "./packages/claude-code",
      "description": "..."
    }
  ]
}
```

### 3. `packages/copilot/` — Shared or symlinked skills

The SKILL.md format for Claude Code and Copilot CLI is **identical** (both use
`name`, `description`, `allowed-tools`, `argument-hint` frontmatter).  Options:

| Option | Pros | Cons |
|--------|------|------|
| Symlinks `packages/copilot/skills/ → ../claude-code/skills/` | True single source; no sync needed | Symlinks require git LFS or careful `.gitattributes` handling |
| Copy via Taskfile `sync` | Explicit; no symlink complexity | Two copies; must re-run sync on change |
| Drop `packages/copilot/` entirely | Simplest | Copilot CLI plugin marketplace discovery requires a dedicated path |

**Recommendation:** Symlink `packages/copilot/skills/` to `../claude-code/skills/` and
add a note in the README that skills are format-compatible between both agents.

### 4. Root `Taskfile.yml` delegates to CLI package

```yaml
# Root Taskfile.yml — orchestrator
version: '3'

tasks:
  build:
    desc: "Build the crit binary"
    dir: packages/crit-cli
    cmds: [task build]

  test:
    dir: packages/crit-cli
    cmds: [task test]

  lint:
    dir: packages/crit-cli
    cmds: [task lint]

  all:
    dir: packages/crit-cli
    cmds: [task all]
```

### 5. `docs/` as the home for demos and assets

`README.md` image and GIF paths update from:
```markdown
![demo](demo/code-review.gif)
![logo](assets/crit_logo.png)
```
to:
```markdown
![demo](docs/demo/code-review.gif)
![logo](docs/assets/crit_logo.png)
```

---

## Migration Steps

1. **Create directory scaffold** — `packages/crit-cli/`, `packages/claude-code/`,
   `packages/copilot/`, `packages/opencode/`, `docs/assets/`, `docs/demo/`
2. **Move Go source** — `cmd/`, `internal/`, `go.mod`, `go.sum`, `.goreleaser.yaml`,
   `.mise.toml`, `CHANGELOG.md`, `Taskfile.yml` → `packages/crit-cli/`
3. **Move plugin files** — `plugin/crit/` → `packages/claude-code/`; add `packages/copilot/`
4. **Move opencode files** — `internal/cli/opencode/*.md` → `packages/opencode/commands/`
5. **Move demos/assets** — `demo/` → `docs/demo/`, `assets/` → `docs/assets/`
6. **Add Taskfile `sync` step** — generates `internal/cli/embed/` before build
7. **Update `.gitignore`** — add `packages/crit-cli/internal/cli/embed/`
8. **Update `go:embed` paths** — point to `./embed/claude-code/` and `./embed/opencode/`
9. **Update `marketplace.json`** — use `git-subdir` or relative path to `packages/claude-code/`
10. **Update `README.md`** — fix asset/demo paths; update install instructions

---

## Trade-offs

| Benefit | Cost |
|---------|------|
| Single source of truth for skills/commands | Requires `task sync` before build |
| Clear per-agent ownership | More directories at top level |
| `git-subdir` for efficient plugin fetch | Slightly more complex `marketplace.json` |
| All docs/demos in one place | Update image paths in README |
| Root stays clean | Root `Taskfile.yml` needed as orchestrator |

---

## Comparison with Current Structure

| Concern | Current | Monorepo alternative |
|---------|---------|---------------------|
| Skills source of truth | Two copies (embed + plugin) | One copy in `packages/claude-code/skills/` |
| AI agent directories | Mixed under `plugin/` and `internal/cli/` | Dedicated `packages/<agent>/` |
| CLI source location | Repository root | `packages/crit-cli/` |
| Demo/asset location | `demo/` and `assets/` at root | `docs/demo/` and `docs/assets/` |
| Marketplace source type | Relative path `./plugin/crit` | `git-subdir` or relative `./packages/claude-code` |
