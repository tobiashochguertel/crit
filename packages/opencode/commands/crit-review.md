---
description: Open crit for review. Routes to code review (multi-file TUI for code changes) or plan/document review (single-file TUI).
---

# Review

Ask the user what they want to review:

> What are you looking to review?
>
> 1. **Code changes** — Review changed files in a tabbed TUI with syntax highlighting and diff markers
> 2. **A document or plan** — Review a specific file in the interactive TUI

If the user chooses **code review**, trigger the `/crit-code-review` command.

If the user chooses **document/plan review**, ask for the file path and trigger the `/crit-plan-review` command with that path.
