package tui

import "os"

// recalculateLayout recomputes viewport dimensions and caches header/tab-bar heights
// for use in mouse hit-testing.
func (m *AppModel) recalculateLayout() {
	headerHeight := 1
	if m.detached {
		headerHeight = 2
	}
	tabBarHeight := 0
	if m.multiFile {
		tabBarHeight = 3 // bordered tabs: top border + content + bottom border
	}
	footerHeight := 1
	tmuxPadding := 0
	if os.Getenv("TMUX") != "" {
		tmuxPadding = 1
	}
	frameBorderHeight := 0
	frameBorderWidth := 0
	if m.multiFile {
		frameBorderHeight = 1 // bottom border
		frameBorderWidth = 2  // left + right borders
	}
	mainHeight := m.height - headerHeight - tabBarHeight - footerHeight - frameBorderHeight - tmuxPadding

	commentWidth := m.width / 4
	if commentWidth < 20 {
		commentWidth = 20
	}
	contentWidth := m.width - commentWidth - frameBorderWidth

	m.contentViewport.SetWidth(contentWidth)
	m.contentViewport.SetHeight(mainHeight)
	m.commentViewport.SetWidth(commentWidth - 3) // -3 for left border + padding + margin
	m.commentViewport.SetHeight(mainHeight - 1)  // -1 for the "Comments (N)" header line

	modalWidth := m.width * 2 / 3
	if modalWidth < 50 {
		modalWidth = 50
	}
	if modalWidth > m.width-4 {
		modalWidth = m.width - 4
	}
	m.modalTextarea.SetWidth(modalWidth - 10)
	m.modalTextarea.SetHeight(6)

	// Cache for mouse hit testing
	m.layoutHeaderHeight = headerHeight
	m.layoutTabBarHeight = tabBarHeight
}
