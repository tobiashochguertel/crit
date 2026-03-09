---
title: "Alternative Repository Structure — Multi-Repo Split"
description: "A proposed split of the crit repository into separate repositories for the CLI tool, the Claude Code plugin, and other AI agent integrations."
last_updated: "2025-03-09"
---

# Alternative Repository Structure — Multi-Repo Split

This document proposes splitting the single crit repository into **multiple focused
repositories**: one for the CLI binary, one for the Claude Code marketplace/plugin,
and one or more for additional AI agent integrations (Copilot CLI, opencode).

See [`repository-structure.md`](./repository-structure.md) for the current structure and
[`repository-structure-alt1-monorepo.md`](./repository-structure-alt1-monorepo.md) for the
monorepo alternative.

---

## Goals

| Goal | How the split achieves it |
|------|--------------------------|
| CLI has zero coupling to plugin distribution | `crit-cli` knows nothing about Claude Code or Copilot — it is a pure Go tool |
| Plugin repos are independently versioned | Skill descriptions can change without a CLI release |
| Each integration can be contributed to independently | Plugin authors need no Go knowledge |
| CLI setup commands stay useful | `crit setup-*` fetches from the respective plugin repos at install time |

---

## Proposed Repository Set

| Repository | Purpose | Primary language |
|------------|---------|-----------------|
| `tobiashochguertel/crit` | CLI binary — `crit review`, `crit status`, `crit setup-*` | Go |
| `tobiashochguertel/crit-claude-code` | Claude Code marketplace + plugin (commands, skills) | Markdown / JSON |
| `tobiashochguertel/crit-copilot` | GitHub Copilot CLI skills | Markdown |
| `tobiashochguertel/crit-opencode` | opencode custom commands | Markdown |

> **Minimal split:** If you want fewer repos, combine all three plugin repos into a single
> `tobiashochguertel/crit-plugins` repository. See [Minimal variant](#minimal-variant) below.

---

## Repository Trees

### `tobiashochguertel/crit` (CLI only)

```tree
crit/
├── .goreleaser.yaml                  # Cross-platform release config
├── .mise.toml                        # Pins Go version
├── CHANGELOG.md
├── Taskfile.yml                      # build, test, lint, format, tidy, clean, all
│                                     # init-claude, init-copilot, init-opencode
├── go.mod                            # module github.com/kevindutra/crit
├── go.sum
├── README.md                         # CLI usage; link to plugin repos for IDE setup
│
├── cmd/
│   └── crit/
│       └── main.go
│
├── internal/
│   ├── cli/
│   │   ├── comment.go
│   │   ├── review.go
│   │   ├── review_test.go
│   │   ├── root.go
│   │   ├── setup.go                  # Shared installer helpers
│   │   ├── setup_claude.go           # [Layer C:claude-code] Downloads/embeds skills
│   │   ├── setup_copilot.go          # [Layer C:copilot]     Downloads/embeds skills
│   │   ├── setup_opencode.go         # [Layer C:opencode]    Downloads/embeds commands
│   │   └── status.go
│   ├── document/
│   ├── git/
│   ├── review/
│   └── tui/
│
├── assets/
│   └── crit_logo.png
│
└── demo/
    ├── demo.gif
    ├── demo.tape
    ├── code-review.gif
    ├── code-review.tape
    └── plan.md
```

> **Note — `.opencode/` is git-ignored.**  Run `task init-opencode` to create
> `.opencode/commands/` locally as **[Layer D:opencode]**.  Similarly `task init-claude`
> and `task init-copilot` create project-local AI agent config for contributors who
> use those tools while working on the CLI.

**What is removed compared to today:**
- `plugin/` directory — moved to `crit-claude-code`
- `.claude-plugin/` at root — moved to `crit-claude-code`
- `internal/cli/skill/` embedded skills — replaced by download-at-install approach
- `internal/cli/opencode/` embedded commands — replaced by download-at-install approach
- `docs/ai-agent-plugins/` — kept but now just links to each plugin repo's own docs

### `tobiashochguertel/crit-claude-code` (Claude Code marketplace + plugin)

```tree
crit-claude-code/
├── .claude-plugin/                   # [Layer A] Makes this repo a marketplace host
│   ├── marketplace.json              #   Lists the crit plugin (points to ./plugin/)
│   └── plugin.json                   #   [Layer A:claude-code] Fallback direct-install manifest
│
├── plugin/                           # [Layer B:claude-code] The installable plugin package
│   ├── .claude-plugin/
│   │   └── plugin.json               #   Plugin manifest (name, version, skills, commands)
│   ├── commands/                     #   [Layer A:claude-code] Slash commands after /plugin install
│   │   ├── review.md
│   │   ├── code-review.md
│   │   └── plan-review.md
│   └── skills/                       # ★ SINGLE SOURCE OF TRUTH for SKILL.md files
│       ├── crit-review/
│       │   └── SKILL.md
│       ├── crit-code-review/
│       │   └── SKILL.md
│       └── crit-plan-review/
│           └── SKILL.md
│
├── docs/
│   └── installation.md               # Claude Code install instructions
│
└── README.md
```

Users add the marketplace:
```
/plugin marketplace add tobiashochguertel/crit-claude-code
/plugin install crit
```

Or install directly:
```
/plugin install https://github.com/tobiashochguertel/crit-claude-code
```

### `tobiashochguertel/crit-copilot` (GitHub Copilot CLI)

```tree
crit-copilot/
├── skills/                           # [Layer B:copilot] SKILL.md files
│   ├── crit-review/
│   │   └── SKILL.md
│   ├── crit-code-review/
│   │   └── SKILL.md
│   └── crit-plan-review/
│       └── SKILL.md
│
├── docs/
│   └── installation.md
│
└── README.md                         # Copy to ~/.copilot/skills/ or use `crit setup-copilot`
```

### `tobiashochguertel/crit-opencode` (opencode custom commands)

```tree
crit-opencode/
├── commands/                         # [Layer C:opencode] opencode command files
│   ├── crit-review.md
│   ├── crit-code-review.md
│   └── crit-plan-review.md
│
├── docs/
│   └── installation.md
│
└── README.md                         # Copy to ~/.config/opencode/commands/ or use `crit setup-opencode`
```

---

## How the CLI `setup-*` commands work in the split model

In the current structure, `setup_claude.go` uses `//go:embed` to bundle skill files
directly in the binary.  With a multi-repo split, three approaches are possible:

### Approach 1: Static embed with pinned version (recommended)

A `fetch-skills` Taskfile task downloads skill files at release time and commits them
to `internal/cli/skill/` + `internal/cli/opencode/`.  The `go:embed` directives remain
unchanged.

```yaml
# Taskfile.yml — fetch step run during release (not on every build)
fetch-skills:
  desc: "Download skill files from crit-claude-code at the tagged release version"
  vars:
    VER: "v1.0.2"
  cmds:
    - rm -rf internal/cli/skill/ internal/cli/opencode/
    - mkdir -p internal/cli/skill/crit-review internal/cli/skill/crit-code-review internal/cli/skill/crit-plan-review
    - mkdir -p internal/cli/opencode
    - curl -sSL "https://raw.githubusercontent.com/tobiashochguertel/crit-claude-code/{{.VER}}/plugin/skills/crit-review/SKILL.md"
           -o internal/cli/skill/crit-review/SKILL.md
    - curl -sSL "https://raw.githubusercontent.com/tobiashochguertel/crit-opencode/{{.VER}}/commands/crit-review.md"
           -o internal/cli/opencode/crit-review.md
```

**Pros:** Self-contained binary; no runtime network access required.  
**Cons:** Skills and CLI versions are coupled; a skill-only update still requires a CLI release.

### Approach 2: Runtime download at `setup` time

`setup_claude.go` downloads the latest skills from GitHub at runtime:

```go
const skillsBaseURL = "https://raw.githubusercontent.com/tobiashochguertel/crit-claude-code/main/plugin/skills/"

func downloadSkill(name, targetPath string) error {
    url := skillsBaseURL + name + "/SKILL.md"
    resp, err := http.Get(url)
    // ... write to targetPath
}
```

**Pros:** Skills stay up to date independently of the CLI version.  
**Cons:** Requires internet access at setup time; tests need network mocking.

### Approach 3: Separate install scripts (no embedding)

Each plugin repo includes an `install.sh`.  Users run:
```bash
curl -sSL https://raw.githubusercontent.com/tobiashochguertel/crit-claude-code/main/install.sh | bash
```

**Pros:** CLI binary is completely decoupled from plugin files.  
**Cons:** More friction for users; shell-pipe installs carry security risks.

**Recommendation:** Use **Approach 1** for simplicity and reproducible releases.
Move to **Approach 2** once the skill format stabilizes.

---

## Minimal variant — `tobiashochguertel/crit-plugins`

If managing four repositories is too much overhead, merge all plugin repos:

```tree
crit-plugins/
├── .claude-plugin/
│   ├── marketplace.json              # [Layer A] Marketplace host
│   └── plugin.json                   # [Layer A:claude-code] Direct-install fallback
│
├── claude-code/                      # [Layer B:claude-code] Plugin for Claude Code
│   ├── .claude-plugin/
│   │   └── plugin.json
│   ├── commands/                     # [Layer A:claude-code]
│   └── skills/                       # ★ Single source of truth for SKILL.md files
│
├── copilot/                          # [Layer B:copilot] Skills for GitHub Copilot CLI
│   └── skills -> ../claude-code/skills  # ★ Symlink — same format, no duplication
│
└── opencode/                         # [Layer C:opencode] Commands for opencode
    └── commands/
```

`marketplace.json` references `./claude-code/` as the plugin source.

---

## Cross-repository versioning

```
crit CLI v1.1.0  ─ tested with ─► crit-claude-code v1.0.2
                               ─ tested with ─► crit-copilot v1.0.1
                               ─ tested with ─► crit-opencode v1.0.0
```

Options for expressing this:

| Method | Mechanism |
|--------|-----------|
| CHANGELOG + manual coordination | Low friction; relies on convention |
| Git tags with a shared prefix | `plugins/v1.0.2` tag in each plugin repo |
| Compatibility matrix in `crit` README | Table mapping CLI ↔ plugin versions |
| Git submodules in `crit` pointing to each plugin repo | `git submodule update --remote` to update |

**Recommendation for now:** Document the matrix in `README.md`.

---

## Trade-offs vs. Monorepo

| Concern | Monorepo (Alt 1) | Multi-repo (Alt 2) |
|---------|-----------------|-------------------|
| **Setup complexity** | One repo, one clone | Four repos, cross-repo coordination |
| **Skills duplication** | Solved by `sync` task | Solved by separate canonical repos |
| **Plugin updates without CLI release** | Not possible (sync is at build time) | ✅ Yes (Approach 2) |
| **Claude Code `git-subdir` for plugins** | Points inside the monorepo | Points to dedicated plugin repo |
| **Contributor onboarding** | Clone one repo | Must find the right repo |
| **Release process** | Single GoReleaser workflow | Per-repo releases |
| **Breaking the "marketplace + CLI" coupling** | Partial (still one repo) | ✅ Complete separation |
| **`[Layer D:opencode]`** | git-ignored; `task init-opencode` | git-ignored; `task init-opencode` |

---

## Migration Steps (from current structure)

1. **Create `tobiashochguertel/crit-claude-code`** — `gh repo create tobiashochguertel/crit-claude-code --public`
2. **Copy plugin files** — `plugin/crit/` + `.claude-plugin/` → new repo
3. **Create `tobiashochguertel/crit-copilot`** — copy `plugin/crit/skills/` → `skills/`
4. **Create `tobiashochguertel/crit-opencode`** — copy `internal/cli/opencode/*.md` → `commands/`
5. **Add `fetch-skills` Taskfile target** in `crit` repo
6. **Run `task fetch-skills`** to populate `internal/cli/skill/` from the new repos
7. **Remove `plugin/`, `.claude-plugin/`, `internal/cli/opencode/`** from `crit`
8. **Update `README.md`** — add links to each plugin repo; update install instructions
9. **Tag first release** on each plugin repo (`v1.0.0`)
10. **Update `crit` CHANGELOG** — document the split

---

## Summary

The multi-repo split gives the cleanest separation between the CLI tool and its AI agent
integrations.  The cost is cross-repository coordination and a slightly more complex
`setup-*` implementation.  The minimal `crit-plugins` variant (one plugin repo instead
of three) is a pragmatic middle ground that reduces overhead while still decoupling the
CLI binary from the plugin distribution.
