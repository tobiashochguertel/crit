---
title: "Alternative Repository Structure — Monorepo"
description: "A proposed refactoring of the crit repository to a clean monorepo where each AI agent integration lives in its own package, with skills and commands downloaded at runtime rather than embedded in the binary."
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
| Eliminate `[Layer C]` duplication | One canonical skills/commands source per agent; CLI downloads at runtime — no binary embedding |
| Clear ownership per AI agent | `packages/claude-code/`, `packages/copilot/`, `packages/opencode/` |
| Marketplace uses `git-subdir` | `marketplace.json` references each package as a subdirectory — no full-repo clone needed |
| Flat root | Only marketplace files and top-level metadata live at the root |
| Docs and demos co-located | `docs/` contains all documentation, specs, demos, and assets |
| No release required for skill/command changes | CLI fetches latest content from configurable source at runtime (when `crit setup-*` runs) |

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
│   │   ├── Taskfile.yml                   #   build, test, lint, format, tidy, clean, all
│   │   │                                  #   init-claude, init-copilot, init-opencode  (optional)
│   │   ├── go.mod                         #   module github.com/kevindutra/crit (path unchanged)
│   │   ├── go.sum
│   │   │
│   │   ├── cmd/
│   │   │   └── crit/
│   │   │       └── main.go
│   │   │
│   │   └── internal/
│   │       ├── cli/
│   │       │   ├── comment.go
│   │       │   ├── review.go
│   │       │   ├── review_test.go
│   │       │   ├── root.go
│   │       │   ├── setup.go               #   Shared installer helpers; makeSetupCmd factory
│   │       │   ├── agent_config.go        #   [Layer C] AgentDef registry; add new agents here
│   │       │   ├── setup_agents.go        #   [Layer C] init() registers all setup-* commands
│   │       │   ├── source.go              #   Config, LoadConfig, ResolveSource, FetchFile, FetchManifest
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
│   │   └── skills/                        #   ★ CANONICAL source for SKILL.md files
│   │       ├── manifest.yaml              #     Runtime-discoverable file list (fetched by `crit setup-claude/copilot`)
│   │       ├── crit-review/               #     Downloaded by `crit setup-claude` and `crit setup-copilot`
│   │       │   └── SKILL.md
│   │       ├── crit-code-review/
│   │       │   └── SKILL.md
│   │       └── crit-plan-review/
│   │           └── SKILL.md
│   │
│   ├── copilot/                           # [Layer B:copilot] Copilot CLI plugin package
│   │   └── skills -> ../claude-code/skills #   ★ Symlink — skills are format-compatible
│   │
│   └── opencode/                          # [Layer B:opencode] opencode commands package
│       └── commands/                      #   ★ CANONICAL source for opencode commands
│           ├── manifest.yaml              #     Runtime-discoverable file list (fetched by `crit setup-opencode`)
│           ├── crit-review.md             #     Downloaded by `crit setup-opencode`
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
> These `init-*` tasks are **optional** convenience shortcuts; the repository
> works without them.  Use them only when you want local project-level
> configuration for the AI agent you develop with.

---

## Key Changes Explained

### 1. Runtime download replaces binary embedding

**The problem with embedding:**

| Issue | Impact |
|-------|--------|
| Skill/command content changes require a new CLI release | Tight coupling between content and binary |
| Same content exists in `internal/cli/skill/` AND `plugin/crit/skills/` | Manual sync required; files can drift |
| Binary grows with every skill added | Unnecessary bloat |

**The monorepo solution:** `packages/claude-code/skills/` and
`packages/opencode/commands/` are the **canonical sources**.  When a user runs
`crit setup-claude`, the CLI downloads the files at that moment from a configurable
source URL — no embedding required.

**Source resolution order (first non-empty wins):**

```
--source <path|url>        CLI flag (local dir or HTTP(S) URL base)
$CRIT_SKILLS_DIR           Environment variable (for skills)
$CRIT_OPENCODE_DIR         Environment variable (for opencode commands)
skills_url: in config      ~/.config/crit/config.yaml
commands_url: in config    ~/.config/crit/config.yaml
(default URL)              https://raw.githubusercontent.com/tobiashochguertel/crit/main/packages/claude-code/skills
```

**Config file** (`~/.config/crit/config.yaml`):

```yaml
# Override default download sources
skills_url: https://raw.githubusercontent.com/tobiashochguertel/crit/main/packages/claude-code/skills
commands_url: https://raw.githubusercontent.com/tobiashochguertel/crit/main/packages/opencode/commands

# Or point to a local checkout for offline/development use:
# skills_url: /home/user/crit/packages/claude-code/skills
# commands_url: /home/user/crit/packages/opencode/commands
```

**`source.go` provides the download primitive:**

```go
// FetchFile reads a file from a local path or a remote URL base.
// source can be:
//   - "https://raw.githubusercontent.com/..." (HTTP GET)
//   - "/path/to/local/dir" (os.ReadFile)
//   - "~/relative/path" (expanded to $HOME/relative/path)
func FetchFile(source, relPath string) ([]byte, error)
```

### 2. `packages/copilot/` — Symlinked skills

The SKILL.md format for Claude Code and Copilot CLI is **identical**.
`packages/copilot/skills` is a symlink to `../claude-code/skills` — no duplication,
no sync step required for Copilot CLI.

```bash
# Create the symlink (once, during repository setup)
cd packages/copilot && ln -s ../claude-code/skills skills
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

> Because skills/commands are downloaded at runtime (when `crit setup-*` runs),
> `init-*` tasks do **not** need a `sync` step before building — there is nothing to
> sync into the binary.  Running an `init-*` task simply invokes the freshly-built
> binary to install files into the current project tree.

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
5. **Move opencode sources** — `plugin/crit/opencode/*.md` → `packages/opencode/commands/`
6. **Move demos/assets** — `demo/` → `docs/demo/`, `assets/` → `docs/assets/`
7. **Remove `//go:embed` and `embed.FS`** from all `setup_*.go` files (already done)
8. **Update default URLs** in `source.go` to point to `packages/claude-code/skills`
   and `packages/opencode/commands`
9. **Update `.gitignore`** — ensure `dist/`, `.opencode/`, `.claude/` are ignored
10. **Update `marketplace.json`** — use relative path `./packages/claude-code`
11. **Update `README.md`** — fix asset/demo paths; update install instructions

---

## Trade-offs

### Benefits

| Area | Benefit |
|------|---------|
| **Maintainability** | Clear ownership per AI agent — `packages/claude-code/`, `packages/copilot/`, `packages/opencode/`; you always know exactly where to look or change |
| **Maintainability** | All related changes land in one PR — updating a skill _and_ the CLI behaviour that depends on it never requires coordinating across repos |
| **Maintainability** | One issue tracker, one CI pipeline, one release workflow — no "which repo do I file this in?" confusion |
| **Maintainability** | Adding a new AI agent requires one `AgentDef` entry in `agent_config.go` plus one new `packages/<agent>/` directory — no existing code to touch |
| **Content** | Skills and commands update without a CLI release — edit a `.md` file, commit; next `crit setup-*` run picks it up immediately |
| **Content** | Zero binary bloat — no embedded Markdown in the compiled binary |
| **Content** | Symlink in `packages/copilot/` eliminates duplicate SKILL.md files — one canonical source, two agents |
| **Content** | Non-Go contributors (technical writers, AI prompt engineers) can improve skills with nothing more than a text editor and a PR |
| **Developer experience** | One `git clone` to get everything — no cross-repo context switching during development |
| **Developer experience** | Fast iteration: change a SKILL.md → run `crit setup-claude --source ./packages/claude-code/skills/` → instant feedback, no rebuild |
| **Developer experience** | Developer chooses which AI agent to configure locally — no vendor lock-in; `init-claude`, `init-copilot`, `init-opencode` are all optional convenience tasks |
| **Developer experience** | Overriding the skills source is first-class: `--source` flag, `CRIT_SKILLS_DIR` env var, or `~/.config/crit/config.yaml` — works for local paths, remote URLs, and pinned git refs |
| **Developer experience** | Marketplace `marketplace.json` and plugin files are always in sync — no cross-repo PR needed to keep them consistent |
| **Cleaner root** | Repository root contains only marketplace metadata and top-level files; implementation detail directories (`internal/`, `cmd/`) move to `packages/crit-cli/` |

### Costs / Mitigations

| Cost | Mitigation |
|------|-----------|
| Symlinks require `.gitattributes` entry and contributors need to be aware of them | Add `task check-symlinks` (or a pre-commit hook) that validates symlinks are intact and correctly listed in `.gitattributes`; run in CI |
| Config file needed to override the default source URL | Optional for most users — hardcoded defaults work out-of-the-box; config only needed for pinning to a version or pointing at a fork |
| Migration effort to reorganise current files into `packages/` layout | One-time, mechanical move; see [Migration Steps](#migration-steps) |
| `crit setup-*` requires network access by default | AI-agent tooling requires internet access as a baseline; offline users set `--source` to a local directory — this is a documented first-class feature, not a workaround |

---

## Comparison with Current Structure

| Concern | Current | Monorepo alternative |
|---------|---------|---------------------|
| Skills source of truth | Two copies (`internal/cli/skill/` + `plugin/crit/skills/`) | One copy in `packages/claude-code/skills/` |
| Binary embedding | `//go:embed` in `setup_claude.go`, `setup_opencode.go` | **Removed** — download at runtime via `agent_config.go` + `FetchManifest` |
| AI agent directories | Mixed under `plugin/` and `internal/cli/` | Dedicated `packages/<agent>/` |
| CLI source location | Repository root | `packages/crit-cli/` |
| Demo/asset location | `demo/` and `assets/` at root | `docs/demo/` and `docs/assets/` |
| Marketplace source | Relative path `./plugin/crit` | Relative `./packages/claude-code` |
| `[Layer D:opencode]` | `.opencode/` committed | `.opencode/` git-ignored; created by `task init-opencode` |
| Skill/command update | Requires new CLI release | Update files in repo — CLI picks up on next `setup-*` run |
