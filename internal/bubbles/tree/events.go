package tree

import (
	"github.com/charmbracelet/bubbles/key"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmd = m.onResize(msg)
	case tea.KeyMsg:
		cmd = m.onKey(msg)
	}

	var tableCmd tea.Cmd
	m.table, tableCmd = m.table.Update(msg)

	return m, tea.Batch(cmd, tableCmd)
}

func (m *Model) onResize(msg tea.WindowSizeMsg) tea.Cmd {
	m.setSize(msg.Width, msg.Height)
	m.table.SetWidth(msg.Width)
	m.table.SetHeight(msg.Height)
	cols := m.table.Columns()
	w := 0

	// Adding `2` due to borders and all
	for _, col := range cols[:len(cols)-1] {
		w += col.Width + 2
	}
	w += 2
	cols[len(cols)-1].Width = (msg.Width - w)
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
