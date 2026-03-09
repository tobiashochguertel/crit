# AI Agent Plugin Support

`crit` integrates with multiple AI coding assistants as a plugin/skill, enabling the TUI code review workflow directly from your AI chat session.

## Supported AI Agents

| Agent | Installation | Plugin Type |
|-------|-------------|-------------|
| [Claude Code](https://claude.ai/code) | `crit setup-claude` | Skills (`~/.claude/skills/`) |
| [GitHub Copilot CLI](https://github.com/github/copilot-cli) | `crit setup-copilot` | Skills (`~/.copilot/skills/`) |
| [opencode](https://opencode.ai/) | `crit setup-opencode` | Commands (`~/.config/opencode/commands/`) |

## Quick Start

```bash
# Install crit
go install github.com/tobiashochguertel/crit/cmd/crit@latest

# Set up for your preferred AI agent
crit setup-claude      # Claude Code (global)
crit setup-copilot     # GitHub Copilot CLI (global)
crit setup-opencode    # opencode (global)

# Project-scoped installation (only affects this repo)
crit setup-claude --project
crit setup-copilot --project
crit setup-opencode --project
```

## Available Commands / Skills

All three agents get the same review workflow:

| Command / Skill | Description |
|----------------|-------------|
| `crit-review` | Router: asks what to review, dispatches to code or plan review |
| `crit-code-review` | Multi-file code change review with diff markers |
| `crit-plan-review` | Single-file document / plan review |

## Detailed Guides

- [Claude Code integration](./claude-code.md)
- [GitHub Copilot CLI integration](./copilot-cli.md)
- [opencode integration](./opencode.md)

## Plugin Marketplace

crit is available in the Claude Code Plugin Marketplace. To install via the marketplace instead of the CLI:

```
/plugin install https://github.com/tobiashochguertel/crit
```

or using `tobiashochguertel`'s fork (includes copilot/opencode support):

```
/plugin install https://github.com/tobiashochguertel/crit
```
