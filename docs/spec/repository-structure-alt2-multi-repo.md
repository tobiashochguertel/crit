---
title: "Alternative Repository Structure — Multi-Repo Split"
description: "A proposed split of the crit repository into separate repositories for the CLI tool, the Claude Code plugin, and other AI agent integrations.  The CLI binary embeds no content — it downloads skills and commands from the respective plugin repos at setup time."
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
| CLI setup commands stay useful | `crit setup-*` downloads from the respective plugin repos at install time |
| No binary embedding | Content lives in plugin repos; CLI fetches it via HTTP or local path |

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
│   │   ├── setup.go                  # Shared installer helpers (FetchFile, resolveTargetDir)
│   │   ├── setup_claude.go           # [Layer C:claude-code] Downloads skills at setup time
│   │   ├── setup_copilot.go          # [Layer C:copilot]     Downloads skills at setup time
│   │   ├── setup_opencode.go         # [Layer C:opencode]    Downloads commands at setup time
│   │   ├── source.go                 # Config, LoadConfig, ResolveSource, FetchFile
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

**Key properties of `crit` in the multi-repo model:**

- **Zero embedded markdown** — no `//go:embed`, no `internal/cli/skill/`,
  no `internal/cli/opencode/` directories committed.
- **`source.go`** provides `FetchFile(source, relPath)` which handles both HTTP URLs
  and local filesystem paths (with `~/` expansion).
- **Default URLs** in `source.go` point to the respective plugin repos:

```go
const (
    DefaultSkillsURL   = "https://raw.githubusercontent.com/tobiashochguertel/crit-claude-code/main/plugin/skills"
    DefaultCommandsURL = "https://raw.githubusercontent.com/tobiashochguertel/crit-opencode/main/commands"
)
```

- **Source resolution order** (first non-empty wins):

```
--source <path|url>        CLI flag (local dir or HTTP(S) URL base)
$CRIT_SKILLS_DIR           Environment variable (for skills)
$CRIT_OPENCODE_DIR         Environment variable (for opencode commands)
skills_url: in config      ~/.config/crit/config.yaml
commands_url: in config    ~/.config/crit/config.yaml
(default URL)              (constants above)
```

> **Note — `.opencode/` is git-ignored.**  Run `task init-opencode` to create
> `.opencode/commands/` locally as **[Layer D:opencode]**.  Similarly `task init-claude`
> and `task init-copilot` create project-local AI agent config for contributors who
> use those tools while working on the CLI.

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
│   └── skills/                       # ★ CANONICAL source for SKILL.md files
│       ├── crit-review/              #   Downloaded by `crit setup-claude` and `crit setup-copilot`
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

`DefaultSkillsURL` in `source.go` points to:
```
https://raw.githubusercontent.com/tobiashochguertel/crit-claude-code/main/plugin/skills
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
│   ├── crit-review/                  #   Downloaded by `crit setup-copilot`
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

> When Copilot CLI skills are format-compatible with Claude Code skills,
> `crit-copilot` can simply reference `crit-claude-code/plugin/skills` via GitHub
> Actions sync, avoiding duplication.  See the [symlink approach in alt1](./repository-structure-alt1-monorepo.md#2-packagescopolit--symlinked-skills).

### `tobiashochguertel/crit-opencode` (opencode custom commands)

```tree
crit-opencode/
├── commands/                         # [Layer C:opencode] opencode command files
│   ├── crit-review.md                #   Downloaded by `crit setup-opencode`
│   ├── crit-code-review.md
│   └── crit-plan-review.md
│
├── docs/
│   └── installation.md
│
└── README.md                         # Copy to ~/.config/opencode/commands/ or use `crit setup-opencode`
```

`DefaultCommandsURL` in `source.go` points to:
```
https://raw.githubusercontent.com/tobiashochguertel/crit-opencode/main/commands
```

---

## How `setup-*` works in the multi-repo model

`crit setup-claude` (and the other `setup-*` subcommands) call `FetchFile(source, relPath)` for
each file to install:

```go
// source.go — used by all three setup_*.go files
func FetchFile(source, relPath string) ([]byte, error) {
    if IsURL(source) {
        url := strings.TrimRight(source, "/") + "/" + relPath
        resp, err := http.Get(url)
        // ...
        return io.ReadAll(resp.Body)
    }
    // Local path — useful during development or offline use
    return os.ReadFile(filepath.Join(source, filepath.FromSlash(relPath)))
}
```

This means:

- **Online (default):** CLI fetches the latest file from the plugin repo over HTTPS.
- **Offline / dev:** Set `$CRIT_SKILLS_DIR` to a local checkout of `crit-claude-code`
  or write the path to `~/.config/crit/config.yaml`.
- **Custom / self-hosted:** Point to any HTTP server or Gitea/GitHub instance that serves
  the same directory structure.

**No new CLI release is needed when a skill description changes** — the next time a user
runs `crit setup-claude`, the updated file is downloaded automatically.

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
│   └── skills/                       # ★ Canonical SKILL.md source
│       ├── crit-review/SKILL.md
│       ├── crit-code-review/SKILL.md
│       └── crit-plan-review/SKILL.md
│
├── copilot/                          # [Layer B:copilot] Skills for GitHub Copilot CLI
│   └── skills -> ../claude-code/skills  # ★ Symlink — same format, no duplication
│
└── opencode/                         # [Layer C:opencode] Commands for opencode
    └── commands/
        ├── crit-review.md
        ├── crit-code-review.md
        └── crit-plan-review.md
```

Default URLs in `source.go` would then be:
```go
const (
    DefaultSkillsURL   = "https://raw.githubusercontent.com/tobiashochguertel/crit-plugins/main/claude-code/skills"
    DefaultCommandsURL = "https://raw.githubusercontent.com/tobiashochguertel/crit-plugins/main/opencode/commands"
)
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
| Pinned URLs in `config.yaml` | Users can lock to a specific tag via `?ref=v1.0.2` |

**Recommendation for now:** Document the matrix in `README.md`.  Users who need a pinned
version can override `skills_url` in `~/.config/crit/config.yaml`.

---

## Trade-offs vs. Monorepo

| Concern | Monorepo (Alt 1) | Multi-repo (Alt 2) |
|---------|-----------------|-------------------|
| **Setup complexity** | One repo, one clone | Four repos, cross-repo coordination |
| **Skills duplication** | Solved by symlink | Solved by separate canonical repos |
| **Plugin updates without CLI release** | ✅ Yes (runtime download) | ✅ Yes (runtime download) |
| **Claude Code `git-subdir` for plugins** | Points inside the monorepo | Points to dedicated plugin repo |
| **Contributor onboarding** | Clone one repo | Must find the right repo |
| **Release process** | Single GoReleaser workflow | Per-repo releases |
| **Breaking the "marketplace + CLI" coupling** | Partial (still one repo) | ✅ Complete separation |
| **`[Layer D:opencode]`** | git-ignored; `task init-opencode` | git-ignored; `task init-opencode` |
| **Binary size** | Small (no embedded markdown) | Small (no embedded markdown) |

---

## Migration Steps (from current structure)

1. **Create `tobiashochguertel/crit-claude-code`** — `gh repo create tobiashochguertel/crit-claude-code --public`
2. **Copy plugin files** — `plugin/crit/` + `.claude-plugin/` → new repo
3. **Create `tobiashochguertel/crit-copilot`** — copy `plugin/crit/skills/` → `skills/`
4. **Create `tobiashochguertel/crit-opencode`** — copy `plugin/crit/opencode/*.md` → `commands/`
5. **Update `DefaultSkillsURL` and `DefaultCommandsURL`** in `source.go` to point to the new repos
6. **Remove `plugin/`, `.claude-plugin/`** from `crit` (no longer needed in the CLI repo)
7. **Update `README.md`** — add links to each plugin repo; update install instructions
8. **Tag first release** on each plugin repo (`v1.0.0`)
9. **Update `crit` CHANGELOG** — document the split

> **Steps 5 and 6 are straightforward** because `//go:embed` has already been removed
> from the CLI source.  The only code change is updating the two URL constants.

---

## Summary

The multi-repo split gives the cleanest separation between the CLI tool and its AI agent
integrations.  Because `//go:embed` is already removed from the CLI source, the only
remaining coupling is the two `DefaultSkillsURL` / `DefaultCommandsURL` constants in
`source.go`.  Updating those constants and creating the plugin repos is all that is needed
to complete the transition.

The minimal `crit-plugins` variant (one plugin repo instead of three) is a pragmatic
middle ground that reduces overhead while still decoupling the CLI binary from the plugin
distribution.
