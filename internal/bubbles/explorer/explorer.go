package explorer

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/explorer/statusbar"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/explorer/viewer"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/tree"
	"github.com/brunoluiz/crossplane-explorer/internal/tui"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	tree      tree.Model
	statusbar statusbar.Model
	viewer    viewer.Model
	width     int
	height    int

	pane      Pane
	resByNode map[*tree.Node]*xplane.Resource
}

func New(data *xplane.Resource) Model {
	nodes := []*tree.Node{
		{Key: "root", Children: make([]*tree.Node, 1)},
	}
	resByNode := map[*tree.Node]*xplane.Resource{}
	addNodes(data, nodes[0], resByNode)

	t := tree.New(nodes, []string{
		HeaderKeyObject,
		HeaderKeyGroup,
		HeaderKeySynced,
		HeaderKeySyncedLast,
		HeaderKeyReady,
		HeaderKeyReadyLast,
		HeaderKeyMessage,
	})

	return Model{
		tree: t,
		statusbar: statusbar.New(
			statusbar.WithInitialPath([]string{nodes[0].Key}),
		),
		viewer:    viewer.New(),
		pane:      PaneTree,
		resByNode: resByNode,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.statusbar.SetSize(m.width)
		m.viewer.Update(msg)

		top, right, _, left := lipgloss.NewStyle().Padding(1).GetPadding()
		m.tree.SetSize(m.width-right-left, m.height-top)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			clipboard.WriteAll(m.tree.Current().Value)
		case "enter":
			v := m.resByNode[m.tree.Current()]

			m.viewer.Setup(v)
			m.pane = PaneSummary
		case "ctrl+c", "ctlr+d":
			return m, tea.Quit
		case "q", "esc":
			if m.pane == PaneTree {
				return m, tea.Quit
			} else {
				m.pane = PaneTree
			}
		case "?":
			return m, nil
		}
	}

	var treeCmd, statusCmd, viewerCmd tea.Cmd
	m.tree, treeCmd = m.tree.Update(msg)
	m.statusbar.SetPath(m.tree.Path())
	m.statusbar, statusCmd = m.statusbar.Update(msg)
	m.viewer, viewerCmd = m.viewer.Update(msg)

	return m, tea.Batch(treeCmd, statusCmd, viewerCmd)
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
		n.Unselected = tree.ColorConfig{Foreground: tui.ColorWarn}
	}

	if synced.Status == k8sv1.ConditionFalse || ready.Status == k8sv1.ConditionFalse {
		n.Unselected = tree.ColorConfig{Foreground: tui.ColorAlert}
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
