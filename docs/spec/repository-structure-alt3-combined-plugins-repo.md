---
title: "Alternative Repository Structure — CLI + Combined Plugins Repo"
description: "A pragmatic middle-ground: keep the CLI in one repo and all AI-agent plugin files (Claude Code, Copilot, opencode) in a single, separate crit-plugins repository.  No Go knowledge needed to contribute to plugins."
last_updated: "2025-03-09"
---

# Alternative Repository Structure — CLI + Combined Plugins Repo

This document proposes splitting the single crit repository into **two focused repositories**:

1. **`tobiashochguertel/crit`** — the CLI binary (Go source, build tooling, release config).
2. **`tobiashochguertel/crit-plugins`** — all AI-agent plugin files: Claude Code marketplace,
   Copilot skills, and opencode commands.

This is the middle ground between:

- **[Alt 1 — Monorepo](./repository-structure-alt1-monorepo.md)** — everything in one repo.
- **[Alt 2 — Multi-repo](./repository-structure-alt2-multi-repo.md)** — a separate repo per AI agent.

> **When to choose Alt 3:**  
> You want to decouple plugin content from the CLI release cycle, but you don't need (or
> don't yet have) separate maintainer teams for each AI agent.  One combined plugin repo
> is easier to manage than three, while still giving all the runtime-download benefits.

See [`repository-structure.md`](./repository-structure.md) for the current structure and
the explanation of the `[Layer …]` label scheme.

---

## Goals

| Goal | How Alt 3 achieves it |
|------|-----------------------|
| Decouple plugin distribution from CLI releases | Skill/command files live in `crit-plugins`; CLI fetches them at runtime |
| No binary embedding | `//go:embed` is removed; CLI only stores a default URL pointing to `crit-plugins` |
| Non-Go contributors can own plugin content | `crit-plugins` has no Go source — just Markdown, JSON, and YAML |
| Minimal overhead vs. full multi-repo | Two repos instead of four; one clone for CLI work, one for plugin work |
| Claude Code marketplace stays functional | `crit-plugins` hosts `.claude-plugin/marketplace.json` and the plugin directory |
| Single point of truth for skills | One `skills/` directory; `copilot/` symlinks into it |

---

## Proposed Repository Set

| Repository | Purpose | Primary language |
|------------|---------|-----------------|
| `tobiashochguertel/crit` | CLI binary — `crit review`, `crit status`, `crit setup-*` | Go |
| `tobiashochguertel/crit-plugins` | All AI-agent plugin files: Claude Code marketplace, Copilot skills, opencode commands | Markdown / JSON / YAML |

---

## Repository Trees

### `tobiashochguertel/crit` (CLI only)

```tree
crit/
├── .github/
│   └── workflows/
│       └── release.yml                    # GoReleaser CI
│
├── .gitignore                              # Ignores dist/, .task/, .claude/, .copilot/, .opencode/
├── .goreleaser.yaml
├── .mise.toml                              # Pins Go version
│
├── cmd/
│   └── crit/
│       └── main.go
│
├── internal/
│   └── cli/
│       ├── agent_config.go                # [Layer C] AgentDef registry — only file to edit per new agent
│       ├── setup.go                       # [Layer C] makeSetupCmd factory; installSkills / installCommands
│       ├── setup_agents.go                # [Layer C] init() — registers all setup-* commands
│       ├── source.go                      # [Layer C] FetchFile, FetchManifest, Config, LoadConfig
│       ├── review.go
│       ├── status.go
│       └── root.go
│
├── Taskfile.yml                            # lint, fmt, build, test, clean, check-symlinks (delegated)
│
├── go.mod
├── go.sum
│
├── docs/
│   ├── ai-agent-plugins/
│   │   ├── claude-code.md                 # Install docs for Claude Code
│   │   ├── copilot.md                     # Install docs for GitHub Copilot CLI
│   │   └── opencode.md                    # Install docs for opencode
│   └── spec/
│       ├── repository-structure.md
│       ├── repository-structure-alt1-monorepo.md
│       ├── repository-structure-alt2-multi-repo.md
│       └── repository-structure-alt3-combined-plugins-repo.md   # ← this file
│
├── CHANGELOG.md
├── LICENSE
└── README.md
```

`DefaultSkillsURL` and `DefaultCommandsURL` in `source.go` point to `crit-plugins`:

```go
const (
    DefaultSkillsURL   = "https://raw.githubusercontent.com/tobiashochguertel/crit-plugins/main/claude-code/skills"
    DefaultCommandsURL = "https://raw.githubusercontent.com/tobiashochguertel/crit-plugins/main/opencode/commands"
)
```

---

### `tobiashochguertel/crit-plugins` (all plugin content)

```tree
crit-plugins/
│
├── .claude-plugin/                        # [Layer A] Marketplace host for Claude Code
│   ├── marketplace.json                   #   Lists claude-code/ as available plugin
│   └── plugin.json                        #   [Layer A:claude-code] Direct-install fallback
│
├── claude-code/                           # [Layer B:claude-code] Plugin for Claude Code
│   ├── .claude-plugin/
│   │   └── plugin.json                    #   Plugin metadata
│   ├── commands/                          #   Claude slash-commands (Markdown)
│   │   ├── crit-review.md
│   │   ├── crit-code-review.md
│   │   └── crit-plan-review.md
│   └── skills/                            # ★ Canonical SKILL.md source (one copy)
│       ├── manifest.yaml                  #   Lists all skills for FetchManifest
│       ├── crit-review/
│       │   └── SKILL.md
│       ├── crit-code-review/
│       │   └── SKILL.md
│       └── crit-plan-review/
│           └── SKILL.md
│
├── copilot/                               # [Layer B:copilot] Skills for GitHub Copilot CLI
│   └── skills -> ../claude-code/skills    # ★ Symlink — identical format, zero duplication
│
└── opencode/                              # [Layer B:opencode] Commands for opencode
    └── commands/
        ├── manifest.yaml                  #   Lists all commands for FetchManifest
        ├── crit-review.md
        ├── crit-code-review.md
        └── crit-plan-review.md
```

The symlink `copilot/skills -> ../claude-code/skills` means both `setup-claude` and
`setup-copilot` read from the same canonical source inside this repo.

---

## How `crit setup-*` works in Alt 3

```
User runs: crit setup-claude

1. source.go: ResolveSource() → DefaultSkillsURL
   "https://raw.githubusercontent.com/tobiashochguertel/crit-plugins/main/claude-code/skills"

2. source.go: FetchManifest(url) → reads manifest.yaml from crit-plugins
   Returns list of skill directory names

3. setup.go: installSkills()
   For each skill in manifest:
     FetchFile(url + "/" + skill + "/SKILL.md") → write to ~/.claude/skills/<skill>/SKILL.md

Override options (first non-empty wins):
  --source ./local-crit-plugins/claude-code/skills   ← local path (offline)
  CRIT_SKILLS_DIR=~/my-skills                         ← env var
  skills_url: … in ~/.config/crit/config.yaml          ← config file
  (default: DefaultSkillsURL above)
```

---

## Versioning and Compatibility

```
crit CLI v1.1.0  ─ downloads from ─► crit-plugins main (latest)
crit CLI v1.0.0  ─ pinned via config ─► crit-plugins @ tag v1.0.0
```

Options for version pinning:

| Method | How |
|--------|-----|
| Default (latest) | `DefaultSkillsURL` points to `main` branch |
| Pin in config file | `skills_url: https://…/crit-plugins/v1.0.0/claude-code/skills` |
| Pin via `--source` | `crit setup-claude --source https://…/crit-plugins/v1.0.0/claude-code/skills` |
| Offline / local | `crit setup-claude --source ./local-plugins/claude-code/skills` |

---

## Trade-offs

### Benefits

| Area | Benefit |
|------|---------|
| **Maintainability** | Plugin content is completely decoupled from CLI source — skill text changes never touch Go code |
| **Maintainability** | Clear two-repo boundary: `crit` = compiled binary; `crit-plugins` = content — no ambiguity about where a change belongs |
| **Maintainability** | `crit-plugins` has no build tooling, no release workflow, no Go modules — it is just files in directories |
| **Content** | Skills and commands update without a CLI release — commit to `crit-plugins`, next `crit setup-*` run picks it up |
| **Content** | Zero binary bloat — no embedded Markdown in the compiled binary |
| **Content** | Symlink eliminates Copilot duplicate — one canonical `skills/` directory, two agents |
| **Content** | Adding a new AI agent in `crit-plugins` only requires a new directory; adding a new agent to the CLI only requires a new `AgentDef` in `agent_config.go` |
| **Contributor experience** | Non-Go contributors can own `crit-plugins` with only Markdown and JSON knowledge |
| **Contributor experience** | `crit-plugins` PR reviews are done by content experts, not Go engineers |
| **Contributor experience** | Plugin repo is tiny and easy to understand — no build system to learn |
| **Developer experience** | Only two repos to manage vs. four (Alt 2) — lower overhead |
| **Developer experience** | `crit setup-claude --source ./local-plugins/…` lets developers test content changes instantly without committing |
| **Flexibility** | Users can fork `crit-plugins` and point their local `config.yaml` at their fork — full customisation without touching the CLI |
| **Flexibility** | Works offline: set `--source` to a local directory |

### Costs / Mitigations

| Cost | Mitigation |
|------|-----------|
| **Two repos to clone** for full development | Most contributors only need one repo; CLI devs never need `crit-plugins` unless they want to test content changes locally |
| **Cross-repo coordination** when adding a new skill format that requires a CLI change | Document a compatibility matrix in `crit` README; use versioned URLs in `config.yaml` to pin plugin content per CLI release |
| **Symlink in `crit-plugins`** needs `.gitattributes` and awareness | Add `task check-symlinks` in `crit-plugins` Taskfile; run in CI; document in CONTRIBUTING |
| **Config file** needed to override default source URL | Optional; hardcoded defaults work for most users; config only needed for pinning or local override |

---

## Comparison across all structures

| Concern | Current | Monorepo (Alt 1) | Multi-repo (Alt 2) | **Combined plugins (Alt 3)** |
|---------|---------|------------------|--------------------|------------------------------|
| Repository count | 1 | 1 | 4 | **2** |
| Skills/command source | `internal/cli/skill/` (embedded) | `packages/claude-code/skills/` | `crit-claude-code` repo | **`crit-plugins/claude-code/skills/`** |
| Binary embedding | `//go:embed` | Removed | Removed | **Removed** |
| Skill update requires CLI release | ✅ Yes | ❌ No | ❌ No | **❌ No** |
| Copilot duplication | Two copies | Symlink | Canonical repo | **Symlink** |
| Non-Go contributor can update skills | ❌ No | Possible (no Go needed for `packages/claude-code/`) | ✅ Yes (`crit-claude-code` repo) | **✅ Yes (`crit-plugins` repo)** |
| Developer onboarding complexity | Low (1 repo) | Low (1 repo) | High (4 repos) | **Medium (2 repos)** |
| Release overhead | Low | Low | High (4 pipelines) | **Low (2 pipelines)** |

---

## Migration Steps (from current structure)

1. **Create `tobiashochguertel/crit-plugins`**
   ```sh
   gh repo create tobiashochguertel/crit-plugins --public --description "Claude Code, Copilot, and opencode plugins for crit"
   ```
2. **Copy plugin files** to `crit-plugins`:
   - `plugin/crit/skills/` → `crit-plugins/claude-code/skills/`
   - `plugin/crit/opencode/*.md` → `crit-plugins/opencode/commands/`
   - `.claude-plugin/` → `crit-plugins/.claude-plugin/`
   - Create `crit-plugins/copilot/skills -> ../claude-code/skills` symlink
   - Add `manifest.yaml` to each content directory
3. **Update `DefaultSkillsURL` and `DefaultCommandsURL`** in `internal/cli/source.go`:
   ```go
   DefaultSkillsURL   = "https://raw.githubusercontent.com/tobiashochguertel/crit-plugins/main/claude-code/skills"
   DefaultCommandsURL = "https://raw.githubusercontent.com/tobiashochguertel/crit-plugins/main/opencode/commands"
   ```
4. **Remove plugin files** from `crit`: `plugin/`, `.claude-plugin/`
5. **Update `README.md`** — add link to `crit-plugins`; update install instructions
6. **Tag `crit-plugins v1.0.0`** and update `crit` README with compatibility matrix
7. **Update `crit` CHANGELOG** — document the split

> **Steps 3–5 are straightforward** because `//go:embed` has already been removed.
> The only code change is updating the two URL constants in `source.go`.

---

## Summary

Alt 3 gives you the key benefit of Alt 2 (plugin content fully decoupled from the CLI binary) with the simplicity of Alt 1 (one fewer repo to manage).

The `crit` repository stays a lean Go tool.  The `crit-plugins` repository is pure content —
no build tooling, no Go source, no release pipeline complexity.  Non-Go contributors can
own and maintain plugin content independently.

Because `//go:embed` is already removed from the CLI source and the runtime-download
infrastructure is already in place (`source.go`, `agent_config.go`, `FetchManifest`),
migration requires only two constants updated in `source.go` and a file reorganisation.

**Choose Alt 3** if you want clean separation without the overhead of managing four repos
(Alt 2) or without keeping plugin files inside the same repo as the CLI source (Alt 1).
