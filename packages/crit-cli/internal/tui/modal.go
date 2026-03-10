package tui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/google/uuid"

	"github.com/tobiashochguertel/crit/internal/review"
)

// modalSubmit saves the current modal textarea content as a new or edited comment.
func (m *AppModel) modalSubmit() {
	t := m.tab()
	body := strings.TrimSpace(m.modalTextarea.Value())
	if body == "" || t.state == nil {
		return
	}

	if m.modal == editModal {
		for i := range t.state.Comments {
			if t.state.Comments[i].ID == m.editingID {
				t.state.Comments[i].Body = body
				break
			}
		}
		m.editingID = ""
	} else {
		startLine, endLine := m.selectionRange()
		snippet := ""
		if t.doc != nil {
			snippet = strings.TrimSpace(t.doc.LineAt(startLine))
		}

		c := review.Comment{
			ID:             uuid.NewString()[:8],
			Line:           startLine,
			ContentSnippet: snippet,
			Body:           body,
			CreatedAt:      time.Now(),
		}
		if startLine != endLine {
			c.EndLine = endLine
		}
		t.state.AddComment(c)
	}

	m.SaveFn(t.state)
	m.modal = noModal
	m.modalTextarea.Blur()
	t.selecting = false
	m.rebuildContent()
	m.updateCommentSidebar()
}

// modalDelete removes the comment currently being edited.
func (m *AppModel) modalDelete() {
	t := m.tab()
	if t.state == nil || m.editingID == "" {
		return
	}
	t.state.DeleteComment(m.editingID)
	m.editingID = ""
	m.SaveFn(t.state)
	m.modal = noModal
	m.modalTextarea.Blur()
	t.cursorOnAnnotation = false
	t.cursorAnnoIdx = 0
	m.rebuildContent()
	m.updateCommentSidebar()
}

// handleTextModal processes key events while a comment modal is open.
func (m *AppModel) handleTextModal(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Number of focusable elements: edit modal has 4 (textarea, save, cancel, delete), comment modal has 3
	focusCount := 3
	if m.modal == editModal {
		focusCount = 4
	}

	switch msg.String() {
	case "esc":
		m.modal = noModal
		m.modalTextarea.Blur()
		return m, nil
	case "tab", "shift+tab":
		if msg.String() == "shift+tab" {
			m.modalFocus = (m.modalFocus + focusCount - 1) % focusCount
		} else {
			m.modalFocus = (m.modalFocus + 1) % focusCount
		}
		if m.modalFocus == 0 {
			m.modalTextarea.Focus()
		} else {
			m.modalTextarea.Blur()
		}
		return m, nil
	case "enter":
		if m.modalFocus == 1 {
			m.modalSubmit()
			return m, nil
		} else if m.modalFocus == 2 {
			m.modal = noModal
			m.modalTextarea.Blur()
			return m, nil
		} else if m.modalFocus == 3 && m.modal == editModal {
			m.modalDelete()
			return m, nil
		}
	case "ctrl+s":
		m.modalSubmit()
		return m, nil
	}

	if m.modalFocus == 0 {
		var cmd tea.Cmd
		m.modalTextarea, cmd = m.modalTextarea.Update(msg)
		return m, cmd
	}
	return m, nil
}
