package tui

import (
	"os"
	"sort"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/tobiashochguertel/crit/internal/document"
	gitpkg "github.com/tobiashochguertel/crit/internal/git"
	"github.com/tobiashochguertel/crit/internal/review"
)

type pane int

const (
	contentPane pane = iota
	commentPane
)

type modalType int

const (
	noModal modalType = iota
	commentModal
	editModal
)

// gutterWidth is the total width of the left gutter: line number (5) + marker (1) + space (1).
const gutterWidth = 7

// AppModel is the root TUI model containing all state for the crit viewer.
type AppModel struct {
	width, height int
	focused       pane
	modal         modalType

	// Multi-file tabs (code review mode)
	tabs         []FileTab
	activeTab    int
	multiFile    bool // true when in code review mode
	tabSearching bool
	tabSearch    string
	tabMatches   []int // indices of matching tabs during search

	// Single-file mode (legacy)
	filePath string

	detached bool

	contentViewport viewport.Model
	commentViewport viewport.Model
	modalTextarea   textarea.Model

	// Editing state
	editingID  string // ID of the comment being edited
	modalFocus int    // 0=textarea, 1=save button, 2=cancel button, 3=delete button (edit modal only)

	// Layout geometry cached for mouse event hit testing
	layoutHeaderHeight int
	layoutTabBarHeight int
	lastTabWidths      []int // rendered pixel widths of each tab in overflow-visible order
	lastTabStart       int   // index of the first tab visible in the tab overflow window

	err error

	// SaveFn is called to persist review state; defaults to review.Save but is
	// injectable for testing (Dependency Inversion Principle).
	SaveFn func(*review.ReviewState) error
}

// tab returns the active FileTab. Panics if no tabs exist.
func (m *AppModel) tab() *FileTab {
	return &m.tabs[m.activeTab]
}

// newModalTextarea creates a consistently-configured textarea for modal dialogs.
func newModalTextarea() textarea.Model {
	ta := textarea.New()
	ta.Placeholder = "Type your comment..."
	ta.ShowLineNumbers = false
	ta.CharLimit = 2000
	return ta
}

// NewApp creates a single-file TUI app.
func NewApp(filePath string) AppModel {
	tab := FileTab{
		path:       filePath,
		cursorLine: 1,
	}

	return AppModel{
		filePath:        filePath,
		tabs:            []FileTab{tab},
		activeTab:       0,
		detached:        os.Getenv("CRIT_DETACHED") == "1",
		contentViewport: viewport.New(),
		commentViewport: viewport.New(),
		modalTextarea:   newModalTextarea(),
		SaveFn:          review.Save,
	}
}

// NewCodeReviewApp creates a multi-file code review TUI.
func NewCodeReviewApp(files []gitpkg.FileChange, ref string) AppModel {
	// Sort files alphabetically by path
	sortedFiles := make([]gitpkg.FileChange, len(files))
	copy(sortedFiles, files)
	sort.Slice(sortedFiles, func(i, j int) bool {
		return sortedFiles[i].Path < sortedFiles[j].Path
	})

	tabs := make([]FileTab, 0, len(sortedFiles))
	for _, f := range sortedFiles {
		var diff *gitpkg.DiffInfo
		if f.Status != gitpkg.StatusDeleted && f.Status != gitpkg.StatusBinary {
			diff, _ = gitpkg.DiffFile(f.Path, ref)
		}
		ft := newFileTab(f.Path, diff)
		if f.Status == gitpkg.StatusBinary {
			ft.isBinary = true
		}
		if f.Status == gitpkg.StatusDeleted {
			ft.isDeleted = true
		}
		tabs = append(tabs, ft)
	}

	return AppModel{
		tabs:            tabs,
		activeTab:       0,
		multiFile:       true,
		detached:        os.Getenv("CRIT_DETACHED") == "1",
		contentViewport: viewport.New(),
		commentViewport: viewport.New(),
		modalTextarea:   newModalTextarea(),
		SaveFn:          review.Save,
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(m.loadDocuments(), tea.RequestBackgroundColor)
}

func (m AppModel) loadDocuments() tea.Cmd {
	return func() tea.Msg {
		for _, tab := range m.tabs {
			if tab.isBinary || tab.isDeleted {
				continue
			}
			if _, err := document.Load(tab.path); err != nil {
				return errMsg{err}
			}
		}
		return docRenderedMsg{}
	}
}

// selectionRange returns the ordered start/end of the current selection.
// If not selecting, returns cursorLine, cursorLine.
func (m *AppModel) selectionRange() (int, int) {
	t := m.tab()
	if !t.selecting {
		return t.cursorLine, t.cursorLine
	}
	start, end := t.selectAnchor, t.cursorLine
	if start > end {
		start, end = end, start
	}
	return start, end
}

// initDocuments is the docRenderedMsg handler extracted for readability.
func (m *AppModel) initDocuments() {
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
}
