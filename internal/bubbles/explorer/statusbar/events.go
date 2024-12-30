package statusbar

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type EventUpdatePath struct {
	Path []string
}

func (m *Model) onPathUpdate(msg EventUpdatePath) tea.Cmd {
	m.path = msg.Path
	m.statusbar.SecondColumn = strings.Join(m.path, m.pathSeparator)
	return nil
}

func (m *Model) onResize(msg tea.WindowSizeMsg) tea.Cmd {
	m.statusbar.Width = msg.Width
	return nil
}

func (m *Model) onKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "y":
		m.statusbar.FourthColumn = "yanked"
		m.statusbar.FourthColumnColors = m.secondaryColor
	}
	return nil
}
