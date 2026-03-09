---
title: "Claude Code Integration"
description: "Install and use crit with Claude Code via skills or the Plugin Marketplace."
last_updated: "2025-03-09"
---

# Claude Code Integration

`crit` integrates with [Claude Code](https://claude.ai/code) in two ways:

1. **Plugin Marketplace** (recommended) — install with one command, no binary required
2. **Standalone skill install** — run `crit setup-claude` after installing the binary

---

## Option A: Plugin Marketplace (recommended)

Claude Code's Plugin Marketplace lets you install crit without the Go binary.

### Install via marketplace

```
/plugin marketplace add tobiashochguertel/crit
/plugin install crit
```

This installs the plugin from `plugin/crit/` in the fork, which includes:
- Slash commands: `/crit:review`, `/crit:code-review`, `/crit:plan-review`
- Skills (for GitHub Copilot CLI skill discovery): `plugin/crit/skills/`

### Or install directly

```
/plugin install https://github.com/tobiashochguertel/crit
```

No marketplace step needed — Claude Code treats the repo root as a plugin.

### Verify installation

After install, the commands appear in Claude Code's command palette:

```
/crit:review           # Routes to code or plan review
/crit:code-review      # Multi-file code review
/crit:plan-review      # Single-file document / plan review
```

---

## Option B: Standalone skill install (binary required)

If you prefer not to use the Plugin Marketplace, install `crit` and run the setup command.

### Install the binary

```bash
go install github.com/kevindutra/crit/cmd/crit@latest
# or from the fork (includes copilot/opencode support):
go install github.com/tobiashochguertel/crit/cmd/crit@latest
```

Make sure `$(go env GOPATH)/bin` is in your `PATH`.

### Install skills

```bash
crit setup-claude          # Global: installs to ~/.claude/skills/
crit setup-claude --project # Project-scoped: installs to .claude/skills/
crit setup-claude --force   # Overwrite existing skill files
```

**Installed files:**

| File | Path (global) |
|------|--------------|
| `crit-review/SKILL.md` | `~/.claude/skills/crit-review/SKILL.md` |
| `crit-code-review/SKILL.md` | `~/.claude/skills/crit-code-review/SKILL.md` |
| `crit-plan-review/SKILL.md` | `~/.claude/skills/crit-plan-review/SKILL.md` |

### Verify installation

```
/crit-review           # Routes to code or plan review
/crit-code-review      # Multi-file code review
/crit-plan-review      # Single-file document / plan review
```

Note: standalone skills use the `/crit-review` prefix (no namespace); marketplace-installed
commands use `/crit:review` (with namespace).

---

## Usage

### Code review workflow

```
/crit:code-review
```

Claude Code will:
1. Ask you to run `crit review --code` in your terminal (requires tmux)
2. Open the TUI — navigate files, leave inline comments, quit when done
3. Run `crit status --code` to read your comments as JSON
4. Address all comments across all reviewed files

### Document / plan review

```
/crit:plan-review docs/plan.md
```

Claude Code will:
1. Ask you to run `crit review docs/plan.md` in your terminal
2. Open the TUI for single-file review with inline comments
3. Run `crit status docs/plan.md` to read your comments
4. Revise the document based on your feedback

### Auto-route

```
/crit:review
```

Asks whether you want to review code changes or a document, then dispatches to the
appropriate skill.

---

## Skills reference

### `crit-review` / `/crit:review`

Router skill. Asks what to review and calls either `crit-code-review` or `crit-plan-review`.

### `crit-code-review` / `/crit:code-review`

Multi-file code review using `crit review --code`.

**Requirements:**
- `crit` binary on `PATH`
- tmux session (for split-pane TUI)

**What it does:**
1. Detects changed files in the git repo
2. Opens a tabbed TUI with syntax-highlighted diffs and line markers
3. You leave inline comments; close TUI when done
4. Claude reads JSON output from `crit status --code`
5. Addresses every comment, file by file

### `crit-plan-review` / `/crit:plan-review`

Single-file document review. Accepts a file path as argument.

**Example:**
```
/crit:plan-review docs/PLAN.md
```

---

## Requirements

| Requirement | Notes |
|-------------|-------|
| **tmux** | Required for the split-pane TUI. Start with `tmux new -s work` before launching Claude Code. |
| **crit binary** | Required for Option B (standalone). Not needed for marketplace install. |
| **Go 1.21+** | Only needed if building from source. |

### Starting a tmux session

If you're not in tmux, `crit review` will warn you. Start a session first:

```bash
tmux new -s work
claude   # Launch Claude Code inside tmux
```

---

## Plugin manifest (for reference)

The marketplace plugin manifest at `plugin/crit/.claude-plugin/plugin.json`:

```json
{
  "name": "crit",
  "description": "Review markdown documents with an interactive TUI. Leave inline comments, then let Claude address the feedback automatically.",
  "version": "1.0.2",
  "author": { "name": "Kevin Dutra" },
  "homepage": "https://github.com/kevindutra/crit",
  "repository": "https://github.com/kevindutra/crit",
  "license": "MIT",
  "keywords": ["review", "markdown", "tui", "diff", "critique"],
  "skills": "skills/"
}
```

The `"skills": "skills/"` field exposes `plugin/crit/skills/` for GitHub Copilot CLI skill
discovery when the plugin is installed from the marketplace.

---

## Difference between Option A and Option B

| | Marketplace (A) | Standalone skills (B) |
|--|----------------|----------------------|
| **Install** | `/plugin install` | `crit setup-claude` |
| **Binary required** | No | Yes |
| **Command prefix** | `/crit:review` (namespaced) | `/crit-review` (flat) |
| **Skills path** | `~/.claude/plugins/crit/` | `~/.claude/skills/` |
| **Updates** | `/plugin marketplace update` | Re-run `crit setup-claude --force` |

---

## Troubleshooting

**Skills not found after `crit setup-claude`:**  
Verify files exist: `ls ~/.claude/skills/crit-*/`. Re-run with `--force`.

**Plugin commands not appearing:**  
Run `/plugin list` to confirm crit is installed. Try `/plugin uninstall crit` and reinstall.

**`crit: command not found`:**  
Ensure `$(go env GOPATH)/bin` is in `PATH`. Add `export PATH="$PATH:$(go env GOPATH)/bin"` to your shell config.

**tmux not found:**  
Install tmux: `brew install tmux` (macOS) or `apt install tmux` (Linux).

**Split pane doesn't open:**  
You must already be inside a tmux session. Claude Code sends the `crit review` command to a new tmux pane — this only works if you launched Claude Code from within tmux.
