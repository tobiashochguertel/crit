package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	tea "charm.land/bubbletea/v2"
)

func (m AppModel) View() tea.View {
	if m.err != nil {
		v := tea.NewView(fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err))
		v.AltScreen = true
		return v
	}

	if m.width == 0 || len(m.tabs) == 0 || m.tab().state == nil {
		v := tea.NewView("Loading...")
		v.AltScreen = true
		return v
	}

	t := m.tab()

	// Header
	commentCount := len(t.state.Comments)
	displayPath := t.path
	if m.filePath != "" {
		displayPath = m.filePath
	}
	var headerContent string
	if t.selecting {
		start, end := m.selectionRange()
		selLabel := visualModeIndicator.Render("VISUAL")
		headerContent = fmt.Sprintf(" Crit: %s  %s L%d-%d", displayPath, selLabel, start, end)
	} else if t.doc != nil {
		headerContent = fmt.Sprintf(" Crit: %s  %d comments  L%d/%d", displayPath, commentCount, t.cursorLine, t.doc.LineCount())
	} else {
		headerContent = fmt.Sprintf(" Crit: %s  %d comments", displayPath, commentCount)
	}
	var header string
	if m.detached {
		claudeBanner := claudeStatusBar.Width(m.width).Render(" Claude Code is paused — review the document, then press q to submit")
		header = claudeBanner + "\n" + headerStyle.Width(m.width).Render(headerContent)
	} else {
		header = headerStyle.Width(m.width).Render(headerContent)
	}

	// Tab bar (multi-file mode)
	var tabBar string
	if m.multiFile {
		tabBar = m.renderTabBar()
	}

	// Content pane
	commentWidth := m.width / 4
	if commentWidth < 20 {
		commentWidth = 20
	}
	contentWidth := m.width - commentWidth

	panelHeight := m.contentViewport.Height()

	contentBox := lipgloss.NewStyle().
		Width(contentWidth).
		Height(panelHeight).
		Render(m.contentViewport.View())

	// Comment sidebar (left border to separate from content)
	sidebarBorderColor := subtle
	if m.focused == commentPane {
		sidebarBorderColor = accent
	}
	sidebarBorder := lipgloss.Border{Left: "│"}
	commentHeader := lipgloss.NewStyle().Bold(true).Foreground(accent).Render(fmt.Sprintf("Comments (%d)", commentCount))
	commentBox := lipgloss.NewStyle().
		Border(sidebarBorder, false, false, false, true).
		BorderForeground(sidebarBorderColor).
		Width(commentWidth - 2).
		Height(panelHeight).
		PaddingLeft(1).
		Render(commentHeader + "\n" + m.commentViewport.View())

	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, contentBox, commentBox)

	// Wrap content in a frame: │ left/right borders, ╰───╯ bottom.
	// The tab bar serves as the top border.
	if m.multiFile {
		borderColor := lipgloss.NewStyle().Foreground(accent)
		lines := strings.Split(mainRow, "\n")
		var framed strings.Builder
		left := borderColor.Render("│")
		right := borderColor.Render("│")
		for _, line := range lines {
			framed.WriteString(left + line + right + "\n")
		}
		bottom := borderColor.Render("╰" + strings.Repeat("─", m.width-2) + "╯")
		framed.WriteString(bottom)
		mainRow = framed.String()
	}

	footer := m.renderFooter()

	var sections []string
	sections = append(sections, header)
	if tabBar != "" {
		sections = append(sections, tabBar)
	}
	sections = append(sections, mainRow, footer)

	full := lipgloss.JoinVertical(lipgloss.Left, sections...)

	if m.modal != noModal {
		full = m.renderWithModal(full)
	}

	v := tea.NewView(full)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// renderTabBar renders the tab bar for multi-file mode.
func (m *AppModel) renderTabBar() string {
	if m.tabSearching {
		prompt := tabSearchPromptStyle.Render("/")
		query := m.tabSearch
		matchInfo := ""
		if m.tabSearch != "" {
			matchInfo = fmt.Sprintf(" (%d matches)", len(m.tabMatches))
		}
		return prompt + query + footerStyle.Render(matchInfo)
	}

	// Disambiguate filenames — use basename unless there are collisions
	basenames := make(map[string]int)
	for _, t := range m.tabs {
		base := filepath.Base(t.path)
		basenames[base]++
	}

	type tabLabel struct {
		text     string
		rendered string
		width    int
	}

	labels := make([]tabLabel, len(m.tabs))
	for i, t := range m.tabs {
		label := filepath.Base(t.path)
		if basenames[label] > 1 {
			label = t.path
		}
		if n := len(t.changedLines); n > 0 {
			label += " " + tabChangeCount.Render(fmt.Sprintf("(+%d)", n))
		}
		labels[i] = tabLabel{text: label}
	}
	// Render a single tab with correct border corners for its position.
	// isFirst adjusts the left corner to connect to the outer frame border.
	renderTab := func(i int, isFirst bool) string {
		var style lipgloss.Style
		isActive := i == m.activeTab
		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		}
		style = style.Border(border)
		return style.Render(labels[i].text)
	}
	for i := range labels {
		rendered := renderTab(i, i == 0)
		labels[i].rendered = rendered
		labels[i].width = lipgloss.Width(rendered)
	}

	// addFiller extends the tab bottom border to the full width,
	// connecting to the outer frame's right border.
	addFiller := func(row string) string {
		rowW := lipgloss.Width(row)
		if rowW >= m.width {
			return row
		}
		// 3 lines matching tab height: empty top, empty middle, ───╮ bottom
		gap := m.width - rowW
		topFill := strings.Repeat(" ", gap)
		midFill := strings.Repeat(" ", gap)
		botFill := strings.Repeat("─", gap-1) + "╮"
		filler := lipgloss.NewStyle().Foreground(accent).Render(
			topFill + "\n" + midFill + "\n" + botFill,
		)
		return lipgloss.JoinHorizontal(lipgloss.Top, row, filler)
	}

	// Check if all tabs fit
	totalWidth := 0
	for _, l := range labels {
		totalWidth += l.width
	}

	if totalWidth <= m.width {
		var tabs []string
		for i := range labels {
			tabs = append(tabs, labels[i].rendered)
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
		return addFiller(row)
	}

	// Overflow: show a window of tabs centered on the active tab.
	// Indicators are styled as bordered tabs to align with the tab row.
	renderIndicator := func(text string, isFirst bool) string {
		style := inactiveTabStyle.Foreground(subtle)
		border, _, _, _, _ := style.GetBorder()
		if isFirst {
			border.BottomLeft = "├"
		}
		style = style.Border(border)
		return style.Render(text)
	}

	leftIndicator := ""
	rightIndicator := ""
	leftW := 0
	rightW := 0
	// Pre-render indicators to know their width for available space calc
	if m.activeTab > 0 {
		leftIndicator = renderIndicator(fmt.Sprintf("↤ %d more", m.activeTab), true)
		leftW = lipgloss.Width(leftIndicator)
	}
	if m.activeTab < len(labels)-1 {
		rightIndicator = renderIndicator(fmt.Sprintf("%d more ↦", len(labels)-m.activeTab-1), false)
		rightW = lipgloss.Width(rightIndicator)
	}

	availWidth := m.width - leftW - rightW

	// Find the window of tabs that fits
	start := m.activeTab
	end := m.activeTab + 1
	used := labels[m.activeTab].width

	// Expand window outward from active tab
	for {
		expanded := false
		if start > 0 && used+labels[start-1].width <= availWidth {
			start--
			used += labels[start].width
			expanded = true
		}
		if end < len(labels) && used+labels[end].width <= availWidth {
			used += labels[end].width
			end++
			expanded = true
		}
		if !expanded {
			break
		}
	}

	// Re-render indicators with actual counts now that we know the visible window
	var parts []string
	if start > 0 {
		ind := renderIndicator(fmt.Sprintf("↤ %d more", start), true)
		parts = append(parts, ind)
	}
	for i := start; i < end; i++ {
		parts = append(parts, renderTab(i, i == start && start == 0))
	}
	if end < len(labels) {
		ind := renderIndicator(fmt.Sprintf("%d more ↦", len(labels)-end), false)
		parts = append(parts, ind)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	return addFiller(row)
}

func (m AppModel) renderFooter() string {
	t := m.tabs[m.activeTab]
	k := func(key, desc string) string {
		return footerKeyStyle.Render(key) + " " + footerStyle.Render(desc)
	}

	var items []string
	if t.selecting {
		items = []string{
			k("j/k", "extend"),
			k("enter", "comment selection"),
			k("esc", "cancel"),
			k("v", "toggle select"),
		}
	} else {
		items = []string{
			k("j/k", "move"),
			k("[/]", "prev/next comment"),
			k("shift+↑↓", "page"),
			k("s", "sidebar"),
			k("v", "select"),
			k("enter", "comment"),
			k("q", "save & quit"),
		}
		if m.multiFile {
			items = append([]string{
				k("tab/S-tab", "next/prev tab"),
				k("n/N", "next/prev change"),
			}, items...)
		}
	}

	return footerStyle.Width(m.width).Render(strings.Join(items, "  "))
}

func (m AppModel) renderModalButton(label, hint string, focused bool) string {
	btn := modalBtnLabel.Render(label)
	h := modalBtnHint.Render(hint)
	content := btn + " " + h
	if focused {
		return modalBtnFocused.Render(content)
	}
	return modalBtnNormal.Render(content)
}

func (m AppModel) renderDeleteButton(label string, focused bool) string {
	if focused {
		return modalDeleteBtnFocused.Render(label)
	}
	return modalBtnNormal.Render(modalDeleteBtnLabel.Render(label))
}

func (m AppModel) renderContextPreview(start, end, maxWidth int) string {
	t := m.tabs[m.activeTab]
	if t.doc == nil {
		return ""
	}
	var lines []string
	maxLineText := maxWidth - 7
	if maxLineText < 10 {
		maxLineText = 10
	}
	for i := start; i <= end && i <= t.doc.LineCount(); i++ {
		lineText := t.doc.LineAt(i)
		wrapped := lipgloss.Wrap(lineText, maxLineText, "")
		num := lineNumStyle.Render(fmt.Sprintf("%d", i))
		wrapLines := strings.Split(wrapped, "\n")
		for wi, wl := range wrapLines {
			if wi == 0 {
				lines = append(lines, num+" "+wl)
			} else {
				lines = append(lines, lipgloss.NewStyle().Width(6).Render("")+wl)
			}
		}
	}
	if len(lines) > 8 {
		lines = append(lines[:7], footerStyle.Render(fmt.Sprintf("  ... +%d more lines", len(lines)-7)))
	}
	return strings.Join(lines, "\n")
}

func (m AppModel) renderWithModal(background string) string {
	var modalContent string
	modalWidth := m.width * 2 / 3
	if modalWidth < 50 {
		modalWidth = 50
	}
	if modalWidth > m.width-4 {
		modalWidth = m.width - 4
	}
	innerWidth := modalWidth - 6

	switch m.modal {
	case commentModal:
		start, end := m.selectionRange()
		var title string
		if start != end {
			title = modalTitleStyle.Render(fmt.Sprintf("Add Comment (lines %d-%d)", start, end))
		} else {
			title = modalTitleStyle.Render(fmt.Sprintf("Add Comment (line %d)", start))
		}
		contextBox := contextBoxStyle.
			Width(innerWidth - 2).
			Render(m.renderContextPreview(start, end, innerWidth-4))

		saveBtn := m.renderModalButton("Save", "ctrl+s", m.modalFocus == 1)
		cancelBtn := m.renderModalButton("Cancel", "esc", m.modalFocus == 2)
		buttons := lipgloss.JoinHorizontal(lipgloss.Center, saveBtn, "  ", cancelBtn)

		modalContent = modalStyle.Width(modalWidth).Render(
			title + "\n" + contextBox + "\n\n" + m.modalTextarea.View() + "\n\n" + buttons)

	case editModal:
		title := modalTitleStyle.Render("Edit Comment")
		var contextSection string
		for _, c := range m.tabs[m.activeTab].state.Comments {
			if c.ID == m.editingID {
				start := c.Line
				end := c.EndLine
				if end == 0 {
					end = start
				}
				contextSection = contextBoxStyle.
					Width(innerWidth - 2).
					Render(m.renderContextPreview(start, end, innerWidth-4))
				break
			}
		}
		saveBtn := m.renderModalButton("Save", "ctrl+s", m.modalFocus == 1)
		cancelBtn := m.renderModalButton("Cancel", "esc", m.modalFocus == 2)
		deleteBtn := m.renderDeleteButton("Delete", m.modalFocus == 3)
		buttons := lipgloss.JoinHorizontal(lipgloss.Center, saveBtn, "  ", cancelBtn, "  ", deleteBtn)

		content := title + "\n"
		if contextSection != "" {
			content += contextSection + "\n\n"
		}
		content += m.modalTextarea.View() + "\n\n" + buttons
		modalContent = modalStyle.Width(modalWidth).Render(content)
	}

	bgW := lipgloss.Width(background)
	bgH := lipgloss.Height(background)

	modalW := lipgloss.Width(modalContent)
	modalH := lipgloss.Height(modalContent)

	mx := (bgW - modalW) / 2
	my := (bgH - modalH) / 2
	if mx < 0 {
		mx = 0
	}
	if my < 0 {
		my = 0
	}

	background = dimRendered(background, bgW, bgH)

	bgLayer := lipgloss.NewLayer(background)
	modalLayer := lipgloss.NewLayer(modalContent).X(mx).Y(my).Z(1)

	comp := lipgloss.NewCompositor(bgLayer, modalLayer)
	return comp.Render()
}

func dimRendered(s string, w, h int) string {
	canvas := lipgloss.NewCanvas(w, h)
	canvas.Compose(lipgloss.NewLayer(s))

	dim := lipgloss.Color("#555555")
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cell := canvas.CellAt(x, y)
			if cell != nil {
				cell.Style.Fg = dim
			}
		}
	}
	return canvas.Render()
}
