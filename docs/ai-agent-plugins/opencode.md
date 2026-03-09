# opencode Integration

`crit` integrates with [opencode](https://opencode.ai/) via the opencode custom commands system.

## Installation

### Global (recommended)

```bash
crit setup-opencode
```

Installs commands to `~/.config/opencode/commands/`:
- `~/.config/opencode/commands/crit-review.md`
- `~/.config/opencode/commands/crit-code-review.md`
- `~/.config/opencode/commands/crit-plan-review.md`

### Project-scoped

```bash
crit setup-opencode --project
```

Installs to `.opencode/commands/` in the current directory. Commands are only available when `opencode` is run from this project.

### Overwrite existing files

```bash
crit setup-opencode --force
```

## Usage

After installation, type `/` in opencode to list available commands, then select a crit command:

```
/crit-review           # Routes to code or plan review
/crit-code-review      # Review code changes
/crit-plan-review docs/PLAN.md   # Review a specific file
```

## Commands

### `/crit-review`

Routes between code review and document/plan review based on user input. Use this when you're unsure which type of review to run.

### `/crit-code-review`

Reviews code changes using `crit review --code`. The command will:
1. Ask you to run `crit review --code` in your terminal
2. Wait for you to confirm
3. Run `crit status --code` to read your comments as JSON
4. Address all comments across all reviewed files

**Requirements:** `crit` binary on PATH

### `/crit-plan-review`

Reviews a specific document. Invoke with a file path argument:

```
/crit-plan-review docs/plan.md
```

The command will:
1. Ask you to run `crit review docs/plan.md` in your terminal
2. Wait for you to confirm
3. Run `crit status docs/plan.md` to read your comments
4. Address all comments in the document

## Command Format

opencode commands use Markdown with YAML frontmatter:

```yaml
---
description: Short description shown in command picker
---

# Command body (prompt template)
...
```

**Available interpolations:**
- `$ARGUMENTS` — arguments passed after the command name

**Shell commands:** Use `!command` syntax to run shell commands inline.

opencode does **not** use `allowed-tools` (all tools are available by default) or `argument-hint` (argument passing is described in the body text).

## How It Works

When you invoke `/crit-code-review`, opencode sends the command's Markdown content as a prompt to the AI assistant. The AI then:

1. Guides you to run the interactive `crit review` TUI in your terminal
2. After you return and confirm, runs `crit status` via the `!command` syntax to read your review comments
3. Systematically edits files to address each comment
4. Summarizes all changes made

## Differences from Claude Code / Copilot CLI

| Feature | Claude Code / Copilot CLI | opencode |
|---------|--------------------------|----------|
| Format | `SKILL.md` in directories | Flat `.md` files |
| Install path | `~/.claude/skills/` or `~/.copilot/skills/` | `~/.config/opencode/commands/` |
| Tool restrictions | `allowed-tools: Bash(crit *)` | All tools available |
| tmux detection | Explicit tmux check + `--detach --wait` | User runs TUI manually |
| Shell execution | `Bash(crit *)` tool | `!command` syntax |

## Troubleshooting

**Command not found:** Ensure `~/.config/opencode/commands/` contains the `.md` files. Re-run `crit setup-opencode --force`.

**`crit` not found:** Install with `go install github.com/kevindutra/crit/cmd/crit@latest`.

**opencode not finding project commands:** Ensure you're running `opencode` from the project root (where `.opencode/commands/` is located).
