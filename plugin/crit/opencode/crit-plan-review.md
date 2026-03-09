---
description: Open a document or plan in crit's interactive TUI for review. After the review, address any comments by editing the document.
---

# Review Document: $ARGUMENTS

Review the document at `$ARGUMENTS` using crit's interactive TUI.

## Prerequisites

The `crit` binary must be installed and on PATH. If not installed:

```bash
go install github.com/kevindutra/crit/cmd/crit@latest
```

## Step 1: Ask the user to run the TUI

Ask the user to run the TUI in their terminal and let you know when they're done reviewing:

> Please run this in your terminal, review the document, then come back and tell me you're done:
>
> ```
> crit review $ARGUMENTS
> ```

Wait for the user to confirm before proceeding.

## Step 2: Read the comments

After the user confirms the review is complete, read the review comments:

```bash
!crit status $ARGUMENTS
```

This outputs JSON with the file path and comments array.

## Step 3: Address comments

For each comment in the `comments` array:

1. Read the `line` number and `content_snippet` to locate where in the document the comment applies
2. Read the `body` for what the reviewer wants changed
3. Edit the document at `$ARGUMENTS` to address the comment

After addressing ALL comments, summarize what you changed.

## Step 4: Re-review (optional)

After making changes, ask the user if they want to re-review:

> "I've addressed all comments. Want to review the changes? I'll open crit again."

If yes, go back to Step 1. If no, done.

## Important notes

- Do NOT modify the document while the TUI is open — only edit after it exits
- The `content_snippet` field shows the line content when the comment was created — use it to find the right location even if line numbers have shifted
