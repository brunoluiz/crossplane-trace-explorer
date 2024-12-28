package explorer

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/brunoluiz/crossplane-trace-explorer/internal/bubbles/explorer/statusbar"
	"github.com/brunoluiz/crossplane-trace-explorer/internal/bubbles/tree"
	"github.com/brunoluiz/crossplane-trace-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/samber/lo"
	k8sv1 "k8s.io/api/core/v1"
)

func addNodes(v *xplane.Resource, n *tree.Node) {
	group := v.Unstructured.GetObjectKind().GroupVersionKind().Group
	ready := v.GetCondition(v1.TypeReady)
	synced := v.GetCondition(v1.TypeSynced)

	n.Key = fmt.Sprintf("%s/%s", v.Unstructured.GetKind(), v.Unstructured.GetName())
	n.Value = fmt.Sprintf("%s.%s/%s", v.Unstructured.GetKind(), group, v.Unstructured.GetName())
	n.Children = make([]tree.Node, len(v.Children))

	paused := v.Unstructured.GetAnnotations()["crossplane.io/paused"]
	if paused == "true" {
		n.Key += " (paused)"
		n.Selected = tree.ColorConfig{Background: lipgloss.Color("#FFB86C"), Foreground: lipgloss.Color("#FFFFFF")}
		n.Unselected = tree.ColorConfig{Foreground: lipgloss.Color("#FFB86C")}
	}

	if synced.Status == k8sv1.ConditionFalse || ready.Status == k8sv1.ConditionFalse {
		n.Selected = tree.ColorConfig{Background: lipgloss.Color("#FF5555"), Foreground: lipgloss.Color("#FFFFFF")}
		n.Unselected = tree.ColorConfig{Foreground: lipgloss.Color("#FF5555")}
	}

	n.Details = []string{
		group,
		string(synced.Status),
		synced.LastTransitionTime.Format("02 Jan 06 15:04"),
		string(ready.Status),
		ready.LastTransitionTime.Format("02 Jan 06 15:04"),
		lo.Elipse(strings.Join(v.GetUnhealthyStatus(), ", "), 96),
	}

	for k, cv := range v.Children {
		addNodes(cv, &n.Children[k])
	}
}

func New(data *xplane.Resource) Model {
	nodes := []tree.Node{
		{Key: "root", Children: make([]tree.Node, 1)},
	}
	addNodes(data, &nodes[0])

	t := tree.New(nodes, []string{"OBJECT", "GROUP", "SYNCED", "SYNC LAST UPDATE", "READY", "READY LAST UPDATE", "MESSAGE"})
	t.OnYank = func(node *tree.Node) {
		//nolint // nothing can be done in case of error
		clipboard.WriteAll(node.Value)
	}

	return Model{
		tree: t,
		statusbar: statusbar.New(
			statusbar.WithInitialPath([]string{nodes[0].Key}),
		),
	}
}

type Model struct {
	tree      tree.Model
	statusbar statusbar.Model
	width     int
	height    int
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

		top, right, _, left := lipgloss.NewStyle().Padding(1).GetPadding()
		m.tree.SetSize(m.width-right-left, m.height-top)

		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "?":
			return m, nil
		}
	}

	var treeCmd, statusCmd tea.Cmd
	m.tree, treeCmd = m.tree.Update(msg)
	m.statusbar.SetPath(m.tree.Path())
	m.statusbar, statusCmd = m.statusbar.Update(msg)

	return m, tea.Batch(treeCmd, statusCmd)
}

func (m Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.NewStyle().Height(m.height-m.statusbar.GetHeight()).Render(m.tree.View()),
		m.statusbar.View(),
	)
}
