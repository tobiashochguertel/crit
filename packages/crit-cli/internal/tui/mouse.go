package tui

import (
	"path/filepath"

	"charm.land/lipgloss/v2"

	tea "charm.land/bubbletea/v2"
)

// handleMouseWheel handles mouse wheel events.
// Ctrl/Alt + wheel scrolls horizontally; plain wheel scrolls vertically.
func (m AppModel) handleMouseWheel(msg tea.MouseWheelMsg) (tea.Model, tea.Cmd) {
	if msg.Mod.Contains(tea.ModCtrl) || msg.Mod.Contains(tea.ModAlt) {
		delta := 4
		switch msg.Button {
		case tea.MouseWheelLeft, tea.MouseWheelUp:
			if m.focused == contentPane {
				m.contentViewport.ScrollLeft(delta)
			} else {
				m.commentViewport.ScrollLeft(delta)
			}
		case tea.MouseWheelRight, tea.MouseWheelDown:
			if m.focused == contentPane {
				m.contentViewport.ScrollRight(delta)
			} else {
				m.commentViewport.ScrollRight(delta)
			}
		}
		return m, nil
	}
	// Default vertical scroll — pass to the focused viewport.
	var cmd tea.Cmd
	if m.focused == contentPane {
		m.contentViewport, cmd = m.contentViewport.Update(msg)
	} else {
		m.commentViewport, cmd = m.commentViewport.Update(msg)
	}
	return m, cmd
}

// handleMouseClick handles mouse click events: tab-bar selection, pane focus,
// and optional cursor positioning inside the content area.
func (m AppModel) handleMouseClick(msg tea.MouseClickMsg) (tea.Model, tea.Cmd) {
	if msg.Button != tea.MouseLeft {
		return m, nil
	}
	y := msg.Y
	x := msg.X

	tabBarStart := m.layoutHeaderHeight
	tabBarEnd := tabBarStart + m.layoutTabBarHeight

	switch {
	case m.multiFile && y >= tabBarStart && y < tabBarEnd:
		// Click inside the tab bar — resolve which tab was clicked.
		if idx := m.tabIndexAtX(x); idx >= 0 {
			m.activeTab = idx
			m.rebuildContent()
		}
	case y >= tabBarEnd:
		// Click in the main content area — split at the comment sidebar boundary.
		commentX := m.contentViewport.Width() + gutterWidth + 1
		if x >= commentX {
			m.focused = commentPane
		} else {
			m.focused = contentPane
		}
	}
	return m, nil
}

// tabIndexAtX recomputes the tab layout and returns the 0-based tab index
// corresponding to a click at terminal column x, or -1 if no tab was hit.
// It replicates the visible-window logic from renderTabBar() so that clicks
// are correctly resolved without relying on state cached during View().
func (m *AppModel) tabIndexAtX(clickX int) int {
	if len(m.tabs) == 0 {
		return -1
	}

	// Build short labels (basename, de-duplicated if needed).
	type labelInfo struct {
		text  string
		width int
	}
	labels := make([]labelInfo, len(m.tabs))
	basenames := make(map[string]int)
	for _, t := range m.tabs {
		basenames[filepath.Base(t.path)]++
	}
	for i, t := range m.tabs {
		base := filepath.Base(t.path)
		text := base
		if basenames[base] > 1 {
			dir := filepath.Base(filepath.Dir(t.path))
			text = dir + "/" + base
		}
		rendered := activeTabStyle.Render(text) // use widest style for measurement
		labels[i] = labelInfo{text: text, width: lipgloss.Width(rendered)}
	}

	// Determine visible window (same logic as renderTabBar).
	totalWidth := 0
	for _, l := range labels {
		totalWidth += l.width
	}

	start := 0
	end := len(labels)
	if totalWidth > m.width {
		indicatorWidth := lipgloss.Width("◀ ") + lipgloss.Width(" ▶")
		budget := m.width - indicatorWidth
		// Grow window around activeTab.
		s := m.activeTab
		e := m.activeTab + 1
		used := labels[m.activeTab].width
		for used < budget {
			grewLeft := false
			if s > 0 {
				w := labels[s-1].width
				if used+w <= budget {
					s--
					used += w
					grewLeft = true
				}
			}
			grewRight := false
			if e < len(labels) {
				w := labels[e].width
				if used+w <= budget {
					e++
					used += w
					grewRight = true
				}
			}
			if !grewLeft && !grewRight {
				break
			}
		}
		start = s
		end = e
		// x starts after left indicator.
		clickX -= lipgloss.Width("◀ ")
	}

	// Walk through visible tabs and find which one contains clickX.
	x := 0
	for i := start; i < end; i++ {
		w := labels[i].width
		if clickX >= x && clickX < x+w {
			return i
		}
		x += w
	}
	return -1
}
