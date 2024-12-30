package viewer

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type Model struct {
	title     string
	sideTitle string
	content   string

	cmdQuit        tea.Cmd
	titleStyle     lipgloss.Style
	sideTitleStyle lipgloss.Style
	viewportStyle  lipgloss.Style
	footerStyle    lipgloss.Style

	// You generally won't need this unless you're processing stuff with
	// complicated ANSI escape sequences. Turn it on if you notice flickering.
	//
	// Also keep in mind that high performance rendering only works for programs
	// that use the full size of the terminal. We're enabling that below with
	// tea.EnterAltScreen().
	useHighPerformanceRenderer bool

	ready    bool
	viewport viewport.Model
}

type WithOpt func(*Model)

func WithQuitCmd(c tea.Cmd) func(m *Model) {
	return func(m *Model) {
		m.cmdQuit = c
	}
}

func WithHighPerformanceRenderer(enabled bool) func(m *Model) {
	return func(m *Model) {
		m.useHighPerformanceRenderer = enabled
	}
}

func New(opts ...WithOpt) Model {
	m := Model{
		cmdQuit:                    nil,
		useHighPerformanceRenderer: false,
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.ANSIColor(ansi.BrightBlack)).
			Foreground(lipgloss.ANSIColor(ansi.White)).
			Padding(0, 1, 0, 1).
			Margin(1, 0, 0, 1),
		sideTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.ANSIColor(ansi.Blue)).
			Foreground(lipgloss.ANSIColor(ansi.White)).
			Padding(0, 1, 0, 1).
			Margin(1, 0, 0, 1),
		viewportStyle: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, true, true, true).
			Margin(0, 1, 0, 1).
			Padding(0, 1, 0, 1),
		footerStyle: lipgloss.NewStyle().
			Padding(0, 1, 0, 1),
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			return m, m.cmdQuit
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.Style = m.viewportStyle
			m.viewport.YPosition = headerHeight
			m.viewport.HighPerformanceRendering = m.useHighPerformanceRenderer
			m.viewport.SetContent(m.content)
			m.ready = true

			// This is only necessary for high performance rendering, which in
			// most cases you won't need.
			//
			// Render the viewport one line below the header.
			m.viewport.YPosition = headerHeight + 1
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

		if m.useHighPerformanceRenderer {
			// Render (or re-render) the whole viewport. Necessary both to
			// initialize the viewport and when the window is resized.
			//
			// This is needed for high-performance rendering only.
			cmds = append(cmds, viewport.Sync(m.viewport))
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m Model) headerView() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.titleStyle.Render(m.title),
		m.sideTitleStyle.Render(m.sideTitle),
	)
}

func (m Model) footerView() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Right,
		m.footerStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100)),
	)
}

func (m Model) GetHeight() int { return m.viewport.Height }
func (m Model) GetWidth() int  { return m.viewport.Width }

func (m *Model) SetTitle(s string)      { m.title = s }
func (m *Model) SetSideTitle(s string)  { m.sideTitle = s }
func (m *Model) SetContent(s string)    { m.content = s; m.viewport.SetContent(s) }
func (m *Model) SetCmdQuit(cmd tea.Cmd) { m.cmdQuit = cmd }