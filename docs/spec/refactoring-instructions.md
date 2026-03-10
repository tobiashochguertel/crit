---
title: "Refactoring Instructions — crit-cli TUI"
description: "Actionable AI-agent coding prompt with acceptance criteria for fixing SOLID/DRY violations"
last_updated: "2025-01-20"
relates_to: "code-quality-report.md"
---

# Refactoring Instructions — crit-cli TUI

> **Purpose:** This document is an instruction prompt for an AI coding agent (or human developer). Execute
> each phase in order. Do not jump ahead — later phases depend on earlier splits being in place.
> Acceptance criteria are provided per task to define "done".

## Prerequisites

- Working directory: `packages/crit-cli/`
- All tests must pass (or be noted as newly added) before each commit
- Build must pass: `cd packages/crit-cli && go build ./...`
- Run `go vet ./...` after each phase; zero warnings required

---

## Phase 1 — Split `app.go` God Object (Highest Impact)

> **Why first:** All other DRY/SOLID fixes require the code to be split first. The single 2,188-line
> file makes diffs, code review, and navigation unusable.

### Task 1.1 — Extract `model.go`

Create `packages/crit-cli/internal/tui/model.go` containing:

- All `const`/`type` declarations currently at the top of `app.go` (lines 1–75):
  - `pane` type + constants (`contentPane`, `commentPane`)
  - `modalType` type + constants (`noModal`, `commentModal`, `editModal`)
  - `gutterWidth` constant
  - `AppModel` struct
  - `sidebarItem` struct (currently defined around line 1004)
  - `annotation` struct (currently defined around line 1007)
- `NewApp()` constructor
- `NewCodeReviewApp()` constructor
- `selectionRange()` method

**Acceptance Criteria:**
- [ ] `model.go` exists and `go build ./...` succeeds
- [ ] `app.go` no longer contains struct definitions or constructors
- [ ] `AppModel` is defined exactly once (in `model.go`)

---

### Task 1.2 — Extract `view.go`

Create `packages/crit-cli/internal/tui/view.go` containing:

- `View()` method (currently ~lines 1724–1835)
- `renderTabBar()` method (currently ~lines 1837–1998)
- `renderFooter()` method (currently ~lines 2000–2033)
- `renderModalButton()` function (currently ~lines 2035–2045)
- `renderDeleteButton()` function (currently ~lines 2047–2053)
- `renderContextPreview()` method (currently ~lines 2052–2079)
- `renderWithModal()` method (currently ~lines 2081–2163)
- `dimRendered()` method (currently ~lines 2165–2179)

**Acceptance Criteria:**
- [ ] `view.go` exists; `app.go` no longer contains `View()` or any `render*` function
- [ ] `go build ./...` passes
- [ ] Running `crit` shows identical output to before the split

---

### Task 1.3 — Extract `content.go`

Create `packages/crit-cli/internal/tui/content.go` containing:

- `rebuildContent()` method (~lines 1019–1258)
- `renderAnnotationBox()` method (~lines 1260–1294)
- `extraLinesPerDocLine()` method (~lines 1608–1650)
- `annotationsAfterLine()` method (~lines 982–1000)
- `updateCommentSidebar()` method (~lines 1655–1722)
- `scrollToCursor()` method (~lines 1500–1545)
- `scrollToChunk()` method (~lines 1547–1580)
- `scrollToAnnotation()` method (~lines 1580–1610)

**Acceptance Criteria:**
- [ ] `content.go` exists; all listed methods are gone from `app.go`
- [ ] `go build ./...` passes
- [ ] Comment annotations render correctly in TUI

---

### Task 1.4 — Extend `highlight.go`

Move to `packages/crit-cli/internal/tui/highlight.go`:

- Markdown regex vars (`reBold`, `reItalic`, `reCode`, `reStrike`, `reLinkText`) (~lines 1296–1308)
- `highlightMarkdown()` function (~lines 1310–1370)
- `highlightInline()` function (~lines 1371–1375)
- `tableBlock` struct (~lines 1376–1380)
- `detectTableBlocks()` function (~lines 1381–1410)
- `parseTableCells()` function (~lines 1411–1435)
- `formatTableRow()` function (~lines 1436–1460)
- `formatTableSep()` function (~lines 1461–1499)
- `clampLines()` function (line ~2181)

**Acceptance Criteria:**
- [ ] All markdown/table helpers are in `highlight.go`; none remain in `app.go`
- [ ] `go build ./...` passes

---

### Task 1.5 — Extract `layout.go`

Create `packages/crit-cli/internal/tui/layout.go` containing:

- `recalculateLayout()` method (~lines 932–978)

**Acceptance Criteria:**
- [ ] `layout.go` exists; `recalculateLayout()` is gone from `app.go`
- [ ] `go build ./...` passes

---

### Task 1.6 — Extract `mouse.go`

Create `packages/crit-cli/internal/tui/mouse.go` containing:

- `handleMouseWheel()` method (~lines 784–813)
- `handleMouseClick()` method (~lines 815–844)
- `tabIndexAtX()` method (~lines 850–930)

**Acceptance Criteria:**
- [ ] `mouse.go` exists; listed methods are gone from `app.go`
- [ ] `go build ./...` passes
- [ ] Mouse click on tabs still switches tabs correctly

---

### Task 1.7 — Extract `modal.go`

Create `packages/crit-cli/internal/tui/modal.go` containing:

- `modalSubmit()` method (~lines 623–667)
- `modalDelete()` method (~lines 669–680)
- `handleTextModal()` method (~lines 682–729)

**Acceptance Criteria:**
- [ ] `modal.go` exists; listed methods are gone from `app.go`
- [ ] `go build ./...` passes
- [ ] Creating and deleting comments still works correctly

---

### Task 1.8 — Verify Phase 1 Complete

After all 7 split tasks, `app.go` should contain only:

- `Init()`, `loadDocuments()`
- `Update()` dispatcher
- `handleKeyPress()` (370 lines — leave for now; Phase 3 handles this)
- `handleTabSearch()` and `updateTabSearchMatches()`

**Acceptance Criteria:**
- [ ] `app.go` is ≤ 650 lines
- [ ] `go build ./...` passes with no warnings
- [ ] `go vet ./...` outputs nothing
- [ ] Commit message: `refactor(tui): split app.go god object into focused files`

---

## Phase 2 — Fix DRY Violations

### Task 2.1 — Deduplicate tab label building

The functions `tabIndexAtX()` (in `mouse.go` after Phase 1) and `renderTabBar()` (in `view.go` after Phase 1) both build basename→deduplication maps and compute a visible window.

**Action:** In `model.go` (or a new `tabs.go`), create a shared helper:

```go
type tabLabel struct {
    text  string
    width int  // measured with inactiveTabStyle (smaller of the two styles)
}

// buildTabLabels returns the display label for each tab, de-duplicating basenames.
func buildTabLabels(tabs []FileTab) []tabLabel { ... }
```

Then update both `tabIndexAtX()` and `renderTabBar()` to call `buildTabLabels()` instead of duplicating the loop.

Also **fix the inconsistency**: `tabIndexAtX` currently uses `dir + "/" + base` while `renderTabBar` uses `t.path`. Both should use `dir + "/" + base`.

**Acceptance Criteria:**
- [ ] The basename-deduplication loop exists in exactly one place
- [ ] `tabIndexAtX` and `renderTabBar` call `buildTabLabels()`
- [ ] Both functions produce identical labels for the same tab list
- [ ] Clicking a tab with a deduplicated name navigates to the correct file

---

### Task 2.2 — Deduplicate comment navigation

In `handleKeyPress()` (or in `modal.go`/`content.go` after Phase 1), the `NextComment` and `PrevComment` cases are near-identical.

**Action:** Extract to `content.go`:

```go
// navigateToAdjacentComment finds the nearest comment after (forward=true) or
// before (forward=false) the current cursor line and sets cursorLine accordingly.
func (m *AppModel) navigateToAdjacentComment(forward bool) {
    t := m.tab()
    if t.state == nil { return }
    type candidate struct { endLine int; startLine int }
    var best *candidate
    for _, c := range t.state.Comments {
        endAt := c.Line
        if c.EndLine > 0 { endAt = c.EndLine }
        past := (forward && endAt > t.cursorLine) || (!forward && endAt < t.cursorLine)
        if past {
            if best == nil ||
                (forward && endAt < best.endLine) ||
                (!forward && endAt > best.endLine) {
                best = &candidate{endLine: endAt, startLine: c.Line}
            }
        }
    }
    // wrap-around
    if best == nil {
        for _, c := range t.state.Comments {
            endAt := c.Line
            if c.EndLine > 0 { endAt = c.EndLine }
            if best == nil ||
                (forward && endAt < best.endLine) ||
                (!forward && endAt > best.endLine) {
                best = &candidate{endLine: endAt, startLine: c.Line}
            }
        }
    }
    if best != nil {
        t.cursorLine = best.startLine
        m.scrollToCursor()
        m.updateCommentSidebar()
        m.tabs[m.activeTab] = *t
    }
}
```

Then replace both `case m.keys.NextComment` and `case m.keys.PrevComment` blocks in `handleKeyPress()` with single-line calls:

```go
case m.keys.NextComment:
    m.navigateToAdjacentComment(true)
case m.keys.PrevComment:
    m.navigateToAdjacentComment(false)
```

**Acceptance Criteria:**
- [x] Comment navigation logic exists in exactly one function
- [x] `NextComment` and `PrevComment` are each one line in `handleKeyPress()`
- [x] Navigate-forward and navigate-backward both still work, including wrap-around

---

### Task 2.3 — Extract `newCommentTextarea()` factory

**Action:** In `model.go`, add:

```go
// newCommentTextarea returns a configured textarea for comment entry.
func newCommentTextarea() textarea.Model {
    ta := textarea.New()
    ta.Placeholder = "Type your comment..."
    ta.ShowLineNumbers = false
    ta.CharLimit = 2000
    return ta
}
```

Replace the identical 4-line blocks in both `NewApp()` and `NewCodeReviewApp()`.

**Acceptance Criteria:**
- [x] The textarea initialization block exists in exactly one place (`newCommentTextarea`)
- [x] Both constructors call `newCommentTextarea()`
- [x] `go build ./...` passes

---

### Task 2.4 — Extract `buildAnnotationIndex()` helper

**Action:** In `content.go`, add:

```go
// buildAnnotationIndex maps each end-line number to the annotations that display after it.
func buildAnnotationIndex(state *review.ReviewState) map[int][]annotation {
    m := make(map[int][]annotation)
    if state == nil { return m }
    for _, c := range state.Comments {
        endAt := c.Line
        if c.EndLine > 0 { endAt = c.EndLine }
        m[endAt] = append(m[endAt], annotation{
            id: c.ID, body: c.Body,
            line: c.Line, endLine: c.EndLine,
        })
    }
    return m
}
```

Replace the duplicated `annosByEndLine` map-building loops in `rebuildContent()` and `extraLinesPerDocLine()`.

**Acceptance Criteria:**
- [ ] The `annosByEndLine` loop exists in exactly one place
- [ ] `rebuildContent()` and `extraLinesPerDocLine()` both call `buildAnnotationIndex()`
- [ ] Comment annotation rendering is identical to before

---

### Task 2.5 — Remove `inlineBg` closure in favour of `bgToAnsi()`

**Action:** In `rebuildContent()`, find the local `inlineBg` closure and replace its call sites with direct calls to the already-existing `bgToAnsi()` function in `styles.go`. Delete the closure.

**Acceptance Criteria:**
- [ ] The `inlineBg` local closure is gone from `rebuildContent()`
- [ ] `bgToAnsi()` is the single implementation of that logic
- [ ] Syntax-highlighted content renders identically

---

### Task 2.6 — Commit Phase 2

**Acceptance Criteria:**
- [ ] `go build ./...` and `go vet ./...` pass cleanly
- [ ] Commit message: `refactor(tui): eliminate DRY violations — shared helpers for tabs, comments, annotations`

---

## Phase 3 — Fix Dependency Inversion (DIP)

### Task 3.1 — Decouple `modalSubmit()`/`modalDelete()` from `review.Save()`

**Action:** Add a save callback field to `AppModel`:

```go
// In model.go — AppModel struct:
// SaveFn is called whenever review state needs to be persisted.
// Defaults to review.Save. Override in tests with a no-op.
SaveFn func(*review.ReviewState) error
```

In `NewApp()` and `NewCodeReviewApp()`, initialize:

```go
SaveFn: review.Save,
```

In `modal.go`, replace:

```go
review.Save(t.state)
```

with:

```go
if m.SaveFn != nil {
    m.SaveFn(t.state)
}
```

**Acceptance Criteria:**
- [x] `tui` package no longer calls `review.Save()` directly
- [x] `SaveFn` is the only persistence call in the TUI
- [x] Default behavior (saving reviews) is unchanged
- [x] A test can construct `AppModel` with `SaveFn: func(*review.ReviewState) error { return nil }` without any filesystem I/O

---

### Task 3.2 — Commit Phase 3

**Acceptance Criteria:**
- [x] `go build ./...` and `go vet ./...` pass
- [x] Commit message: `refactor(tui): decouple persistence via SaveFn callback`

---

## Phase 4 — `styles.go` Thread-Safety Cleanup (Optional / Lower Priority)

> This phase addresses the mutable globals in `styles.go`. It is lower-priority than Phases 1–3
> but recommended before adding any concurrent code.

### Task 4.1 — Replace `continuationGutter` eager init

**Action:** Change `continuationGutter` from a package-level `var` to a function:

```go
// continuationGutter returns the rendered empty gutter for soft-wrap continuation lines.
// Computed on-demand so it respects any style re-initialization.
func continuationGutter() string {
    return gutterBaseStyle.Render(strings.Repeat(" ", gutterWidth))
}
```

Update all call sites from `continuationGutter` to `continuationGutter()`.

**Acceptance Criteria:**
- [ ] `continuationGutter` is a function, not a `var`
- [ ] All call sites updated
- [ ] `go build ./...` passes

---

### Task 4.2 — Move `bgToAnsi()` to a utility file

**Action:** Move `bgToAnsi()` (and any private helpers it calls) from `styles.go` to either `highlight.go` or a new `color_utils.go`.

**Acceptance Criteria:**
- [ ] `styles.go` contains only style variable declarations and `initAdaptiveStyles()`
- [ ] `bgToAnsi()` exists in exactly one place
- [ ] `go build ./...` passes

---

## Phase 5 — Configurable Chroma Theme

### Task 5.1 — Allow overriding syntax highlight style

**File:** `highlight.go`

**Action:**

```go
// chromaStyle returns the Chroma style name to use.
// Respects CRIT_CHROMA_STYLE env var; defaults to "monokai".
func chromaStyle() string {
    if s := os.Getenv("CRIT_CHROMA_STYLE"); s != "" {
        return s
    }
    return "monokai"
}

// chromaFormatter returns the Chroma formatter name.
// Respects CRIT_CHROMA_FORMATTER env var; defaults to "terminal256".
func chromaFormatter() string {
    if s := os.Getenv("CRIT_CHROMA_FORMATTER"); s != "" {
        return s
    }
    return "terminal256"
}
```

Update `HighlightCode()` to call these functions.

**Acceptance Criteria:**
- [ ] `CRIT_CHROMA_STYLE=github` produces light-theme highlighting
- [ ] Default behavior unchanged when env vars are not set
- [ ] `go build ./...` passes

---

## Summary Checklist

| Phase | Task | Priority | Done |
|---|---|---|---|
| 1 | Split `app.go` → `model.go` | Critical | [ ] |
| 1 | Split `app.go` → `view.go` | Critical | [ ] |
| 1 | Split `app.go` → `content.go` | Critical | [ ] |
| 1 | Move markdown helpers to `highlight.go` | Critical | [ ] |
| 1 | Split `app.go` → `layout.go` | High | [ ] |
| 1 | Split `app.go` → `mouse.go` | High | [ ] |
| 1 | Split `app.go` → `modal.go` | High | [ ] |
| 2 | Shared `buildTabLabels()` | High | [ ] |
| 2 | `navigateToAdjacentComment()` | High | [ ] |
| 2 | `newCommentTextarea()` factory | Medium | [ ] |
| 2 | `buildAnnotationIndex()` helper | Medium | [ ] |
| 2 | Remove `inlineBg` closure | Medium | [ ] |
| 3 | `SaveFn` callback for persistence | High | [ ] |
| 4 | `continuationGutter()` function | Low | [ ] |
| 4 | Move `bgToAnsi()` | Low | [ ] |
| 5 | Configurable Chroma theme | Low | [ ] |

---

*Instructions authored 2025-01-20 based on `code-quality-report.md` findings.*
