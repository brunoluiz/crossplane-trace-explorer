package viewer

import (
	"fmt"
	"strings"
	"time"

	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/viewer"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/goccy/go-yaml"
	"github.com/samber/lo"
	k8sv1 "k8s.io/api/core/v1"
)

type Model struct {
	viewer viewer.Model

	styles Styles
}

func New() Model {
	return Model{
		viewer: viewer.New(),
		styles: DefaultStyles(),
	}
}

func (m Model) Init() tea.Cmd { return nil }
func (m Model) View() string  { return m.viewer.View() }

type ContentInput struct {
	Trace *xplane.Resource
}

func (m *Model) SetContent(msg ContentInput) {
	val, err := yaml.Marshal(msg.Trace.Unstructured.Object)
	if err != nil {
		panic(err)
	}

	hr := "────"
	if m.viewer.GetWidth() > 4 {
		hr = strings.Repeat("─", m.viewer.GetWidth()-4)
	}

	m.viewer.SetContent(viewer.ContentInput{
		Title:     fmt.Sprintf("%s/%s", msg.Trace.Unstructured.GetKind(), msg.Trace.Unstructured.GetName()),
		SideTitle: msg.Trace.Unstructured.GetAPIVersion(),
		Content: m.styles.Main.Render(lipgloss.JoinVertical(
			lipgloss.Top,
			m.renderHealth("synced", msg.Trace.GetCondition(xpv1.TypeSynced)),
			m.renderHealth("ready", msg.Trace.GetCondition(xpv1.TypeReady)),
			m.renderMetadata(msg.Trace.Unstructured.GetAnnotations()),
			hr,
			string(val),
		)),
	})
}

func (m Model) renderHealth(name string, c xpv1.Condition) string {
	info := []string{}
	s := m.styles.BadHealth
	n := lo.Capitalize(name)
	if c.Status == k8sv1.ConditionTrue {
		s = m.styles.OkHealth
	}

	if c.Reason == "" {
		info = append(info, s.Render(fmt.Sprintf("%s: %s", n, c.Status)))
	} else {
		info = append(info, s.Render(fmt.Sprintf("%s: %s (%s)", n, c.Status, c.Reason)))
	}

	if c.Message != "" {
		info = append(info, m.styles.Idented.Render(fmt.Sprintf("Message: %s", c.Message)))
	}

	info = append(info, m.styles.Idented.Render(fmt.Sprintf("Last Transition Time: %s", c.LastTransitionTime.Format(time.RFC822))))
	return lipgloss.JoinVertical(lipgloss.Top, info...)
}

func (m Model) renderMetadata(annotations map[string]string) string {
	info := []string{}
	info = append(info, m.styles.Metadata.Render("Annotations:"))
	for k, v := range annotations {
		info = append(info, m.styles.Idented.Render(fmt.Sprintf("%s: \"%s\"", k, v)))
	}

	return lipgloss.JoinVertical(lipgloss.Top, info...)
}
