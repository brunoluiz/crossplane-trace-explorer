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
	"github.com/muesli/reflow/truncate"
	"github.com/samber/lo"
)

const (
	bottomLeft string = " └──"
)

type State struct {
	Status    string
	UpdatedAt time.Time
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
	// TODO: deal with errors
	clipboard.WriteAll(m.nodesByCursor[m.cursor].Object)
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
	var sections []string

	nodes := m.Nodes()

	var help string
	if m.showHelp {
		help = m.helpView()
		availableHeight -= lipgloss.Height(help)
	}

	count := 0 // This is used to keep track of the index of the node we are on (important because we are using a recursive function)
	sections = append(sections, lipgloss.NewStyle().Height(availableHeight).Render(m.renderTree(m.nodes, []string{}, 0, &count)), help)

	if len(nodes) == 0 {
		return "No data"
	}
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) renderTree(remainingNodes []Node, path []string, indent int, count *int) string {
	var b strings.Builder

	for _, node := range remainingNodes {
		var str string

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

		f := fmt.Sprintf("%s%-*s", shape, 100, valueStr)

		r := fmt.Sprintf("%s (%s)", node.Ready.Status, node.Ready.UpdatedAt.Format("02 Jan 06 15:04"))
		s := fmt.Sprintf("%s (%s)", node.Synced.Status, node.Synced.UpdatedAt.Format("02 Jan 06 15:04"))
		msg := lo.Elipse(node.Message, 50)
		desc := fmt.Sprintf("%30s%30s  %-50s", s, r, msg)
		str += fmt.Sprintf("%s%s\n", truncate.String(f, 60), desc)

		b.WriteString(str)
		m.nodesByCursor[idx] = &node
		node.path = append(path, node.Object)

		if node.Children != nil {
			childStr := m.renderTree(node.Children, node.path, indent+1, count)
			b.WriteString(childStr)
		}
	}

	return b.String()
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
	}}

	return append(kb,
		[]key.Binding{
			m.KeyMap.Quit,
			m.KeyMap.CloseFullHelp,
		})
}
