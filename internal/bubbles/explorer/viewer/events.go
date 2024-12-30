package viewer

import (
	"fmt"

	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/viewer"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

func (m *Model) onSetup(msg *xplane.Resource) (cmd tea.Cmd) {
	m.viewer, cmd = m.viewer.Update(viewer.EventSetup{
		Title:     fmt.Sprintf("%s/%s", msg.Unstructured.GetKind(), msg.Unstructured.GetName()),
		SideTitle: msg.Unstructured.GetAPIVersion(),
		Content: m.mainStyle.Render(lipgloss.JoinVertical(
			lipgloss.Top,
			m.renderHealth("synced", msg.GetCondition(xpv1.TypeSynced)),
			m.renderHealth("ready", msg.GetCondition(xpv1.TypeReady)),
			m.renderMetadata(msg.Unstructured.GetAnnotations()),
		)),
	})
	return cmd
}
