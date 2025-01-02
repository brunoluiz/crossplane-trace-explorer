package explorer

import (
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/explorer/viewer"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/tree"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func (m *Model) onLoad(data *xplane.Resource) tea.Cmd {
	nodes := []*tree.Node{
		{Key: "root", Children: make([]*tree.Node, 1)},
	}
	resByNode := map[*tree.Node]*xplane.Resource{}
	addNodes(data, nodes[0], resByNode)

	m.tree.Update(tree.EventUpdateNodes{Nodes: nodes})
	m.resByNode = resByNode

	return nil
}

func (m *Model) onResize(msg tea.WindowSizeMsg) tea.Cmd {
	m.width = msg.Width
	m.height = msg.Height

	top, right, _, left := lipgloss.NewStyle().Padding(1).GetPadding()
	m.tree.Update(tea.WindowSizeMsg{Width: m.width - right - left, Height: m.height - top})
	m.statusbar.Update(msg)
	m.viewer.Update(msg)

	return nil
}

func (m *Model) onKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c", "ctlr+d":
		return tea.Interrupt
	case "y":
		//nolint // ignore errors
		clipboard.WriteAll(m.tree.Current().Value)
	case "enter", "d":
		v := m.resByNode[m.tree.Current()]

		m.viewer.Update(viewer.EventSetup{Trace: v})
		m.pane = PaneSummary
	case "q", "esc":
		if m.pane == PaneTree {
			return tea.Interrupt
		} else {
			m.pane = PaneTree
		}
	}

	return nil
}

func addNodes(v *xplane.Resource, n *tree.Node, resByNode map[*tree.Node]*xplane.Resource) {
	name := fmt.Sprintf("%s/%s", v.Unstructured.GetKind(), v.Unstructured.GetName())
	resStatus := xplane.GetResourceStatus(v, name)
	group := v.Unstructured.GetObjectKind().GroupVersionKind().Group

	n.Key = name
	n.Value = fmt.Sprintf("%s.%s/%s", v.Unstructured.GetKind(), group, v.Unstructured.GetName())
	n.Children = make([]*tree.Node, len(v.Children))

	if resStatus.Status != "" {
		n.Color = lipgloss.ANSIColor(ansi.Red)
	}

	if v.Unstructured.GetAnnotations()["crossplane.io/paused"] == "true" {
		n.Key += " (paused)"
		n.Color = lipgloss.ANSIColor(ansi.Yellow)
	}

	n.Details = map[string]string{
		HeaderKeyGroup:      group,
		HeaderKeySynced:     resStatus.Synced,
		HeaderKeySyncedLast: resStatus.SyncedLastTransition.Format(time.RFC822),
		HeaderKeyReady:      resStatus.Ready,
		HeaderKeyReadyLast:  resStatus.ReadyLastTransition.Format(time.RFC822),
		HeaderKeyStatus:     resStatus.Status,
	}

	resByNode[n] = v

	for k, cv := range v.Children {
		n.Children[k] = &tree.Node{}
		addNodes(cv, n.Children[k], resByNode)
	}
}
