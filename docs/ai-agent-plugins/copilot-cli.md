# GitHub Copilot CLI Integration

`crit` integrates with [GitHub Copilot CLI](https://github.com/github/copilot-cli) (`ghcs` / `gh copilot suggest`) via the Copilot CLI skill system.

## Installation

### Global (recommended)

```bash
crit setup-copilot
```

Installs skills to `~/.copilot/skills/`:
- `~/.copilot/skills/crit-review/SKILL.md`
- `~/.copilot/skills/crit-code-review/SKILL.md`
- `~/.copilot/skills/crit-plan-review/SKILL.md`

### Project-scoped

```bash
crit setup-copilot --project
```

Installs to `.copilot/skills/` in the current directory. Skills are only available when working in this project.

### Overwrite existing files

```bash
crit setup-copilot --force
```

## Usage

After installation, mention the skill in your Copilot CLI conversation:

```bash
# Review code changes
ghcs "use the crit-code-review skill to review my changes"

# Review a document
ghcs "use the crit-plan-review skill on docs/plan.md"

# Auto-route
ghcs "use the crit-review skill"
```

## Skills

### `crit-review`

Routes between code review and document/plan review based on user input.

### `crit-code-review`

Reviews code changes using `crit review --code`. Reads comments via `crit status --code` and addresses them.

**Requirements:** `crit` binary on PATH

### `crit-plan-review`

Reviews a specific document using `crit review <file>`. Reads comments via `crit status <file>` and addresses them.

**Usage:** Provide the file path as an argument, e.g.:
```
use crit-plan-review on docs/PLAN.md
```

## Plugin Marketplace

The plugin can also be installed via the Claude Code Plugin Marketplace — Copilot CLI reads compatible plugin directories:

```
/plugin install https://github.com/tobiashochguertel/crit
```

The marketplace plugin includes pre-built skills in `plugin/crit/skills/`.

## How It Works

Copilot CLI reads `.md` files from `~/.copilot/skills/<skill-name>/SKILL.md` (global) or `.copilot/skills/<skill-name>/SKILL.md` (project) and makes them available as named skills in conversations.

The skill format uses YAML frontmatter:
- `name`: skill identifier
- `description`: used for routing/discovery
- `allowed-tools`: tools the skill can use (e.g., `Bash(crit *)`)
- `argument-hint`: hint shown when invoking with arguments

The `crit-code-review` and `crit-plan-review` skills work differently from typical skills: they guide the AI to ask the user to run the TUI in their terminal (since `crit review` is interactive), then read the resulting JSON output to address comments programmatically.

## Troubleshooting

**Skill not found:** Ensure `~/.copilot/skills/` contains the `SKILL.md` files. Re-run `crit setup-copilot --force`.

**`crit` not found:** Install with `go install github.com/tobiashochguertel/crit/cmd/crit@latest` or via your package manager.

**tmux integration:** `crit review --code --detach --wait` requires tmux. If you're not in a tmux session, the skill will fall back to asking you to run the command manually.
