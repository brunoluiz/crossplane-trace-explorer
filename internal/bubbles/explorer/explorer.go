package explorer

import (
	"fmt"
	"strings"

	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/explorer/statusbar"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/explorer/viewer"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/tree"
	"github.com/brunoluiz/crossplane-explorer/internal/tui"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/samber/lo"
	k8sv1 "k8s.io/api/core/v1"
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
	case tea.KeyMsg:
		cmd = m.onKey(msg)
	}

	t, treeCmd := m.tree.Update(msg)
	s, statusCmd := m.statusbar.Update(msg)
	v, viewerCmd := m.viewer.Update(msg)
	m.tree, m.statusbar, m.viewer = t.(*tree.Model), s.(*statusbar.Model), v.(*viewer.Model)

	return m, tea.Batch(cmd, treeCmd, statusCmd, viewerCmd)
}

func (m Model) View() string {
	switch m.pane {
	case PaneSummary:
		return m.viewer.View()
	case PaneTree:
		return lipgloss.JoinVertical(
			lipgloss.Top,
			lipgloss.NewStyle().Height(m.height-m.statusbar.GetHeight()).Render(m.tree.View()),
			m.statusbar.View(),
		)
	default:
		return "No pane selected"
	}
}

func addNodes(v *xplane.Resource, n *tree.Node, resByNode map[*tree.Node]*xplane.Resource) {
	group := v.Unstructured.GetObjectKind().GroupVersionKind().Group
	ready := v.GetCondition(v1.TypeReady)
	synced := v.GetCondition(v1.TypeSynced)

	n.Key = fmt.Sprintf("%s/%s", v.Unstructured.GetKind(), v.Unstructured.GetName())
	n.Value = fmt.Sprintf("%s.%s/%s", v.Unstructured.GetKind(), group, v.Unstructured.GetName())
	n.Children = make([]*tree.Node, len(v.Children))

	if v.Unstructured.GetAnnotations()["crossplane.io/paused"] == "true" {
		n.Key += " (paused)"
		n.Unselected = tree.ColorConfig{Foreground: lipgloss.ANSIColor(ansi.Yellow)}
	}

	if synced.Status == k8sv1.ConditionFalse || ready.Status == k8sv1.ConditionFalse {
		n.Unselected = tree.ColorConfig{Foreground: lipgloss.ANSIColor(ansi.Red)}
	}

	n.Details = map[string]string{
		HeaderKeyGroup:      group,
		HeaderKeySynced:     string(synced.Status),
		HeaderKeySyncedLast: synced.LastTransitionTime.Format(tui.DateFormat),
		HeaderKeyReady:      string(ready.Status),
		HeaderKeyReadyLast:  ready.LastTransitionTime.Format(tui.DateFormat),
		HeaderKeyMessage:    lo.Elipse(strings.Join(v.GetUnhealthyStatus(), ", "), 96),
	}

	resByNode[n] = v

	for k, cv := range v.Children {
		n.Children[k] = &tree.Node{}
		addNodes(cv, n.Children[k], resByNode)
	}
}
