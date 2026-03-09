---
description: Review code changes in crit's multi-file TUI with syntax highlighting and diff markers. After the review, address any comments.
---

# Code Review

Review code changes using crit's multi-file code review TUI.

## Prerequisites

The `crit` binary must be installed and on PATH. If not installed:

```bash
go install github.com/kevindutra/crit/cmd/crit@latest
```

## Step 1: Ask the user to run the TUI

Ask the user to run the TUI in their terminal and let you know when they're done reviewing:

> Please run this in your terminal, review the changes, then come back and tell me you're done:
>
> ```
> crit review --code
> ```

Wait for the user to confirm before proceeding.

## Step 2: Read the comments

After the user confirms the review is complete, read the aggregate review comments:

```bash
!crit status --code
```

This outputs JSON with all files and their comments.

## Step 3: Address comments

For each file in the `files` array, for each comment:

1. Read the `line` number and `content_snippet` to locate where the comment applies
2. Read the `body` for what the reviewer wants changed
3. Edit the file to address the comment

After addressing ALL comments across ALL files, summarize what you changed.

## Step 4: Re-review (optional)

After making changes, ask the user if they want to re-review:

> "I've addressed all comments. Want to review the changes? I'll open crit again."

If yes, go back to Step 1. If no, done.

## Important notes

- Do NOT modify files while the TUI is open — only edit after it exits
- The `content_snippet` field shows the line content when the comment was created — use it to find the right location even if line numbers have shifted
