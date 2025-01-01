package tree

import (
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/table"
	"github.com/charmbracelet/bubbles/key"

	tea "github.com/charmbracelet/bubbletea"
)

type EventResize struct {
	Width  int
	Height int
}

type EventUpdateNodes struct {
	Nodes []*Node
}

func (m *Model) onNodesUpdate(msg EventUpdateNodes) tea.Cmd {
	m.nodes = msg.Nodes

	count := 0 // This is used to keep track of the index of the node we are on (important because we are using a recursive function)
	rows := []table.Row{}
	m.renderTree(&rows, m.nodes, []string{}, 0, &count)
	m.table.SetRows(rows)
	m.table.Focus()

	return nil
}

func (m *Model) onResize(msg tea.WindowSizeMsg) tea.Cmd {
	m.setSize(msg.Width, msg.Height)
	m.table.SetWidth(msg.Width)
	m.table.SetHeight(msg.Height)
	return nil
}

func (m *Model) onNavUp() {
	m.cursor--
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.onSelectionChange(m.nodesByCursor[m.cursor])
}

func (m *Model) onNavDown() {
	m.cursor++
	if m.cursor >= m.numberOfNodes() {
		m.cursor = m.numberOfNodes() - 1
	}
	m.onSelectionChange(m.nodesByCursor[m.cursor])
}

func (m *Model) onSelectionChange(node *Node) {
	if m.OnSelectionChange == nil {
		return
	}
	m.OnSelectionChange(node)
}

func (m *Model) onKey(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.KeyMap.Up):
		m.onNavUp()
	case key.Matches(msg, m.KeyMap.Down):
		m.onNavDown()
	case key.Matches(msg, m.KeyMap.ShowFullHelp):
		fallthrough
	case key.Matches(msg, m.KeyMap.CloseFullHelp):
		m.Help.ShowAll = !m.Help.ShowAll
	}
	return nil
}
