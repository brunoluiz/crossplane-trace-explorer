package explorer

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/tree"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/samber/lo"
	k8sv1 "k8s.io/api/core/v1"
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
		clipboard.WriteAll(m.tree.Current().Value)
	case "enter":
		v := m.resByNode[m.tree.Current()]

		m.viewer.Update(v)
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
		HeaderKeySyncedLast: synced.LastTransitionTime.Format(time.RFC822),
		HeaderKeyReady:      string(ready.Status),
		HeaderKeyReadyLast:  ready.LastTransitionTime.Format(time.RFC822),
		HeaderKeyMessage:    lo.Elipse(strings.Join(v.GetUnhealthyStatus(), ", "), 96),
	}

	resByNode[n] = v

	for k, cv := range v.Children {
		n.Children[k] = &tree.Node{}
		addNodes(cv, n.Children[k], resByNode)
	}
}
