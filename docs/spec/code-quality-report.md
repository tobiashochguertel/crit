---
title: "Code Quality Report — crit-cli"
description: "SOLID/DRY violations, technical debt, and dependency analysis for the crit-cli package"
last_updated: "2025-01-20"
status: "analysis-complete"
---

# Code Quality Report — crit-cli

## Executive Summary

| Category | Finding Count | Critical | High | Medium | Low |
|---|---|---|---|---|---|
| SRP Violations | 4 | 1 | 2 | 1 | 0 |
| DRY Violations | 7 | 1 | 2 | 3 | 1 |
| DIP Violations | 2 | 0 | 2 | 0 | 0 |
| OCP Violations | 1 | 0 | 1 | 0 | 0 |
| Tech Debt | 6 | 0 | 1 | 3 | 2 |

**Primary concern:** `packages/crit-cli/internal/tui/app.go` is a **2,188-line God Object** that owns rendering, input handling, persistence, layout calculation, modal management, Markdown parsing, and scroll logic simultaneously. This file alone accounts for 80% of all violations.

**Positive highlights:** The CLI layer (`cli/`) is well-structured with data-driven agent registration. The runtime download system (`source.go`) is clean and single-purpose. The `filetab.go` type is a good example of isolated, focused design.

---

## 1. SRP Violations — Single Responsibility Principle

### 1.1 `app.go` — God Object (CRITICAL)

**File:** `packages/crit-cli/internal/tui/app.go` (~2,188 lines)

`AppModel` and `app.go` own at least **10 distinct responsibilities**:

| Responsibility | Where in file |
|---|---|
| TUI state management | struct fields, `Init()` |
| Window/layout geometry | `recalculateLayout()` line 932 |
| Input dispatch & key handling | `handleKeyPress()` line 249 |
| Content rendering | `rebuildContent()` line 1019 |
| Table block parsing & rendering | `detectTableBlocks()`, `formatTableRow()` lines 1376-1461 |
| Markdown/inline syntax highlighting | `highlightMarkdown()`, `highlightInline()` lines 1311-1499 |
| Comment sidebar state + rendering | `updateCommentSidebar()` lines 1655-1722 |
| Scroll position management | `scrollToCursor()`, `scrollToChunk()`, `scrollToAnnotation()` lines 1500-1652 |
| Modal lifecycle (open/submit/delete) | `modalSubmit()`, `modalDelete()`, `handleTextModal()` lines 623-729 |
| View composition | `View()`, `renderTabBar()`, `renderFooter()`, `renderWithModal()` lines 1724-2188 |

**Concrete examples:**

```go
// rebuildContent() alone is 240 lines (lines 1019-1258) and mixes:
// - document line iteration
// - diff annotation injection
// - Chroma syntax highlighting
// - Markdown block detection + rendering
// - table detection + rendering
// - line number gutter rendering
// - selection highlight rendering
// - annotation inline rendering
```

```go
// handleKeyPress() is 370 lines (lines 249-621) mixing:
// - tab switching navigation
// - cursor movement (j/k/up/down/page)
// - visual selection mode
// - comment navigation (next/prev annotation)
// - diff chunk navigation
// - modal open trigger
// - sidebar toggle
// - horizontal scroll
```

**Recommendation:** Split into 8–10 focused files (see Refactoring Instructions).

---

### 1.2 `paths.go` — Side Effect in Utility (HIGH)

**File:** `packages/crit-cli/internal/document/paths.go`

`EnsureDirs()` both computes path and **creates the `.crit/` directory + writes `.gitignore`**:

```go
// This function creates filesystem structure — not expected from a "paths" utility
func EnsureDirs(filePath string) (string, error) {
    dir := filepath.Join(filepath.Dir(filePath), ".crit")
    if err := os.MkdirAll(dir, 0755); err != nil { ... }
    gitignore := filepath.Join(dir, ".gitignore")
    // writes "*.yaml\n" to .gitignore
}
```

A `paths` package should only compute paths; filesystem operations should be in a `setup` or `store` layer.

---

### 1.3 `styles.go` — Mixed Concerns (HIGH)

**File:** `packages/crit-cli/internal/tui/styles.go`

The file mixes three distinct concerns:
1. **Style declarations** — 200+ `var` definitions
2. **Initialization logic** — `initAdaptiveStyles()` mutates globals at runtime
3. **Utility function** — `bgToAnsi()` string conversion

Additionally, `continuationGutter` is computed eagerly at package init via `.Render()`:

```go
var continuationGutter = gutterBaseStyle.Render(strings.Repeat(" ", gutterWidth))
```

This means styles are not recomputable after terminal background detection.

---

### 1.4 `store.go` — Migration Code Leakage (MEDIUM)

**File:** `packages/crit-cli/internal/review/store.go`

Legacy JSON→YAML migration logic lives permanently in `Load()`:

```go
// This will stay forever:
if err := json.Unmarshal(data, &state); err == nil {
    // migrate: re-save as YAML
    return &state, Save(&state)
}
```

The migration concern is mixed into the load path indefinitely.

---

## 2. DRY Violations — Don't Repeat Yourself

### 2.1 `tabIndexAtX()` duplicates `renderTabBar()` label building (CRITICAL)

**File:** `app.go`, lines 850-929 vs 1837-1998

`tabIndexAtX()` (80 lines) re-implements the tab label disambiguation logic and visible-window overflow calculation from `renderTabBar()` (160 lines). The key duplicated logic:

```go
// In tabIndexAtX() (lines 856-874):
basenames := make(map[string]int)
for _, t := range m.tabs { basenames[filepath.Base(t.path)]++ }
for i, t := range m.tabs {
    base := filepath.Base(t.path)
    text := base
    if basenames[base] > 1 {
        dir := filepath.Base(filepath.Dir(t.path))
        text = dir + "/" + base
    }
    ...
}

// In renderTabBar() (lines 1849-1872) — same logic, different variable names:
basenames := make(map[string]int)
for _, t := range m.tabs { basenames[filepath.Base(t.path)]++ }
for i, t := range m.tabs {
    label := filepath.Base(t.path)
    if basenames[label] > 1 { label = t.path }
    ...
}
```

The two implementations are **slightly inconsistent**: `tabIndexAtX` uses `dir + "/" + base` while `renderTabBar` uses the full `t.path`. This means click targets can misalign with displayed tab labels.

**Recommendation:** Extract `buildTabLabels()` returning `[]tabLabel`, shared by both functions. Fix the inconsistency as a bonus.

---

### 2.2 `NextComment` / `PrevComment` handler duplication (HIGH)

**File:** `app.go`, lines 416-487

Both handlers have identical structure — only the comparison operators differ (`>` vs `<`, `<` vs `>`):

```go
// NextComment (lines 416-451):
for _, c := range t.state.Comments {
    endAt := c.Line
    if c.EndLine > 0 { endAt = c.EndLine }
    if endAt > t.cursorLine || (...) {       // ← differs
        if best == nil || endAt < best.endLine { // ← differs
            ...
        }
    }
}
// Wrap-around loop: identical structure, same differs

// PrevComment (lines 452-487): identical structure, operators reversed
```

**Recommendation:** Extract `navigateToAdjacentComment(forward bool)` that parameterizes the comparison.

---

### 2.3 `NewApp()` / `NewCodeReviewApp()` textarea initialization (HIGH)

**File:** `app.go`, lines 82-143

Both constructors duplicate the same 4-line textarea setup:

```go
// NewApp() lines 83-86:
ta := textarea.New()
ta.Placeholder = "Type your comment..."
ta.ShowLineNumbers = false
ta.CharLimit = 2000

// NewCodeReviewApp() lines 106-109: identical
ta := textarea.New()
ta.Placeholder = "Type your comment..."
ta.ShowLineNumbers = false
ta.CharLimit = 2000
```

**Recommendation:** Extract `newCommentTextarea() textarea.Model`.

---

### 2.4 `annosByEndLine` map building duplicated (MEDIUM)

**File:** `app.go`, lines ~1053 and ~1608

`rebuildContent()` and `extraLinesPerDocLine()` both independently build a `map[int][]annotation` by iterating `t.state.Comments`:

```go
// In rebuildContent():
annosByEndLine := map[int][]annotation{}
for _, c := range t.state.Comments {
    endAt := c.Line
    if c.EndLine > 0 { endAt = c.EndLine }
    annosByEndLine[endAt] = append(...)
}

// In extraLinesPerDocLine(): same loop, same map
```

**Recommendation:** Extract `buildAnnotationIndex(state *review.ReviewState) map[int][]annotation`.

---

### 2.5 `inlineBg` closure duplicates `bgToAnsi()` (MEDIUM)

**File:** `app.go`, inside `rebuildContent()` at line ~1045

```go
// In rebuildContent() — an inline closure:
inlineBg := func(c lipgloss.Color) string {
    return "\033[48;2;" + bgToAnsiComponents(c) + "m"
}

// In styles.go — the already-existing utility:
func bgToAnsi(c lipgloss.Color) string { ... }
```

The closure was added inline instead of calling the existing `bgToAnsi()` function.

**Recommendation:** Delete the closure, call `bgToAnsi()` directly.

---

### 2.6 Scroll helper duplication (MEDIUM)

**File:** `app.go`, lines 1500-1652

`scrollToCursor()`, `scrollToChunk()`, and `scrollToAnnotation()` each re-traverse document lines and call `extraLinesPerDocLine()`:

```go
// scrollToCursor (lines ~1500-1545):
offset := 0
for i := 1; i < t.cursorLine; i++ {
    offset += 1 + m.extraLinesPerDocLine(i) // same pattern
}
m.contentViewport.SetYOffset(...)

// scrollToAnnotation (lines ~1580-1610): same loop pattern
```

**Recommendation:** Extract `lineToViewportOffset(lineNum int) int` helper.

---

### 2.7 Edit modal context lookup duplicated (LOW)

**File:** `app.go`, `renderWithModal()` lines 2112-2127

`renderWithModal()` searches `t.state.Comments` by `m.editingID` to get context line numbers — the same lookup is done in `modalSubmit()`. Minor duplication, low impact.

---

## 3. DIP Violations — Dependency Inversion Principle

### 3.1 TUI directly calls `review.Save()` (HIGH)

**File:** `app.go`, lines 658 and 673

```go
// modalSubmit() — TUI layer calls persistence directly:
review.Save(t.state)

// modalDelete() — same:
review.Save(t.state)
```

The TUI layer (`tui` package) directly imports and calls `review.Save()`. This couples UI to storage. If storage format changes or we want to test the TUI in isolation, all tests must handle real file I/O.

**Recommendation:** Add a `SaveFn func(*review.ReviewState) error` field to `AppModel`. Default it to `review.Save` in constructors. Tests can inject a no-op.

---

### 3.2 `session.go` imports `document.EnsureDirs()` (HIGH)

**File:** `packages/crit-cli/internal/review/session.go`

The `review` package imports `document` to call `EnsureDirs()`. This creates a cross-package dependency from storage to document utilities that should flow the other way.

**Recommendation:** Pass a `basePath string` into `session.go` functions instead of calling `document.EnsureDirs()` internally.

---

## 4. OCP Violations — Open/Closed Principle

### 4.1 `renderWithModal()` switch on modal type (HIGH)

**File:** `app.go`, lines 2092-2139

```go
switch m.modal {
case commentModal:
    // build comment modal UI
case editModal:
    // build edit modal UI
}
```

Adding a new modal type requires modifying `renderWithModal()`, `handleTextModal()`, and the `modalType` constants — three separate changes for one new concept.

**Recommendation:** Define a `Modal` interface with `Render(innerWidth int) string` and `HandleKey(msg) bool`. Each modal type becomes a struct implementing the interface. `AppModel.modal` becomes a `Modal` (or nil).

---

## 5. Technical Debt

### 5.1 `highlight.go` hardcodes Chroma style (HIGH)

**File:** `packages/crit-cli/internal/tui/highlight.go`

```go
style := styles.Get("monokai")    // hardcoded
formatter := formatters.Get("terminal256") // hardcoded
```

The syntax highlighting style is not configurable. Users with light-background terminals get unreadable output.

**Recommendation:** Allow overriding via `CRIT_CHROMA_STYLE` and `CRIT_CHROMA_FORMATTER` env vars, falling back to adaptive selection based on the `tea.BackgroundColorMsg`.

---

### 5.2 `git/diff.go` — external `git` binary not abstracted (MEDIUM)

**File:** `packages/crit-cli/internal/git/diff.go`

```go
func gitCommand(args ...string) ([]byte, error) {
    cmd := exec.Command("git", args...)
    // ...
}
```

All git operations run the `git` binary directly with no abstraction. This makes unit testing impossible without a real git repo. `IsGitRepo()` also discards error information:

```go
func IsGitRepo() bool {
    _, err := gitCommand("rev-parse", "--git-dir")
    return err == nil  // error context lost
}
```

**Recommendation:** Extract a `GitRunner` interface with `Run(args ...string) ([]byte, error)`. Tests inject a fake.

---

### 5.3 `review.go` CLI — package-level flag vars (MEDIUM)

**File:** `packages/crit-cli/internal/cli/review.go`

```go
var reviewDetach bool
var reviewWait   bool
var reviewCode   bool
```

Package-level mutable variables are not safe for concurrent use and make the CLI hard to test.

**Recommendation:** Bind flags to local variables within the command constructor, or use a config struct.

---

### 5.4 `store.go` — permanent migration code (MEDIUM)

**File:** `packages/crit-cli/internal/review/store.go`

Legacy JSON→YAML migration code will remain in the load path indefinitely:

```go
// Legacy JSON path — try JSON first, fall back to YAML
if err := json.Unmarshal(data, &state); err == nil {
    return &state, Save(&state)  // re-save as YAML
}
```

**Recommendation:** Ship a `crit migrate` command that converts all `.crit/*.yaml` files in one pass, then remove the migration code in the next major version.

---

### 5.5 Missing test coverage (LOW)

No test files exist anywhere in `packages/crit-cli/`. The entire package has 0% automated test coverage. The architecture makes testing hard in some places (god-object TUI, package-level flag vars, direct `git` calls).

**Recommendation:** Start with pure-function units — `annotationsAfterLine()`, `buildTabLabels()`, `navigateToAdjacentComment()`, `buildAnnotationIndex()` — as these have no external dependencies and are easy to test once extracted.

---

### 5.6 `clampLines()` defined at package level in `app.go` (LOW)

**File:** `app.go`, line 2181

A general-purpose string utility (`clampLines`) is defined as a package-level function inside the largest file. Low severity, but should be co-located with the markdown/text utilities in `highlight.go`.

---

## 6. Overall Dependency Graph

```
CLI layer (cmd/crit, cli/)
    ↓
review/ ←→ document/  ← coupling (session imports document)
    ↓
tui/  ← app.go directly imports review/ and calls review.Save()
    ↓
git/  ← no abstraction over exec.Command("git")
```

The clean direction should be:
```
cmd/ → cli/ → tui/ → review/ → document/ → git/
                            ↘ git/
```

---

## 7. Files in Good Shape (No Action Needed)

| File | Why it's clean |
|---|---|
| `filetab.go` | Focused single type, all fields well-scoped |
| `messages.go` | Thin DTO file, no logic |
| `keys.go` | Simple key binding declarations |
| `cli/agent_config.go` | Data-driven, adding agents requires one map entry |
| `cli/source.go` | Single-purpose, clear priority chain |
| `review/store.go` | Atomic write + file locking, correct |
| `review/types.go` | Clean value objects |
| `document/document.go` | Well-scoped document representation |

---

*Report generated by code analysis — 2025-01-20*
