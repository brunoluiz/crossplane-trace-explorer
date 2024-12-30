package explorer

import (
	"github.com/atotto/clipboard"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/tree"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
