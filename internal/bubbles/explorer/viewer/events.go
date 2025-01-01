package viewer

import (
	"fmt"
	"strings"

	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/viewer"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/goccy/go-yaml"
)

type EventSetup struct {
	Trace *xplane.Resource
}

func (m *Model) onSetup(msg EventSetup) (cmd tea.Cmd) {
	val, err := yaml.Marshal(msg.Trace.Unstructured.Object)
	if err != nil {
		panic(err)
	}

	m.viewer, cmd = m.viewer.Update(viewer.EventSetup{
		Title:     fmt.Sprintf("%s/%s", msg.Trace.Unstructured.GetKind(), msg.Trace.Unstructured.GetName()),
		SideTitle: msg.Trace.Unstructured.GetAPIVersion(),
		Content: m.mainStyle.Render(lipgloss.JoinVertical(
			lipgloss.Top,
			m.renderHealth("synced", msg.Trace.GetCondition(xpv1.TypeSynced)),
			m.renderHealth("ready", msg.Trace.GetCondition(xpv1.TypeReady)),
			m.renderMetadata(msg.Trace.Unstructured.GetAnnotations()),
			strings.Repeat("─", m.viewer.GetWidth()-4),
			string(val),
		)),
	})
	return cmd
}
