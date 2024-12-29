package viewer

import (
	"fmt"

	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/viewer"
	"github.com/brunoluiz/crossplane-explorer/internal/tui"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/mistakenelf/teacup/statusbar"
	"github.com/samber/lo"
	k8sv1 "k8s.io/api/core/v1"
)

var (
	mainStyle = lipgloss.NewStyle().UnsetBackground().UnsetForeground()

	identedStyle = lipgloss.NewStyle().
			MarginLeft(2)

	okHealthStyle = lipgloss.NewStyle().
			Bold(true).Foreground(lipgloss.Color(tui.ColorGreen))

	badHealthStyle = lipgloss.NewStyle().
			Bold(true).Foreground(lipgloss.Color(tui.ColorRed))

	metadataStyle = lipgloss.NewStyle().
			Bold(true)
)

type Model struct {
	viewer viewer.Model
}

func New() Model {
	v := viewer.New()
	v.SetCmdQuit(nil)
	return Model{
		viewer: v,
	}
}

func (m *Model) Setup(v *xplane.Resource) {
	m.viewer.SetTitle(fmt.Sprintf("%s/%s", v.Unstructured.GetKind(), v.Unstructured.GetName()))
	m.viewer.SetSideTitle(v.Unstructured.GetAPIVersion())

	content := mainStyle.Render(lipgloss.JoinVertical(
		lipgloss.Top,
		renderHealth("synced", v.GetCondition(xpv1.TypeSynced)),
		renderHealth("ready", v.GetCondition(xpv1.TypeReady)),
		renderMetadata(v.Unstructured.GetAnnotations()),
	))

	m.viewer.SetContent(content)
}

func (m *Model) Update(msg tea.Msg) (_ Model, cmd tea.Cmd) {
	m.viewer, cmd = m.viewer.Update(msg)
	return *m, cmd
}

func (m *Model) GetHeight() int { return statusbar.Height }
func (m Model) Init() tea.Cmd   { return nil }
func (m Model) View() string    { return m.viewer.View() }

func renderHealth(name string, c xpv1.Condition) string {
	info := []string{}
	s := badHealthStyle
	n := lo.Capitalize(name)
	if c.Status == k8sv1.ConditionTrue {
		s = okHealthStyle
	}

	if c.Reason == "" {
		info = append(info, s.Render(fmt.Sprintf("%s: %s", n, c.Status)))
	} else {
		info = append(info, s.Render(fmt.Sprintf("%s: %s (%s)", n, c.Status, c.Reason)))
	}

	if c.Message != "" {
		info = append(info, identedStyle.Render(fmt.Sprintf("Message: %s", c.Message)))
	}

	info = append(info, identedStyle.Render(fmt.Sprintf("Last Transition Time: %s", c.LastTransitionTime.Format(tui.DateFormat))))
	return lipgloss.JoinVertical(lipgloss.Top, info...)
}

func renderMetadata(annotations map[string]string) string {
	info := []string{}
	info = append(info, metadataStyle.Render("Annotations:"))
	for k, v := range annotations {
		info = append(info, identedStyle.Render(fmt.Sprintf("%s: \"%s\"", k, v)))
	}

	return lipgloss.JoinVertical(lipgloss.Top, info...)
}
