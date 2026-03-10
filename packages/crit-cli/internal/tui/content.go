package tui

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
)

// sidebarItem represents a comment in the sidebar list.
type sidebarItem struct {
	id      string
	line    int
	endLine int
	body    string
}

// annotation represents an inline comment to render.
type annotation struct {
	id      string
	body    string
	line    int
	endLine int
}

// annotationsAfterLine returns annotations that render after the given line
// (keyed by their endLine).
func (m *AppModel) annotationsAfterLine(lineNum int) []annotation {
	t := m.tab()
	if t.state == nil {
		return nil
	}
	var anns []annotation
	for _, c := range t.state.Comments {
		endAt := c.Line
		if c.EndLine > 0 {
			endAt = c.EndLine
		}
		if endAt == lineNum {
			anns = append(anns, annotation{
				id: c.ID, body: c.Body,
				line: c.Line, endLine: c.EndLine,
			})
		}
	}
	return anns
}

// rebuildContent renders the document line-by-line with cursor, selection,
// line numbers, and bordered inline annotations.
func (m *AppModel) rebuildContent() {
	t := m.tab()

	// Handle placeholder tabs
	if t.isBinary {
		m.contentViewport.SetContent("\n  Binary file changed — cannot display content.\n")
		return
	}
	if t.isDeleted {
		m.contentViewport.SetContent("\n  File deleted.\n")
		return
	}

	if t.doc == nil {
		return
	}

	// Collect annotations keyed by the line they appear AFTER
	annosByEndLine := make(map[int][]annotation)
	if t.state != nil {
		for _, c := range t.state.Comments {
			endAt := c.Line
			if c.EndLine > 0 {
				endAt = c.EndLine
			}
			annosByEndLine[endAt] = append(annosByEndLine[endAt], annotation{
				id: c.ID, body: c.Body,
				line: c.Line, endLine: c.EndLine,
			})
		}
	}

	// Count how many comments cover each line (for overlap detection)
	annotatedLines := make(map[int]int)
	if t.state != nil {
		for _, c := range t.state.Comments {
			end := c.Line
			if c.EndLine > 0 {
				end = c.EndLine
			}
			for l := c.Line; l <= end; l++ {
				annotatedLines[l]++
			}
		}
	}

	selStart, selEnd := m.selectionRange()

	// Determine which lines to highlight from selected annotation
	var sidebarHighlightStart, sidebarHighlightEnd int
	if m.focused == commentPane && len(t.sidebarItems) > 0 && t.sidebarCursor < len(t.sidebarItems) {
		sel := t.sidebarItems[t.sidebarCursor]
		sidebarHighlightStart = sel.line
		sidebarHighlightEnd = sel.line
		if sel.endLine > 0 {
			sidebarHighlightEnd = sel.endLine
		}
	} else if m.focused == contentPane && t.cursorOnAnnotation {
		anns := m.annotationsAfterLine(t.cursorLine)
		if t.cursorAnnoIdx < len(anns) {
			ann := anns[t.cursorAnnoIdx]
			sidebarHighlightStart = ann.line
			sidebarHighlightEnd = ann.line
			if ann.endLine > 0 {
				sidebarHighlightEnd = ann.endLine
			}
		}
	}

	contentWidth := m.contentViewport.Width()
	boxWidth := contentWidth - 7
	if boxWidth < 20 {
		boxWidth = 20
	}

	textWidth := contentWidth - 8
	if textWidth < 10 {
		textWidth = 10
	}

	// Use cached syntax highlighting
	isMarkdown := t.isMarkdown
	chromaLines := t.chromaLines

	// Detect table blocks so we can align columns across rows
	tableBlocks := detectTableBlocks(t.doc.Lines)
	tableBlockMap := make(map[int]*tableBlock)
	for i := range tableBlocks {
		tb := &tableBlocks[i]
		for l := tb.startLine; l <= tb.endLine; l++ {
			tableBlockMap[l] = tb
		}
	}

	inlineBg := renderInlineBg

	var b strings.Builder
	b.Grow(len(t.doc.Lines) * 200) // pre-allocate to reduce allocations
	for i, line := range t.doc.Lines {
		lineNum := i + 1

		// Render deleted lines that appear before this line
		if dels, ok := t.deletedAfter[lineNum-1]; ok {
			cachedHL := t.deletedLineCache[lineNum-1]
			for di, del := range dels {
				delMarker := diffDeletedGutter.Render("-")
				delNum := diffDeletedLineNum.Render(fmt.Sprintf("%d", del.OldLineNum))
				delContent := del.Content
				if cachedHL != nil && di < len(cachedHL) {
					delContent = cachedHL[di]
				}
				b.WriteString(fmt.Sprintf("%s%s %s\n", delMarker, delNum, inlineBg(diffDeletedLineBg, delContent)))
			}
		}

		isCursor := lineNum == t.cursorLine
		isSelected := t.selecting && lineNum >= selStart && lineNum <= selEnd
		isSidebarHighlight := sidebarHighlightStart > 0 && lineNum >= sidebarHighlightStart && lineNum <= sidebarHighlightEnd
		isChanged := t.changedLines != nil && t.changedLines[lineNum]

		// Marker column
		var marker string
		if isCursor && !t.cursorOnAnnotation {
			marker = cursorMarker.Render(">")
		} else if isSelected {
			marker = selectedMarker.Render("|")
		} else if isSidebarHighlight {
			marker = cursorMarker.Render(">")
		} else if count, ok := annotatedLines[lineNum]; ok && count > 0 {
			if count > 1 {
				marker = gutterOverlap.Render("◆")
			} else {
				marker = annotationGutter.Render("■")
			}
		} else if isChanged {
			marker = diffAddedGutter.Render("+")
		} else {
			marker = " "
		}

		// Line number
		var numStr string
		if isCursor {
			numStr = cursorLineNumStyle.Render(fmt.Sprintf("%d", lineNum))
		} else if isSelected {
			numStr = selectedLineNumStyle.Render(fmt.Sprintf("%d", lineNum))
		} else {
			numStr = lineNumStyle.Render(fmt.Sprintf("%d", lineNum))
		}

		// Check if this line is part of a table block
		if tb, inTable := tableBlockMap[lineNum]; inTable {
			var styledLine string
			if reTableSep.MatchString(line) {
				styledLine = formatTableSep(tb.colWidths)
			} else {
				isHeader := lineNum == tb.startLine
				styledLine = formatTableRow(line, tb.colWidths, isHeader)
			}

			if isSelected {
				styledLine = inlineBg(selectedLineBg, styledLine)
			} else if isSidebarHighlight {
				styledLine = inlineBg(sidebarHighlightBg, styledLine)
			} else if isChanged {
				styledLine = inlineBg(diffChangedLineBg, styledLine)
			}

			b.WriteString(fmt.Sprintf("%s%s %s\n", marker, numStr, styledLine))
		} else {
			// Get the display content: Chroma-highlighted or raw
			displayLine := line
			if !isMarkdown && chromaLines != nil && i < len(chromaLines) {
				displayLine = chromaLines[i]
			}

			// For Chroma-highlighted content, we skip wrapping (ANSI codes break lipgloss.Wrap)
			// and apply background overlays directly.
			if !isMarkdown && chromaLines != nil && i < len(chromaLines) {
				styledLine := displayLine
				if isSelected {
					styledLine = inlineBg(selectedLineBg, styledLine)
				} else if isSidebarHighlight {
					styledLine = inlineBg(sidebarHighlightBg, styledLine)
				} else if isChanged {
					styledLine = inlineBg(diffChangedLineBg, styledLine)
				}
				b.WriteString(fmt.Sprintf("%s%s %s\n", marker, numStr, styledLine))
			} else {
				// Markdown or plain text path with wrapping
				styleFunc := func(s string) string { return s }
				if isMarkdown {
					styleFunc = func(s string) string { return highlightMarkdown(s) }
				}
				if isSelected {
					styleFunc = func(s string) string { return inlineBg(selectedLineBg, s) }
				} else if isSidebarHighlight {
					styleFunc = func(s string) string { return inlineBg(sidebarHighlightBg, s) }
				} else if isChanged {
					base := styleFunc
					styleFunc = func(s string) string { return inlineBg(diffChangedLineBg, base(s)) }
				}

				wrapped := line
				if m.contentViewport.SoftWrap {
					wrapped = lipgloss.Wrap(line, textWidth, "")
				}
				wrappedLines := strings.Split(wrapped, "\n")
				for wi, wl := range wrappedLines {
					if wi == 0 {
						b.WriteString(fmt.Sprintf("%s%s %s\n", marker, numStr, styleFunc(wl)))
					} else {
						b.WriteString(fmt.Sprintf(" %s %s\n", continuationGutter, styleFunc(wl)))
					}
				}
			}
		}

		// Render inline annotations after this line
		if anns, ok := annosByEndLine[lineNum]; ok {
			for idx, ann := range anns {
				focused := m.focused == contentPane && t.cursorOnAnnotation && t.cursorLine == lineNum && t.cursorAnnoIdx == idx
				b.WriteString(m.renderAnnotationBox(ann, boxWidth, focused))
			}
		}
	}

	m.contentViewport.SetContent(b.String())
}

// renderAnnotationBox renders a bordered annotation box indented under the gutter.
func (m *AppModel) renderAnnotationBox(ann annotation, maxWidth int, focused bool) string {
	var lineLabel string
	if ann.endLine > 0 {
		lineLabel = fmt.Sprintf("L%d-%d", ann.line, ann.endLine)
	} else {
		lineLabel = fmt.Sprintf("L%d", ann.line)
	}

	var boxContent strings.Builder
	label := inlineLabelComment.Render("comment")
	lineRef := commentLineStyle.Render(lineLabel)
	boxContent.WriteString(fmt.Sprintf("%s %s\n", label, lineRef))
	boxContent.WriteString(clampLines(ann.body, 3))
	boxStyle := inlineCommentBox

	if focused {
		boxStyle = boxStyle.BorderForeground(warning)
	}
	box := boxStyle.Width(maxWidth).Render(boxContent.String())

	var prefix string
	if focused {
		cursor := lipgloss.NewStyle().Width(2).Render(cursorMarker.Render(">"))
		prefix = cursor + strings.Repeat(" ", gutterWidth-2)
	} else {
		prefix = strings.Repeat(" ", gutterWidth)
	}

	var b strings.Builder
	for _, line := range strings.Split(box, "\n") {
		b.WriteString(prefix + line + "\n")
	}
	return b.String()
}

const chunkScrollPadding = 4

func (m *AppModel) scrollToCursor() {
	t := m.tab()
	if t.doc == nil {
		return
	}

	renderedLine := 0
	extraCounts := m.extraLinesPerDocLine()
	for i := 1; i < t.cursorLine; i++ {
		renderedLine++
		renderedLine += extraCounts[i]
	}

	cursorBottom := renderedLine + 1 + extraCounts[t.cursorLine]

	vpHeight := m.contentViewport.Height()
	currentTop := m.contentViewport.YOffset()

	if renderedLine < currentTop {
		m.contentViewport.SetYOffset(renderedLine)
	}
	if cursorBottom > currentTop+vpHeight {
		m.contentViewport.SetYOffset(cursorBottom - vpHeight)
	}
}

// scrollToChunk scrolls the viewport to show the entire change chunk
// plus padding lines above and below for context.
func (m *AppModel) scrollToChunk(chunk changeChunk) {
	t := m.tab()
	if t.doc == nil {
		return
	}

	extraCounts := m.extraLinesPerDocLine()

	// Compute rendered line for chunk start - padding
	startLine := chunk.startLine - chunkScrollPadding
	if startLine < 1 {
		startLine = 1
	}
	startRendered := 0
	for i := 1; i < startLine; i++ {
		startRendered++
		startRendered += extraCounts[i]
	}

	// Compute rendered line for chunk end + padding
	endLine := chunk.endLine + chunkScrollPadding
	if endLine > t.doc.LineCount() {
		endLine = t.doc.LineCount()
	}
	endRendered := 0
	for i := 1; i <= endLine; i++ {
		endRendered++
		endRendered += extraCounts[i]
	}

	vpHeight := m.contentViewport.Height()

	// If the whole chunk+padding fits, position start at top
	if endRendered-startRendered <= vpHeight {
		m.contentViewport.SetYOffset(startRendered)
	} else {
		// Chunk is taller than viewport — just put cursor near top with padding
		m.contentViewport.SetYOffset(startRendered)
	}
}

func (m *AppModel) scrollToAnnotation(startLine, endLine int) {
	t := m.tab()
	if t.doc == nil {
		return
	}
	if endLine == 0 {
		endLine = startLine
	}

	extraCounts := m.extraLinesPerDocLine()

	startRendered := 0
	for i := 1; i < startLine; i++ {
		startRendered++
		startRendered += extraCounts[i]
	}

	endRendered := 0
	for i := 1; i <= endLine; i++ {
		endRendered++
		endRendered += extraCounts[i]
	}

	vpHeight := m.contentViewport.Height()

	offset := endRendered - vpHeight
	if offset < 0 {
		offset = 0
	}
	if offset > startRendered {
		offset = startRendered
	}

	m.contentViewport.SetYOffset(offset)
}

func (m *AppModel) extraLinesPerDocLine() map[int]int {
	t := m.tab()
	counts := make(map[int]int)
	if t.doc == nil {
		return counts
	}

	contentWidth := m.contentViewport.Width()
	textWidth := contentWidth - 8
	if textWidth < 10 {
		textWidth = 10
	}

	for i, line := range t.doc.Lines {
		lineNum := i + 1
		wrapped := lipgloss.Wrap(line, textWidth, "")
		wrapCount := strings.Count(wrapped, "\n")
		if wrapCount > 0 {
			counts[lineNum] += wrapCount
		}
	}

	if t.state != nil {
		for _, c := range t.state.Comments {
			endAt := c.Line
			if c.EndLine > 0 {
				endAt = c.EndLine
			}
			bodyLines := strings.Count(c.Body, "\n") + 1
			counts[endAt] += bodyLines + 3
		}
	}

	// Account for deleted lines rendered before each doc line
	if t.deletedAfter != nil {
		for afterLine, dels := range t.deletedAfter {
			targetLine := afterLine + 1
			if targetLine < 1 {
				targetLine = 1
			}
			counts[targetLine] += len(dels)
		}
	}

	return counts
}

func (m *AppModel) updateCommentSidebar() {
	t := m.tab()
	if t.state == nil {
		return
	}

	t.sidebarItems = nil
	for _, c := range t.state.Comments {
		t.sidebarItems = append(t.sidebarItems, sidebarItem{
			id: c.ID, line: c.Line, endLine: c.EndLine,
			body: c.Body,
		})
	}
	sort.Slice(t.sidebarItems, func(i, j int) bool { return t.sidebarItems[i].line < t.sidebarItems[j].line })

	if t.sidebarCursor >= len(t.sidebarItems) {
		t.sidebarCursor = len(t.sidebarItems) - 1
	}
	if t.sidebarCursor < 0 {
		t.sidebarCursor = 0
	}

	var b strings.Builder

	if len(t.sidebarItems) == 0 {
		b.WriteString(commentStyle.Render("No comments yet.\n\nPress enter to comment.\n\nUse 'v' to select\nmultiple lines first."))
		m.commentViewport.SetContent(b.String())
		return
	}

	for idx, it := range t.sidebarItems {
		isSelected := m.focused == commentPane && idx == t.sidebarCursor

		var lineInfo string
		if it.endLine > 0 {
			lineInfo = fmt.Sprintf("L%d-%d", it.line, it.endLine)
		} else {
			lineInfo = fmt.Sprintf("L%d", it.line)
		}
		lineInfo = commentLineStyle.Render(lineInfo)

		cursorCol := lipgloss.NewStyle().Width(2)
		prefix := cursorCol.Render("")
		if isSelected {
			prefix = cursorCol.Render(cursorMarker.Render(">"))
		}

		b.WriteString(fmt.Sprintf("%s%s\n", prefix, lineInfo))

		clamped := clampLines(it.body, 3)
		bodyLines := strings.Split(clamped, "\n")
		for i, bl := range bodyLines {
			styled := bl
			if isSelected {
				styled = sidebarSelectedText.Render(bl)
			} else {
				styled = commentStyle.Render(bl)
			}
			b.WriteString(" " + styled)
			if i < len(bodyLines)-1 {
				b.WriteString("\n")
			}
		}
		b.WriteString("\n\n")
	}

	m.commentViewport.SetContent(b.String())
}

// renderInlineBg applies a background colour to just the content text (no full-width padding).
// For content with embedded ANSI codes (Chroma), it re-injects the bg after every reset so
// the chosen background remains visible across syntax-highlighted spans.
func renderInlineBg(style lipgloss.Style, content string) string {
bgAnsi := bgToAnsi(style.GetBackground())
if bgAnsi == "" {
return style.Render(content)
}
patched := strings.ReplaceAll(content, "\033[0m", "\033[0m"+bgAnsi)
return bgAnsi + patched + "\033[0m"
}

// navigateToComment moves the cursor to the next (forward=true) or previous
// (forward=false) comment in the active tab, wrapping around if necessary.
// Returns true if the cursor was moved.
func (m *AppModel) navigateToComment(forward bool) bool {
t := m.tab()
if t.state == nil || len(t.state.Comments) == 0 {
return false
}

// isBetter returns true if candidate endLine is a better match than current best.
isBetter := func(best, candidate int) bool {
if forward {
return candidate < best
}
return candidate > best
}

// isCandidate returns true if endLine qualifies as a next/prev target.
isCandidate := func(endAt int) bool {
if forward {
return endAt > t.cursorLine || (endAt == t.cursorLine && !t.cursorOnAnnotation)
}
return endAt < t.cursorLine || (endAt == t.cursorLine && !t.cursorOnAnnotation)
}

best := -1
for _, c := range t.state.Comments {
endAt := c.Line
if c.EndLine > 0 {
endAt = c.EndLine
}
if isCandidate(endAt) && (best == -1 || isBetter(best, endAt)) {
best = endAt
}
}

// Wrap around: pick the first (or last) comment overall.
if best == -1 {
for _, c := range t.state.Comments {
endAt := c.Line
if c.EndLine > 0 {
endAt = c.EndLine
}
if best == -1 || isBetter(best, endAt) {
best = endAt
}
}
}

if best != -1 {
t.cursorLine = best
t.cursorOnAnnotation = true
t.cursorAnnoIdx = 0
return true
}
return false
}
