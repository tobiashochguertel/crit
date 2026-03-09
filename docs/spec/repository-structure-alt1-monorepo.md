---
title: "Alternative Repository Structure — Monorepo"
description: "A proposed refactoring of the crit repository to a clean monorepo where each AI agent integration lives in its own package, eliminating duplication between embedded binary files and plugin distribution files."
last_updated: "2025-03-09"
---

# Alternative Repository Structure — Monorepo

This document proposes refactoring the crit repository into a **monorepo** where every
AI agent integration lives in a dedicated `packages/<agent>/` directory.  The root
remains the Claude Code Plugin **Marketplace host** (`[Layer A]`).  The CLI source code
moves to `packages/crit-cli/`.  All documentation and demo assets move into `docs/`.

See [`repository-structure.md`](./repository-structure.md) for the current structure and
the explanation of the layer label scheme.

---

## Goals

| Goal | How the monorepo achieves it |
|------|------------------------------|
| Eliminate `[Layer C]` duplication | One canonical skills/commands source per agent; binary embeds from generated copies only |
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
│   └── plugin.json                       #   [Layer A:claude-code] Fallback: direct-install manifest
│
├── .github/
│   └── workflows/
│       └── release.yml                   # GoReleaser CI; only triggers for packages/crit-cli/
│
├── .gitignore                             # Root-level ignores (dist/, .task/, .claude/, .opencode/)
│
├── packages/
│   │
│   ├── crit-cli/                          # [CLI] Go source for the `crit` binary
│   │   ├── .goreleaser.yaml               #   Cross-platform release config
│   │   ├── .mise.toml                     #   Pins Go version (e.g. 1.24.2)
│   │   ├── CHANGELOG.md                   #   Version history (moved from root)
│   │   ├── Taskfile.yml                   #   build, test, lint, format, tidy, clean
│   │   │                                  #   sync, all, init-claude, init-copilot, init-opencode
│   │   ├── go.mod                         #   module github.com/kevindutra/crit (path unchanged)
│   │   ├── go.sum
│   │   │
│   │   ├── cmd/
│   │   │   └── crit/
│   │   │       └── main.go
│   │   │
│   │   └── internal/
│   │       ├── cli/
│   │       │   ├── embed/                 #   BUILD-TIME GENERATED — git-ignored
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
│   │       │   ├── setup.go               #   Shared installer helpers
│   │       │   ├── setup_claude.go        #   [Layer C:claude-code] Embeds from ./embed/claude-code/
│   │       │   ├── setup_copilot.go       #   [Layer C:copilot]     Embeds from ./embed/claude-code/
│   │       │   ├── setup_opencode.go      #   [Layer C:opencode]    Embeds from ./embed/opencode/
│   │       │   └── status.go
│   │       ├── document/
│   │       ├── git/
│   │       ├── review/
│   │       └── tui/
│   │
│   ├── claude-code/                       # [Layer B:claude-code] Claude Code plugin package
│   │   ├── .claude-plugin/
│   │   │   └── plugin.json                #   Plugin manifest (includes "skills": "skills/")
│   │   ├── commands/                      #   [Layer A:claude-code] Slash commands after /plugin install
│   │   │   ├── review.md
│   │   │   ├── code-review.md
│   │   │   └── plan-review.md
│   │   └── skills/                        #   ★ SINGLE SOURCE OF TRUTH for SKILL.md files
│   │       ├── crit-review/
│   │       │   └── SKILL.md               #   Used by both Claude Code and Copilot CLI
│   │       ├── crit-code-review/
│   │       │   └── SKILL.md
│   │       └── crit-plan-review/
│   │           └── SKILL.md
│   │
│   ├── copilot/                           # [Layer B:copilot] Copilot CLI plugin package
│   │   └── skills -> ../claude-code/skills #   ★ Symlink — skills are format-compatible
│   │
│   └── opencode/                          # [Layer C:opencode] opencode commands package
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
│   │   ├── repository-structure.md
│   │   ├── repository-structure-alt1-monorepo.md    # THIS FILE
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

> **Note — `.opencode/` is git-ignored at the repo root and inside `packages/crit-cli/`.**
> Run `task init-opencode` (from the repo root or from `packages/crit-cli/`) to create
> `.opencode/commands/` locally — labelled **[Layer D:opencode]** — for contributors who
> use opencode.  The same applies for `init-claude` and `init-copilot`.

---

## Key Changes Explained

### 1. Eliminating `[Layer C]` Duplication

**The problem today:** The same SKILL.md content appears in two places:

| Location | Label | Role |
|----------|-------|------|
| `internal/cli/skill/*/SKILL.md` | `[Layer C:claude-code]` / `[Layer C:copilot]` | Embedded in the Go binary |
| `plugin/crit/skills/*/SKILL.md` | `[Layer B:copilot]` | Distributed as part of the marketplace plugin |

If a skill description changes, both copies must be updated manually — they can drift.

**The monorepo fix:** `packages/claude-code/skills/` becomes the **single source of truth**.
The Go binary no longer has its own copy committed.  A Taskfile `sync` step copies files
into `packages/crit-cli/internal/cli/embed/` (git-ignored) before every build.

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

### 2. `packages/copilot/` — Symlinked skills

The SKILL.md format for Claude Code and Copilot CLI is **identical**.
`packages/copilot/skills` is a symlink to `../claude-code/skills` — no duplication,
no sync step required for Copilot CLI.

```bash
# Create the symlink (once, during repository setup)
cd packages/copilot && ln -s ../claude-code/skills skills
```

Add to root `.gitattributes` so git tracks the symlink:
```
packages/copilot/skills export-subst
```

### 3. Marketplace `marketplace.json` using `git-subdir`

```json
{
  "plugins": [
    {
      "name": "crit",
      "source": "./packages/claude-code",
      "description": "Review documents and code with an interactive TUI, then let Claude address feedback automatically."
    }
  ]
}
```

### 4. `init-*` Taskfile tasks (project-local AI agent setup)

```yaml
# In packages/crit-cli/Taskfile.yml
init-claude:
  desc: "Install Claude Code skills for this project"
  deps: [build]
  cmds: [./dist/crit setup-claude --project --force]

init-copilot:
  desc: "Install GitHub Copilot CLI skills for this project"
  deps: [build]
  cmds: [./dist/crit setup-copilot --project --force]

init-opencode:
  desc: "Install opencode commands for this project"
  deps: [build]
  cmds: [./dist/crit setup-opencode --project --force]
```

### 5. Root `Taskfile.yml` delegates to CLI package

```yaml
version: "3"

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

---

## Migration Steps

1. **Create directory scaffold** — `packages/crit-cli/`, `packages/claude-code/`,
   `packages/copilot/`, `packages/opencode/`, `docs/assets/`, `docs/demo/`
2. **Move Go source** — `cmd/`, `internal/`, `go.mod`, `go.sum`, `.goreleaser.yaml`,
   `.mise.toml`, `CHANGELOG.md`, `Taskfile.yml` → `packages/crit-cli/`
3. **Move plugin files** — `plugin/crit/` → `packages/claude-code/`
4. **Create symlink** — `packages/copilot/skills -> ../claude-code/skills`
5. **Move opencode sources** — `internal/cli/opencode/*.md` → `packages/opencode/commands/`
6. **Move demos/assets** — `demo/` → `docs/demo/`, `assets/` → `docs/assets/`
7. **Add Taskfile `sync` step** — generates `internal/cli/embed/` before build
8. **Update `.gitignore`** — add `packages/crit-cli/internal/cli/embed/`, `.opencode/`
9. **Update `go:embed` paths** — point to `./embed/claude-code/` and `./embed/opencode/`
10. **Update `marketplace.json`** — use relative path `./packages/claude-code`
11. **Update `README.md`** — fix asset/demo paths; update install instructions

---

## Trade-offs

| Benefit | Cost |
|---------|------|
| Single source of truth for skills/commands | Requires `task sync` before build |
| Symlink eliminates Copilot duplicate | Symlinks need care in `.gitattributes` |
| Clear per-agent ownership | More directories at top level |
| `init-*` tasks for local AI agent config | Developers must run `task init-opencode` etc. |
| All docs/demos in one place | Update image paths in README |

---

## Comparison with Current Structure

| Concern | Current | Monorepo alternative |
|---------|---------|---------------------|
| Skills source of truth | Two copies (`[Layer C]` + `[Layer B:copilot]`) | One copy in `packages/claude-code/skills/` |
| AI agent directories | Mixed under `plugin/` and `internal/cli/` | Dedicated `packages/<agent>/` |
| CLI source location | Repository root | `packages/crit-cli/` |
| Demo/asset location | `demo/` and `assets/` at root | `docs/demo/` and `docs/assets/` |
| Marketplace source | Relative path `./plugin/crit` | Relative `./packages/claude-code` |
| `[Layer D:opencode]` | `.opencode/` committed | `.opencode/` git-ignored; created by `task init-opencode` |
