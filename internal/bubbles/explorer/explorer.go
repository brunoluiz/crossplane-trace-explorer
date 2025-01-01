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
	HeaderKeyMessage    = "MESSAGE"
)

type Pane string

const (
	PaneTree    Pane = "tree"
	PaneSummary Pane = "summary"
)

type Model struct {
	tree      *tree.Model
	statusbar *statusbar.Model
	viewer    *viewer.Model
	width     int
	height    int

	pane      Pane
	resByNode map[*tree.Node]*xplane.Resource
}

func New(
	treeModel *tree.Model,
	viewerModel *viewer.Model,
	statusbarModel *statusbar.Model,
) *Model {
	treeModel.OnSelectionChange = func(n *tree.Node) {
		statusbarModel.Update(statusbar.EventUpdatePath{
			Path: n.Path,
		})
	}

	return &Model{
		tree:      treeModel,
		statusbar: statusbarModel,
		viewer:    viewerModel,
		width:     0,
		height:    0,

		pane:      PaneTree,
		resByNode: map[*tree.Node]*xplane.Resource{},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case *xplane.Resource:
		cmd = m.onLoad(msg)
	case tea.WindowSizeMsg:
		cmd = m.onResize(msg)
		return m, nil
	case tea.KeyMsg:
		cmd = m.onKey(msg)
	}

	t, treeCmd := m.tree.Update(msg)
	s, statusCmd := m.statusbar.Update(msg)
	v, viewerCmd := m.viewer.Update(msg)

	//nolint // trust the typecast
	m.tree, m.statusbar, m.viewer = t.(*tree.Model), s.(*statusbar.Model), v.(*viewer.Model)

	return m, tea.Batch(cmd, treeCmd, statusCmd, viewerCmd)
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
