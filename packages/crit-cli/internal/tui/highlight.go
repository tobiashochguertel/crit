package tui

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"charm.land/lipgloss/v2"
)

// highlightCode returns syntax-highlighted lines for a file using Chroma.
// Each line is pre-rendered with ANSI escape codes.
func highlightCode(filename string, content string) []string {
	lexer := lexers.Match(filename)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("monokai")
	formatter := formatters.Get("terminal256")

	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		// Fallback to raw lines
		return strings.Split(content, "\n")
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return strings.Split(content, "\n")
	}

	// Split the ANSI output by newlines
	lines := strings.Split(buf.String(), "\n")

	// Chroma sometimes adds a trailing empty line from the final newline
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
}

// ── Markdown highlighting ─────────────────────────────────────────────────────

var (
	reBold       = regexp.MustCompile(`\*\*(.+?)\*\*`)
	reItalic     = regexp.MustCompile(`(?:^|[^*])\*([^*]+?)\*(?:[^*]|$)`)
	reCode       = regexp.MustCompile("`([^`]+)`")
	reLink       = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	reListItem   = regexp.MustCompile(`^(\s*[-*+]\s)(.*)$`)
	reCheckbox   = regexp.MustCompile(`^(\s*[-*+]\s)\[([ xX])\]\s(.*)$`)
	reNumList    = regexp.MustCompile(`^(\s*\d+\.\s)(.*)$`)
	reBlockquote = regexp.MustCompile(`^(\s*>\s?)(.*)$`)
	reHr         = regexp.MustCompile(`^(\s*)([-*_]{3,})\s*$`)
	reTableRow   = regexp.MustCompile(`^\s*\|.*\|\s*$`)
	reTableSep   = regexp.MustCompile(`^\s*\|[\s:]*[-]+[\s:|-]*\|\s*$`)
)

// highlightMarkdown applies markdown syntax highlighting to a single line.
func highlightMarkdown(line string) string {
	trimmed := strings.TrimSpace(line)

	if reHr.MatchString(line) {
		return mdHrStyle.Render("─────────────────────────────────")
	}

	if strings.HasPrefix(trimmed, "#### ") {
		return mdH4Style.Render(line)
	}
	if strings.HasPrefix(trimmed, "### ") {
		return mdH3Style.Render(line)
	}
	if strings.HasPrefix(trimmed, "## ") {
		return mdH2Style.Render(line)
	}
	if strings.HasPrefix(trimmed, "# ") {
		return mdH1Style.Render(line)
	}

	if reTableSep.MatchString(line) {
		return mdTableSepStyle.Render(line)
	}
	if reTableRow.MatchString(line) {
		cells := strings.Split(line, "|")
		var parts []string
		for i, cell := range cells {
			if i == 0 || i == len(cells)-1 {
				parts = append(parts, cell)
			} else {
				parts = append(parts, highlightInline(cell))
			}
		}
		return strings.Join(parts, mdTablePipe.Render("|"))
	}

	if loc := reBlockquote.FindStringSubmatchIndex(line); loc != nil {
		rest := line[loc[4]:loc[5]]
		return mdBlockquoteBar.Render("▎") + " " + mdBlockquoteStyle.Render(rest)
	}

	if loc := reCheckbox.FindStringSubmatchIndex(line); loc != nil {
		indent := line[loc[2]:loc[3]]
		checked := line[loc[4]:loc[5]]
		rest := line[loc[6]:loc[7]]
		if checked == "x" || checked == "X" {
			return indent + mdCheckboxDone.Render("✓") + " " + mdCheckboxDoneText.Render(rest)
		}
		return indent + mdCheckboxOpen.Render("☐") + " " + highlightInline(rest)
	}

	if loc := reListItem.FindStringSubmatchIndex(line); loc != nil {
		indent := line[loc[2]:loc[3]]
		rest := line[loc[4]:loc[5]]
		return mdListMarkerStyle.Render(indent) + highlightInline(rest)
	}
	if loc := reNumList.FindStringSubmatchIndex(line); loc != nil {
		marker := line[loc[2]:loc[3]]
		rest := line[loc[4]:loc[5]]
		return mdListMarkerStyle.Render(marker) + highlightInline(rest)
	}

	return highlightInline(line)
}

// tableBlock represents a contiguous range of markdown table lines.
type tableBlock struct {
	startLine int
	endLine   int
	colWidths []int
}

func detectTableBlocks(lines []string) []tableBlock {
	var blocks []tableBlock
	inTable := false
	var current tableBlock

	for i, line := range lines {
		isTable := reTableRow.MatchString(line) || reTableSep.MatchString(line)
		if isTable {
			if !inTable {
				inTable = true
				current = tableBlock{startLine: i + 1}
			}
			current.endLine = i + 1

			if !reTableSep.MatchString(line) {
				cells := parseTableCells(line)
				for len(current.colWidths) < len(cells) {
					current.colWidths = append(current.colWidths, 0)
				}
				for ci, cell := range cells {
					if len(cell) > current.colWidths[ci] {
						current.colWidths[ci] = len(cell)
					}
				}
			}
		} else {
			if inTable {
				blocks = append(blocks, current)
				inTable = false
			}
		}
	}
	if inTable {
		blocks = append(blocks, current)
	}
	return blocks
}

func parseTableCells(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.Trim(line, "|")
	parts := strings.Split(line, "|")
	cells := make([]string, len(parts))
	for i, p := range parts {
		cells[i] = strings.TrimSpace(p)
	}
	return cells
}

func formatTableRow(line string, colWidths []int, isHeader bool) string {
	cells := parseTableCells(line)
	pipe := mdTablePipe.Render("│")

	var parts []string
	for ci := 0; ci < len(colWidths); ci++ {
		w := colWidths[ci]
		cell := ""
		if ci < len(cells) {
			cell = cells[ci]
		}
		padded := lipgloss.NewStyle().Width(w).Render(cell)
		if isHeader {
			parts = append(parts, mdTableHeaderStyle.Render(" "+padded+" "))
		} else {
			parts = append(parts, mdTableCellStyle.Render(" "+padded+" "))
		}
	}

	return pipe + strings.Join(parts, pipe) + pipe
}

func formatTableSep(colWidths []int) string {
	pipe := mdTablePipe.Render("│")
	var parts []string
	for _, w := range colWidths {
		parts = append(parts, mdTableSepStyle.Render(strings.Repeat("─", w+2)))
	}
	return pipe + strings.Join(parts, mdTablePipe.Render("┼")) + pipe
}

func highlightInline(line string) string {
	line = reCode.ReplaceAllStringFunc(line, func(match string) string {
		inner := match[1 : len(match)-1]
		return mdCodeStyle.Render(" " + inner + " ")
	})

	line = reBold.ReplaceAllStringFunc(line, func(match string) string {
		inner := match[2 : len(match)-2]
		return mdBoldStyle.Render(inner)
	})

	line = reItalic.ReplaceAllStringFunc(line, func(match string) string {
		start := 0
		end := len(match)
		if match[0] != '*' {
			start = 1
		}
		if match[end-1] != '*' {
			end--
		}
		inner := match[start+1 : end-1]
		prefix := match[:start]
		suffix := match[end:]
		return prefix + mdItalicStyle.Render(inner) + suffix
	})

	line = reLink.ReplaceAllStringFunc(line, func(match string) string {
		idx := strings.Index(match, "](")
		if idx < 0 {
			return match
		}
		text := match[1:idx]
		return mdLinkStyle.Render(text)
	})

	return line
}

// clampLines truncates text to maxLines, appending "…" if trimmed.
func clampLines(text string, maxLines int) string {
	lines := strings.Split(text, "\n")
	if len(lines) <= maxLines {
		return text
	}
	return strings.Join(lines[:maxLines], "\n") + "\n…"
}
