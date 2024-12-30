package viewer

import (
	"fmt"
	"time"

	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/viewer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/samber/lo"
	k8sv1 "k8s.io/api/core/v1"
)

type Model struct {
	viewer viewer.Model

	mainStyle      lipgloss.Style
	identedStyle   lipgloss.Style
	okHealthStyle  lipgloss.Style
	badHealthStyle lipgloss.Style
	metadataStyle  lipgloss.Style
}

func New() *Model {
	return &Model{
		viewer:         viewer.New(),
		mainStyle:      lipgloss.NewStyle().UnsetBackground().UnsetForeground(),
		identedStyle:   lipgloss.NewStyle().MarginLeft(2),
		okHealthStyle:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.ANSIColor(ansi.Green)),
		badHealthStyle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.ANSIColor(ansi.Red)),
		metadataStyle:  lipgloss.NewStyle().Bold(true),
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	//nolint // use of switches is recommended
	switch msg := msg.(type) {
	case EventSetup:
		cmd = m.onSetup(msg)
	}

	var viewerCmd tea.Cmd
	m.viewer, cmd = m.viewer.Update(msg)

	return m, tea.Batch(cmd, viewerCmd)
}

func (m Model) Init() tea.Cmd { return nil }
func (m Model) View() string  { return m.viewer.View() }

func (m Model) renderHealth(name string, c xpv1.Condition) string {
	info := []string{}
	s := m.badHealthStyle
	n := lo.Capitalize(name)
	if c.Status == k8sv1.ConditionTrue {
		s = m.okHealthStyle
	}

	if c.Reason == "" {
		info = append(info, s.Render(fmt.Sprintf("%s: %s", n, c.Status)))
	} else {
		info = append(info, s.Render(fmt.Sprintf("%s: %s (%s)", n, c.Status, c.Reason)))
	}

	if c.Message != "" {
		info = append(info, m.identedStyle.Render(fmt.Sprintf("Message: %s", c.Message)))
	}

	info = append(info, m.identedStyle.Render(fmt.Sprintf("Last Transition Time: %s", c.LastTransitionTime.Format(time.RFC822))))
	return lipgloss.JoinVertical(lipgloss.Top, info...)
}

func (m Model) renderMetadata(annotations map[string]string) string {
	info := []string{}
	info = append(info, m.metadataStyle.Render("Annotations:"))
	for k, v := range annotations {
		info = append(info, m.identedStyle.Render(fmt.Sprintf("%s: \"%s\"", k, v)))
	}

	return lipgloss.JoinVertical(lipgloss.Top, info...)
}
