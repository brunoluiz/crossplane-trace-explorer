package tree

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/samber/lo"
)

const (
	bottomLeft string = " └──"
)

type State struct {
	LastTransitionTime time.Time
	Status             string
}

type Node struct {
	Object  string
	Group   string
	Ready   State
	Synced  State
	Message string

	Children []Node
	path     []string
}

func (n *Node) GetFullName() string {
	s := strings.Split(n.Object, "/")
	return fmt.Sprintf("%s.%s/%s", s[0], n.Group, s[1])
}

type Model struct {
	KeyMap KeyMap
	Styles Styles

	width         int
	height        int
	nodes         []Node
	nodesByCursor map[int]*Node
	cursor        int

	Help     help.Model
	showHelp bool

	AdditionalShortHelpKeys func() []key.Binding
}

func New(
	nodes []Node,
	width int,
	height int,
) Model {
	return Model{
		KeyMap: DefaultKeyMap(),
		Styles: defaultStyles(),

		width:         width,
		height:        height,
		nodes:         nodes,
		nodesByCursor: map[int]*Node{},

		showHelp: true,
		Help:     help.New(),
	}
}

func (m Model) Nodes() []Node {
	return m.nodes
}

func (m *Model) SetNodes(nodes []Node) {
	m.nodes = nodes
}

func (m *Model) NumberOfNodes() int {
	count := 0

	var countNodes func([]Node)
	countNodes = func(nodes []Node) {
		for _, node := range nodes {
			count++
			if node.Children != nil {
				countNodes(node.Children)
			}
		}
	}

	countNodes(m.nodes)

	return count
}

func (m Model) Width() int {
	return m.width
}

func (m Model) Height() int {
	return m.height
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) SetWidth(newWidth int) {
	m.SetSize(newWidth, m.height)
}

func (m *Model) SetHeight(newHeight int) {
	m.SetSize(m.width, newHeight)
}

func (m Model) Cursor() int {
	return m.cursor
}

func (m *Model) SetCursor(cursor int) {
	m.cursor = cursor
}

func (m *Model) SetShowHelp() bool {
	return m.showHelp
}

func (m *Model) NavUp() {
	m.cursor--

	if m.cursor < 0 {
		m.cursor = 0
		return
	}
}

func (m *Model) NavDown() {
	m.cursor++

	if m.cursor >= m.NumberOfNodes() {
		m.cursor = m.NumberOfNodes() - 1
		return
	}
}

func (m *Model) Yank() {
	clipboard.WriteAll(m.nodesByCursor[m.cursor].GetFullName())
}

func (m *Model) Path() []string {
	return m.nodesByCursor[m.cursor].path
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Up):
			m.NavUp()
		case key.Matches(msg, m.KeyMap.Down):
			m.NavDown()
		case key.Matches(msg, m.KeyMap.Yank):
			m.Yank()
		case key.Matches(msg, m.KeyMap.ShowFullHelp):
			fallthrough
		case key.Matches(msg, m.KeyMap.CloseFullHelp):
			m.Help.ShowAll = !m.Help.ShowAll
		}
	}

	return m, nil
}

func (m Model) View() string {
	availableHeight := m.height
	nodes := m.Nodes()

	var help string
	if m.showHelp {
		help = m.helpView()
		availableHeight -= lipgloss.Height(help)
	}

	t := table.New().
		Border(lipgloss.HiddenBorder()).
		BorderTop(false).
		BorderHeader(true).
		StyleFunc(func(row, col int) lipgloss.Style {
			return lipgloss.NewStyle().PaddingRight(2)
		}).
		Headers("OBJECT", "GROUP", "SYNCED", "SYNC LAST UPDATE", "READY", "READY LAST UPDATE", "MESSAGE")

	count := 0 // This is used to keep track of the index of the node we are on (important because we are using a recursive function)
	m.renderTree(t, m.nodes, []string{}, 0, &count)

	if len(nodes) == 0 {
		return "No data"
	}
	return lipgloss.JoinVertical(lipgloss.Left, lipgloss.NewStyle().Height(availableHeight).Render(t.Render()), help)
}

func (m *Model) renderTree(t *table.Table, remainingNodes []Node, path []string, indent int, count *int) {
	for _, node := range remainingNodes {
		// If we aren't at the root, we add the arrow shape to the string
		shape := ""
		if indent > 0 {
			shape = strings.Repeat(" ", (indent-1)) + m.Styles.Shapes.Render(bottomLeft) + " "
		}

		// Generate the correct index for the node
		idx := *count
		*count++

		// Format the string with fixed width for the value and description fields
		valueStr := ""

		// If we are at the cursor, we add the selected style to the string
		if m.cursor == idx {
			valueStr = m.Styles.Selected.Render(node.Object)
		} else {
			valueStr = m.Styles.Unselected.Render(node.Object)
		}

		t.Row(
			fmt.Sprintf("%s%s", shape, valueStr),
			node.Group,
			node.Synced.Status,
			node.Synced.LastTransitionTime.Format("02 Jan 06 15:04"),
			node.Ready.Status,
			node.Ready.LastTransitionTime.Format("02 Jan 06 15:04"),
			lo.Elipse(node.Message, 140),
		)
		m.nodesByCursor[idx] = &node
		node.path = append(path, node.Object)

		if node.Children != nil {
			m.renderTree(t, node.Children, node.path, indent+1, count)
		}
	}
}

func (m Model) helpView() string {
	return m.Styles.Help.Render(m.Help.View(m))
}

func (m Model) ShortHelp() []key.Binding {
	kb := []key.Binding{
		m.KeyMap.Up,
		m.KeyMap.Down,
		m.KeyMap.Yank,
	}

	if m.AdditionalShortHelpKeys != nil {
		kb = append(kb, m.AdditionalShortHelpKeys()...)
	}

	return append(kb,
		m.KeyMap.Quit,
	)
}

func (m Model) FullHelp() [][]key.Binding {
	kb := [][]key.Binding{{
		m.KeyMap.Up,
		m.KeyMap.Down,
		m.KeyMap.Yank,
	}}

	return append(kb,
		[]key.Binding{
			m.KeyMap.Quit,
			m.KeyMap.CloseFullHelp,
		})
}
