package explorer

import (
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/explorer/statusbar"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/explorer/viewer"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/tree"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	HeaderKeyObject     = "OBJECT"
	HeaderKeyGroup      = "GROUP"
	HeaderKeySynced     = "SYNCED"
	HeaderKeySyncedLast = "SYNCED LAST"
	HeaderKeyReady      = "READY"
	HeaderKeyReadyLast  = "READY LAST"
	HeaderKeyStatus     = "STATUS"
)

type Pane string

const (
	PaneTree    Pane = "tree"
	PaneSummary Pane = "summary"
)

type Model struct {
	tree      tree.Model
	statusbar *statusbar.Model // requires pointer here
	viewer    viewer.Model
	width     int
	height    int

	pane      Pane
	resByNode map[*tree.Node]*xplane.Resource
}

func New(
	treeModel tree.Model,
	viewerModel viewer.Model,
	statusbarModel statusbar.Model,
) *Model {
	m := &Model{
		tree:      treeModel,
		statusbar: &statusbarModel,
		viewer:    viewerModel,
		width:     0,
		height:    0,

		pane:      PaneTree,
		resByNode: map[*tree.Node]*xplane.Resource{},
	}
	m.tree.OnSelectionChange = func(n *tree.Node) {
		m.statusbar.SetPath(n.Path)
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) View() string {
	switch m.pane {
	case PaneSummary:
		return m.viewer.View()
	case PaneTree:
		return lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Height(m.height-m.statusbar.GetHeight()).Render(m.tree.View()),
			m.statusbar.View(),
		)
	default:
		return "No pane selected"
	}
}
