package tui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/tobiashochguertel/crit/internal/document"
	"github.com/tobiashochguertel/crit/internal/review"
)

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		initAdaptiveStyles(msg.IsDark())
		if len(m.tabs) > 0 && m.tab().state != nil {
			m.rebuildContent()
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalculateLayout()
		if len(m.tabs) > 0 && m.tab().state != nil {
			m.rebuildContent()
		}
		return m, nil

	case docRenderedMsg:
		// Load documents and initialize fresh review state for each tab
		for i := range m.tabs {
			t := &m.tabs[i]
			if t.isBinary || t.isDeleted {
				t.state = &review.ReviewState{
					File:     t.path,
					Comments: []review.Comment{},
				}
				continue
			}
			doc, _ := document.Load(t.path)
			t.doc = doc
			t.state = &review.ReviewState{
				File:     t.path,
				Comments: []review.Comment{},
			}
			t.ensureHighlightCache()
		}

		m.rebuildContent()
		m.updateCommentSidebar()
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case tea.MouseWheelMsg:
		return m.handleMouseWheel(msg)

	case tea.MouseClickMsg:
		return m.handleMouseClick(msg)

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	var cmd tea.Cmd
	if m.modal != noModal {
		m.modalTextarea, cmd = m.modalTextarea.Update(msg)
		return m, cmd
	}

	switch m.focused {
	case contentPane:
		m.contentViewport, cmd = m.contentViewport.Update(msg)
	case commentPane:
		m.commentViewport, cmd = m.commentViewport.Update(msg)
	}

	return m, cmd
}

func (m *AppModel) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.modal == commentModal || m.modal == editModal {
		return m.handleTextModal(msg)
	}

	// Tab search input mode
	if m.tabSearching {
		return m.handleTabSearch(msg)
	}

	t := m.tab()

	switch {
	case key.Matches(msg, keys.Quit):
		// Auto-save all tabs on quit
		for i := range m.tabs {
			if m.tabs[i].state != nil {
				m.SaveFn(m.tabs[i].state)
			}
		}
		return m, tea.Quit

	case key.Matches(msg, keys.Cancel):
		// Esc cancels selection
		if t.selecting {
			t.selecting = false
			m.rebuildContent()
			return m, nil
		}
		return m, nil

	case key.Matches(msg, keys.Tab):
		if !t.selecting {
			if m.focused == contentPane {
				m.focused = commentPane
			} else {
				m.focused = contentPane
			}
			m.updateCommentSidebar()
			m.rebuildContent()
		}
		return m, nil

	case key.Matches(msg, keys.VisualMode):
		if m.focused == contentPane && t.doc != nil {
			if t.selecting {
				t.selecting = false
			} else {
				t.selecting = true
				t.selectAnchor = t.cursorLine
			}
			m.rebuildContent()
			return m, nil
		}
	}

	// Tab switching (multi-file mode)
	if m.multiFile && m.focused == contentPane && !t.selecting {
		switch {
		case key.Matches(msg, keys.PrevTab):
			if m.activeTab > 0 {
				m.activeTab--
				m.rebuildContent()
				m.updateCommentSidebar()
			}
			return m, nil
		case key.Matches(msg, keys.NextTab):
			if m.activeTab < len(m.tabs)-1 {
				m.activeTab++
				m.rebuildContent()
				m.updateCommentSidebar()
			}
			return m, nil
		case key.Matches(msg, keys.TabSearch):
			m.tabSearching = true
			m.tabSearch = ""
			m.tabMatches = nil
			return m, nil
		}
		// Number keys 1-9 for direct tab access
		if s := msg.String(); len(s) == 1 && s[0] >= '1' && s[0] <= '9' {
			idx := int(s[0]-'0') - 1
			if idx < len(m.tabs) {
				m.activeTab = idx
				m.rebuildContent()
				m.updateCommentSidebar()
			}
			return m, nil
		}
	}

	// Content pane cursor movement (annotation-aware)
	if m.focused == contentPane && t.doc != nil {
		moved := false
		switch {
		case key.Matches(msg, keys.Down):
			if t.cursorOnAnnotation {
				anns := m.annotationsAfterLine(t.cursorLine)
				if t.cursorAnnoIdx < len(anns)-1 {
					t.cursorAnnoIdx++
				} else {
					t.cursorOnAnnotation = false
					t.cursorAnnoIdx = 0
					if t.cursorLine < t.doc.LineCount() {
						t.cursorLine++
					}
				}
			} else {
				anns := m.annotationsAfterLine(t.cursorLine)
				if len(anns) > 0 {
					t.cursorOnAnnotation = true
					t.cursorAnnoIdx = 0
				} else if t.cursorLine < t.doc.LineCount() {
					t.cursorLine++
				}
			}
			moved = true
		case key.Matches(msg, keys.Up):
			if t.cursorOnAnnotation {
				if t.cursorAnnoIdx > 0 {
					t.cursorAnnoIdx--
				} else {
					t.cursorOnAnnotation = false
					t.cursorAnnoIdx = 0
				}
			} else {
				prevLine := t.cursorLine - 1
				if prevLine >= 1 {
					anns := m.annotationsAfterLine(prevLine)
					if len(anns) > 0 {
						t.cursorLine = prevLine
						t.cursorOnAnnotation = true
						t.cursorAnnoIdx = len(anns) - 1
					} else {
						t.cursorLine = prevLine
					}
				}
			}
			moved = true
		case key.Matches(msg, keys.HalfPageDown):
			t.cursorOnAnnotation = false
			t.cursorAnnoIdx = 0
			jump := m.contentViewport.Height() / 2
			t.cursorLine += jump
			if t.cursorLine > t.doc.LineCount() {
				t.cursorLine = t.doc.LineCount()
			}
			moved = true
		case key.Matches(msg, keys.HalfPageUp):
			t.cursorOnAnnotation = false
			t.cursorAnnoIdx = 0
			jump := m.contentViewport.Height() / 2
			t.cursorLine -= jump
			if t.cursorLine < 1 {
				t.cursorLine = 1
			}
			moved = true
		case key.Matches(msg, keys.Top):
			t.cursorOnAnnotation = false
			t.cursorAnnoIdx = 0
			t.cursorLine = 1
			moved = true
		case key.Matches(msg, keys.Bottom):
			t.cursorOnAnnotation = false
			t.cursorAnnoIdx = 0
			t.cursorLine = t.doc.LineCount()
			moved = true
		case key.Matches(msg, keys.NextComment):
			if m.navigateToComment(true) {
				moved = true
			}
		case key.Matches(msg, keys.PrevComment):
			if m.navigateToComment(false) {
				moved = true
			}
		case key.Matches(msg, keys.NextChange):
			if len(t.changeChunks) > 0 {
				target := -1
				for i, chunk := range t.changeChunks {
					if chunk.startLine > t.cursorLine {
						target = i
						break
					}
				}
				if target == -1 {
					target = 0
				}
				chunk := t.changeChunks[target]
				t.cursorLine = chunk.startLine
				t.cursorOnAnnotation = false
				t.cursorAnnoIdx = 0
				m.rebuildContent()
				m.scrollToChunk(chunk)
				return m, nil
			}
		case key.Matches(msg, keys.PrevChange):
			if len(t.changeChunks) > 0 {
				target := -1
				for i := len(t.changeChunks) - 1; i >= 0; i-- {
					if t.changeChunks[i].startLine < t.cursorLine {
						target = i
						break
					}
				}
				if target == -1 {
					target = len(t.changeChunks) - 1
				}
				chunk := t.changeChunks[target]
				t.cursorLine = chunk.startLine
				t.cursorOnAnnotation = false
				t.cursorAnnoIdx = 0
				m.rebuildContent()
				m.scrollToChunk(chunk)
				return m, nil
			}
		case key.Matches(msg, keys.Confirm):
			if t.cursorOnAnnotation {
				anns := m.annotationsAfterLine(t.cursorLine)
				if t.cursorAnnoIdx < len(anns) {
					ann := anns[t.cursorAnnoIdx]
					m.editingID = ann.id
					m.modal = editModal
					m.modalFocus = 0
					m.modalTextarea.Reset()
					m.modalTextarea.SetValue(ann.body)
					m.modalTextarea.Placeholder = "Edit comment..."
					m.modalTextarea.Focus()
					return m, nil
				}
			} else if t.state != nil {
				m.modal = commentModal
				m.modalFocus = 0
				m.modalTextarea.Placeholder = "Type your comment..."
				m.modalTextarea.Reset()
				m.modalTextarea.Focus()
				return m, nil
			}
		}

		if moved {
			m.rebuildContent()
			m.scrollToCursor()
			return m, nil
		}
	}

	// Comment pane navigation
	if m.focused == commentPane && len(t.sidebarItems) > 0 {
		sidebarMoved := false
		switch {
		case key.Matches(msg, keys.Down):
			if t.sidebarCursor < len(t.sidebarItems)-1 {
				t.sidebarCursor++
				sidebarMoved = true
			}
		case key.Matches(msg, keys.Up):
			if t.sidebarCursor > 0 {
				t.sidebarCursor--
				sidebarMoved = true
			}
		case key.Matches(msg, keys.Top):
			t.sidebarCursor = 0
			sidebarMoved = true
		case key.Matches(msg, keys.Bottom):
			t.sidebarCursor = len(t.sidebarItems) - 1
			sidebarMoved = true
		}
		if sidebarMoved {
			m.updateCommentSidebar()
			m.rebuildContent()
			sel := t.sidebarItems[t.sidebarCursor]
			t.cursorLine = sel.line
			m.scrollToAnnotation(sel.line, sel.endLine)
			return m, nil
		}

		// Enter to edit selected annotation
		if key.Matches(msg, keys.Confirm) {
			sel := t.sidebarItems[t.sidebarCursor]
			m.editingID = sel.id
			m.modal = editModal
			m.modalFocus = 0
			m.modalTextarea.Reset()
			m.modalTextarea.SetValue(sel.body)
			m.modalTextarea.Placeholder = "Edit comment..."
			m.modalTextarea.Focus()
			return m, nil
		}
	}

	if key.Matches(msg, keys.ScrollLeft) {
		if m.focused == contentPane {
			m.contentViewport.ScrollLeft(4)
		} else {
			m.commentViewport.ScrollLeft(4)
		}
		return m, nil
	}
	if key.Matches(msg, keys.ScrollRight) {
		if m.focused == contentPane {
			m.contentViewport.ScrollRight(4)
		} else {
			m.commentViewport.ScrollRight(4)
		}
		return m, nil
	}

	return m, nil
}

func (m *AppModel) handleTabSearch(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.tabSearching = false
		m.tabSearch = ""
		m.tabMatches = nil
		return m, nil
	case "enter":
		if len(m.tabMatches) > 0 {
			m.activeTab = m.tabMatches[0]
			m.rebuildContent()
			m.updateCommentSidebar()
		}
		m.tabSearching = false
		m.tabSearch = ""
		m.tabMatches = nil
		return m, nil
	case "backspace":
		if len(m.tabSearch) > 0 {
			m.tabSearch = m.tabSearch[:len(m.tabSearch)-1]
			m.updateTabSearchMatches()
		}
		return m, nil
	case "tab":
		// Cycle to next match
		if len(m.tabMatches) > 1 {
			// Rotate matches
			m.tabMatches = append(m.tabMatches[1:], m.tabMatches[0])
		}
		return m, nil
	default:
		s := msg.String()
		if len(s) == 1 && s[0] >= ' ' && s[0] <= '~' {
			m.tabSearch += s
			m.updateTabSearchMatches()
		}
		return m, nil
	}
}

func (m *AppModel) updateTabSearchMatches() {
	m.tabMatches = nil
	if m.tabSearch == "" {
		return
	}
	query := strings.ToLower(m.tabSearch)
	for i, t := range m.tabs {
		if strings.Contains(strings.ToLower(t.path), query) {
			m.tabMatches = append(m.tabMatches, i)
		}
	}
}

