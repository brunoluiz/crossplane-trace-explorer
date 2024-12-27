package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/brunoluiz/crossplane-trace-explorer/internal/bubbles/tree"
	"github.com/brunoluiz/crossplane-trace-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/mistakenelf/teacup/statusbar"
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
	group := v.Unstructured.GetObjectKind().GroupVersionKind().Group
	ready := v.GetCondition(v1.TypeReady)
	synced := v.GetCondition(v1.TypeSynced)

	n.Object = fmt.Sprintf("%s/%s", v.Unstructured.GetKind(), v.Unstructured.GetName())
	n.Group = group
	n.Children = make([]tree.Node, len(v.Children))
	n.Ready = tree.State{
		Status:             string(ready.Status),
		LastTransitionTime: ready.LastTransitionTime.Time,
	}
	n.Synced = tree.State{
		Status:             string(synced.Status),
		LastTransitionTime: synced.LastTransitionTime.Time,
	}
	n.Message = strings.Join(v.GetUnhealthyStatus(), ", ")

	for k, cv := range v.Children {
		addNodes(cv, &n.Children[k])
	}
}

func initialModel(data *xplane.Resource) model {
	nodes := []tree.Node{
		{
			Object:   "root",
			Children: make([]tree.Node, 1),
		},
	}
	addNodes(data, &nodes[0])

	return model{
		tree: tree.New(nodes, 0, 0),
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
	tree      tree.Model
	statusbar statusbar.Model
	width     int
	height    int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	statusRoot := "$"
	statusOp, statusOpColor := "", neutralStatusColor
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.statusbar.SetSize(m.width)
		m.statusbar.SetContent(statusRoot, "", statusOp, "")

		top, right, _, left := lipgloss.NewStyle().Padding(1).GetPadding()
		m.tree.SetSize(m.width-right-left, m.height-top)

		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			statusOp, statusOpColor = "yanked", statusbar.ColorConfig{
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
