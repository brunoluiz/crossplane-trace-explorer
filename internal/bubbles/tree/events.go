package tree

import (
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
	return nil
}

func (m *Model) onResize(msg tea.WindowSizeMsg) tea.Cmd {
	m.setSize(msg.Width, msg.Height)
	return nil
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
