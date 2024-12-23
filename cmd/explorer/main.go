package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/concordalabs/crossplane-trace-explorer/internal/bubbles/tree"
	"github.com/concordalabs/crossplane-trace-explorer/internal/xplane"
	"github.com/mistakenelf/teacup/statusbar"
	"golang.org/x/term"
)

var (
	styleDoc = lipgloss.NewStyle().Padding(1)
	data     *xplane.Resource
)

func main() {
	res, err := xplane.Parse(os.Stdin)
	if err != nil {
		fmt.Printf("Error while parsing Crossplane: %s\n", err)
		os.Exit(1)
	}

	_, err = tea.NewProgram(initialModel(res)).Run()
	if err != nil {
		os.Exit(1)
	}
}

func addNodes(v *xplane.Resource, n *tree.Node) {
	n.Value = fmt.Sprintf("%s/%s", v.Unstructured.GetKind(), v.Unstructured.GetName())
	n.Children = make([]tree.Node, len(v.Children))
	n.Desc = fmt.Sprintf("%30s%30s", v.Unstructured.GetAPIVersion(), v.Error)

	for k, cv := range v.Children {
		addNodes(cv, &n.Children[k])
	}
}

func initialModel(data *xplane.Resource) model {
	nodes := []tree.Node{
		{
			Value:    "root",
			Children: make([]tree.Node, 1),
		},
	}
	addNodes(data, &nodes[0])

	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		w = 80
		h = 24
	}
	top, right, bottom, left := styleDoc.GetPadding()
	w = w - left - right
	h = h - top - bottom

	return model{
		tree: tree.New(nodes, w, h),
		statusbar: statusbar.New(
			statusbar.ColorConfig{
				Foreground: lipgloss.AdaptiveColor{Dark: "#ffffff", Light: "#ffffff"},
				Background: lipgloss.AdaptiveColor{Light: "#F25D94", Dark: "#F25D94"},
			},
			neutralStatusColor,
			neutralStatusColor,
			neutralStatusColor,
		),
	}
}

var neutralStatusColor = statusbar.ColorConfig{
	Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
	Background: lipgloss.AdaptiveColor{Light: "#3c3836", Dark: "#3c3836"},
}

type model struct {
	height    int
	tree      tree.Model
	statusbar statusbar.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	statusRoot := "$"
	statusOp, statusOpColor := "", neutralStatusColor
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.statusbar.SetSize(msg.Width)
		m.statusbar.SetContent(statusRoot, "", statusOp, "")

		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			statusOp, statusOpColor = "copied value", statusbar.ColorConfig{
				Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
				Background: lipgloss.AdaptiveColor{Light: "#6124DF", Dark: "#6124DF"},
			}
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "?":
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.tree, cmd = m.tree.Update(msg)
	m.statusbar.SetContent(statusRoot, strings.Join(m.tree.Path(), " > "), "", statusOp)
	m.statusbar.FourthColumnColors = statusOpColor

	return m, cmd
}

func (m model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.NewStyle().Height(m.height-statusbar.Height).Render(m.tree.View()),
		m.statusbar.View(),
	)
}
