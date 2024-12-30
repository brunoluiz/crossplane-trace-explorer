package tree

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type KeyMap struct {
	Bottom      key.Binding
	Top         key.Binding
	SectionDown key.Binding
	SectionUp   key.Binding
	Down        key.Binding
	Up          key.Binding
	Quit        key.Binding

	Yank          key.Binding
	ShowFullHelp  key.Binding
	CloseFullHelp key.Binding
}

type Styles struct {
	Shapes     lipgloss.Style
	Selected   lipgloss.Style
	Unselected lipgloss.Style
	Help       lipgloss.Style
}

type State struct {
	LastTransitionTime time.Time
	Status             string
}

type ColorConfig struct {
	Foreground lipgloss.ANSIColor
	Background lipgloss.ANSIColor
}

type Node struct {
	Key     string
	Value   string
	Details map[string]string

	Selected   ColorConfig
	Unselected ColorConfig

	Children []*Node
	Path     []string
}

type Model struct {
	KeyMap KeyMap
	Styles Styles
	Help   help.Model

	OnSelectionChange func(node *Node)
	OnYank            func(node *Node)

	width         int
	height        int
	nodes         []*Node
	nodesByCursor map[int]*Node
	cursor        int
	headers       []string

	showHelp bool
}

func New(headers []string) *Model {
	return &Model{
		KeyMap: KeyMap{
			Bottom: key.NewBinding(
				key.WithKeys("bottom"),
				key.WithHelp("end", "bottom"),
			),
			Top: key.NewBinding(
				key.WithKeys("top"),
				key.WithHelp("home", "top"),
			),
			SectionDown: key.NewBinding(
				key.WithKeys("secdown"),
				key.WithHelp("secdown", "section down"),
			),
			SectionUp: key.NewBinding(
				key.WithKeys("secup"),
				key.WithHelp("secup", "section up"),
			),
			Down: key.NewBinding(
				key.WithKeys("down", "j"),
				key.WithHelp("↓/j", "down"),
			),
			Up: key.NewBinding(
				key.WithKeys("up", "k"),
				key.WithHelp("↑/k", "up"),
			),

			Yank: key.NewBinding(
				key.WithKeys("y"),
				key.WithHelp("y", "yank"),
			),
			ShowFullHelp: key.NewBinding(
				key.WithKeys("?"),
				key.WithHelp("?", "help"),
			),
			CloseFullHelp: key.NewBinding(
				key.WithKeys("?"),
				key.WithHelp("?", "close help"),
			),

			Quit: key.NewBinding(
				key.WithKeys("q", "esc"),
				key.WithHelp("q", "quit"),
			),
		},
		Styles: Styles{
			Shapes:     lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(lipgloss.Color("#999")),
			Selected:   lipgloss.NewStyle().Margin(0, 0, 0, 0).Background(lipgloss.AdaptiveColor{Light: "#333", Dark: "#eee"}).Foreground(lipgloss.AdaptiveColor{Light: "#eee", Dark: "#333"}),
			Unselected: lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(lipgloss.AdaptiveColor{Light: "#333", Dark: "#eee"}),
			Help:       lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(lipgloss.AdaptiveColor{Light: "#333", Dark: "#eee"}),
		},

		width:         0,
		height:        0,
		nodesByCursor: map[int]*Node{},
		headers:       headers,

		showHelp: true,
		Help:     help.New(),
	}
}

func (m Model) Nodes() []*Node { return m.nodes }
func (m Model) Width() int     { return m.width }
func (m Model) Height() int    { return m.height }
func (m Model) Cursor() int    { return m.cursor }
func (m Model) Current() *Node { return m.nodesByCursor[m.cursor] }
func (m Model) Path() []string {
	if m.nodesByCursor != nil && m.nodesByCursor[m.cursor] != nil {
		return m.nodesByCursor[m.cursor].Path
	}
	return []string{}
}

func (m *Model) SetNodes(nodes []*Node)    { m.nodes = nodes }
func (m *Model) SetSize(width, height int) { m.width = width; m.height = height }
func (m *Model) SetWidth(newWidth int)     { m.SetSize(newWidth, m.height) }
func (m *Model) SetHeight(newHeight int)   { m.SetSize(m.width, newHeight) }
func (m *Model) SetCursor(cursor int)      { m.cursor = cursor }
func (m *Model) SetShowHelp() bool         { return m.showHelp }

func (m *Model) onSelectionChange(node *Node) {
	if m.OnSelectionChange == nil {
		return
	}
	m.OnSelectionChange(node)
}

func (m *Model) numberOfNodes() int {
	count := 0

	var countNodes func([]*Node)
	countNodes = func(nodes []*Node) {
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

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	//nolint // I prefer switch statements for this
	switch msg := msg.(type) {
	case tea.KeyMsg:
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
		StyleFunc(func(_, _ int) lipgloss.Style {
			return lipgloss.NewStyle().PaddingRight(2)
		}).
		Headers(m.headers...)

	count := 0 // This is used to keep track of the index of the node we are on (important because we are using a recursive function)
	m.renderTree(t, m.nodes, []string{}, 0, &count)

	if len(nodes) == 0 {
		return "No data"
	}

	return lipgloss.JoinVertical(lipgloss.Left, lipgloss.NewStyle().Height(availableHeight).Render(t.Render()), help)
}

func (m *Model) renderTree(t *table.Table, remainingNodes []*Node, path []string, indent int, count *int) {
	const treeNodePrefix string = " └──"

	for _, node := range remainingNodes {
		// If we aren't at the root, we add the arrow shape to the string
		shape := ""
		if indent > 0 {
			shape = strings.Repeat(" ", (indent-1)) + m.Styles.Shapes.Render(treeNodePrefix) + " "
		}

		// Generate the correct index for the node
		idx := *count
		*count++

		// Format the string with fixed width for the value and description fields
		valueStr := ""
		unselectedStyle := lipgloss.NewStyle().Inherit(m.Styles.Unselected)
		if node.Unselected.Foreground != 0 {
			unselectedStyle = unselectedStyle.Foreground(node.Unselected.Foreground)
		}

		// If we are at the cursor, we add the selected style to the string
		if m.cursor == idx {
			s := lipgloss.NewStyle().Inherit(m.Styles.Selected)
			if node.Selected.Background != 0 {
				s = s.Foreground(node.Selected.Foreground).Background(node.Selected.Background)
			}
			valueStr = s.Render(node.Key)
		} else {
			valueStr = unselectedStyle.Render(node.Key)
		}

		cols := []string{fmt.Sprintf("%s%s", shape, valueStr)}
		for _, v := range m.headers[1:] {
			cols = append(cols, unselectedStyle.Render(node.Details[v]))
		}
		t.Row(cols...)
		m.nodesByCursor[idx] = node

		// Used to be able to trace back the path on the tree
		node.Path = path
		node.Path = append(node.Path, node.Key)

		if node.Children != nil {
			m.renderTree(t, node.Children, node.Path, indent+1, count)
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
