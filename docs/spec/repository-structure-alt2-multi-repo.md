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
├── go.mod                            # module github.com/kevindutra/crit
├── go.sum
├── README.md                         # Focus on CLI usage; link to plugin repos for IDE setup
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
│   │   ├── setup_claude.go           # Downloads or embeds stubs; see "Setup commands" below
│   │   ├── setup_copilot.go
│   │   ├── setup_opencode.go
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

**What is removed compared to today:**
- `plugin/` directory — moved to `crit-claude-code`
- `.claude-plugin/` at root — moved to `crit-claude-code`
- `internal/cli/skill/` embedded skills — replaced by download-at-install approach
- `internal/cli/opencode/` embedded commands — replaced by download-at-install approach
- `.opencode/commands/` — moved to `crit-opencode`
- `docs/ai-agent-plugins/` — kept but now just link to each plugin repo's own docs

### `tobiashochguertel/crit-claude-code` (Claude Code marketplace + plugin)

```tree
crit-claude-code/
├── .claude-plugin/                   # [Layer A] Makes this repo a marketplace host
│   ├── marketplace.json              #   Lists the crit plugin (points to ./plugin/)
│   └── plugin.json                  #   Fallback: direct-install plugin manifest
│
├── plugin/                           # [Layer B] The installable plugin package
│   ├── .claude-plugin/
│   │   └── plugin.json               #   Plugin manifest (name, version, skills, commands)
│   ├── commands/
│   │   ├── review.md
│   │   ├── code-review.md
│   │   └── plan-review.md
│   └── skills/                       # ★ SINGLE SOURCE OF TRUTH for skills
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
├── skills/
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
└── README.md                         # Explains: copy to ~/.copilot/skills/ or use `crit setup-copilot`
```

### `tobiashochguertel/crit-opencode` (opencode custom commands)

```tree
crit-opencode/
├── commands/
│   ├── crit-review.md
│   ├── crit-code-review.md
│   └── crit-plan-review.md
│
├── docs/
│   └── installation.md
│
└── README.md                         # Explains: copy to ~/.config/opencode/commands/ or use `crit setup-opencode`
```

---

## How the CLI `setup-*` commands work in the split model

In the monorepo or current structure, `setup_claude.go` uses `//go:embed` to bundle
the skill files directly in the binary.  With a multi-repo split, three approaches are
possible:

### Approach 1: Static embed with pinned version (recommended)

Each `setup_*.go` file embeds a **minimal stub** that points users to the latest
skills/commands, or bundles a pinned snapshot at build time via a script:

```
# Taskfile.yml — fetch step run during release (not on every build)
fetch-skills:
  desc: "Download skill files from crit-claude-code at the tagged release version"
  cmds:
    - rm -rf internal/cli/skill/
    - mkdir -p internal/cli/skill/crit-review internal/cli/skill/crit-code-review internal/cli/skill/crit-plan-review
    - curl -sSL "https://raw.githubusercontent.com/tobiashochguertel/crit-claude-code/v1.0.2/plugin/skills/crit-review/SKILL.md"       -o internal/cli/skill/crit-review/SKILL.md
    - curl -sSL "https://raw.githubusercontent.com/tobiashochguertel/crit-claude-code/v1.0.2/plugin/skills/crit-code-review/SKILL.md" -o internal/cli/skill/crit-code-review/SKILL.md
    - curl -sSL "https://raw.githubusercontent.com/tobiashochguertel/crit-claude-code/v1.0.2/plugin/skills/crit-plan-review/SKILL.md" -o internal/cli/skill/crit-plan-review/SKILL.md
```

The fetched files are committed to the CLI repo for each release (gitignored in between
releases).  The binary always contains exactly the skills matching the release tag.

**Pros:** Self-contained binary; no runtime network access required.  
**Cons:** Skills and CLI versions are coupled; a skill-only update still requires a CLI
release.

### Approach 2: Runtime download at `setup` time

`setup_claude.go` downloads the latest skills from GitHub at runtime when the user runs
`crit setup-claude`:

```go
const skillsBaseURL = "https://raw.githubusercontent.com/tobiashochguertel/crit-claude-code/main/plugin/skills/"

func downloadSkill(name, targetPath string) error {
    url := skillsBaseURL + name + "/SKILL.md"
    resp, err := http.Get(url)
    // ... write to targetPath
}
```

**Pros:** Skills stay up to date independently of the CLI version; plugin repo can ship
improvements without a CLI release.  
**Cons:** Requires internet access at setup time; `go test` needs mocking for network
calls.

### Approach 3: Separate `crit-setup` binary or scripts (no embedding)

Skip embedding altogether and have each plugin repo include a simple install script:

```bash
# crit-claude-code/install.sh
cp -r plugin/skills/* ~/.claude/skills/
```

Users run:
```bash
curl -sSL https://raw.githubusercontent.com/tobiashochguertel/crit-claude-code/main/install.sh | bash
```

**Pros:** CLI binary is completely decoupled from plugin files.  
**Cons:** More friction for users; shell-pipe installs carry security risks.

**Recommendation:** Use **Approach 1** (static embed with `fetch-skills` task) for simplicity
and reproducible releases.  Move to **Approach 2** once the skill format stabilizes.

---

## Minimal variant — `tobiashochguertel/crit-plugins`

If managing four repositories is too much overhead, merge all plugin repos:

```tree
crit-plugins/
├── .claude-plugin/
│   ├── marketplace.json
│   └── plugin.json
│
├── claude-code/                      # Plugin for Claude Code
│   ├── .claude-plugin/
│   │   └── plugin.json
│   ├── commands/
│   └── skills/
│
├── copilot/                          # Skills for GitHub Copilot CLI
│   └── skills/
│
├── opencode/                         # Commands for opencode
│   └── commands/
│
└── README.md
```

`marketplace.json` references `./claude-code/` as the plugin source.

The CLI `setup-*` commands use the pinned-embed approach, fetching from this single
`crit-plugins` repo.

---

## Cross-repository versioning

When using multiple repos, versioning must be coordinated:

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

**Recommendation for now:** Document the matrix in `README.md`.  Move to git submodules if
the plugin and CLI drift apart frequently.

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
| **PR noise** | All changes mixed together | Plugin changes isolated |

---

## Migration Steps (from current structure)

1. **Create `tobiashochguertel/crit-claude-code`** — `gh repo create tobiashochguertel/crit-claude-code --public`
2. **Copy plugin files** — `plugin/crit/` + `.claude-plugin/` → new repo
3. **Create `tobiashochguertel/crit-copilot`** — copy `plugin/crit/skills/` → `skills/`
4. **Create `tobiashochguertel/crit-opencode`** — copy `internal/cli/opencode/*.md` + `.opencode/commands/` → `commands/`
5. **Add `fetch-skills` Taskfile target** in `crit` repo
6. **Run `task fetch-skills`** to populate `internal/cli/skill/` from the new repos
7. **Remove `plugin/`, `.claude-plugin/`, `internal/cli/opencode/`, `.opencode/`** from `crit`
8. **Update README.md** — add links to each plugin repo; update install instructions
9. **Tag first release** on each plugin repo (`v1.0.0`)
10. **Update `crit` CHANGELOG** — document the split

---

## Summary

The multi-repo split gives the cleanest separation between the CLI tool and its AI agent
integrations.  The cost is cross-repository coordination and a slightly more complex
`setup-*` implementation.  The minimal `crit-plugins` variant (one plugin repo instead
of three) is a pragmatic middle ground that reduces overhead while still decoupling the
CLI binary from the plugin distribution.
