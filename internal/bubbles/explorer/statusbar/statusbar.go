package statusbar

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakenelf/teacup/statusbar"
)

type Model struct {
	statusbar      statusbar.Model
	path           []string
	pathSeparator  string
	rootSymbol     string
	primaryColor   statusbar.ColorConfig
	secondaryColor statusbar.ColorConfig
	neutralColor   statusbar.ColorConfig
}

type config struct {
	path           []string
	pathSeparator  string
	rootSymbol     string
	primaryColor   statusbar.ColorConfig
	secondaryColor statusbar.ColorConfig
	neutralColor   statusbar.ColorConfig
}

type WithOpt func(*config)

func WithPrimaryStatusColor(cl statusbar.ColorConfig) func(c *config) {
	return func(c *config) { c.primaryColor = cl }
}

func WithSecondaryStatusColor(cl statusbar.ColorConfig) func(c *config) {
	return func(c *config) { c.secondaryColor = cl }
}

func WithNeutralStatusColor(cl statusbar.ColorConfig) func(c *config) {
	return func(c *config) { c.neutralColor = cl }
}

func WithPathSeparator(p string) func(c *config) {
	return func(c *config) { c.pathSeparator = p }
}

func WithInitialPath(p []string) func(c *config) {
	return func(c *config) { c.path = p }
}

func New(opts ...WithOpt) Model {
	cfg := &config{
		path:          []string{},
		pathSeparator: "\ueab6 ",
		primaryColor: statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Dark: "#ffffff", Light: "#ffffff"},
			Background: lipgloss.AdaptiveColor{Light: "#F25D94", Dark: "#F25D94"},
		},
		secondaryColor: statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
			Background: lipgloss.AdaptiveColor{Light: "#6124DF", Dark: "#6124DF"},
		},
		neutralColor: statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
			Background: lipgloss.AdaptiveColor{Light: "#3c3836", Dark: "#3c3836"},
		},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	s := statusbar.New(
		cfg.primaryColor,
		cfg.neutralColor,
		cfg.neutralColor,
		cfg.neutralColor,
	)
	s.FirstColumn = "$"
	s.SecondColumn = strings.Join(cfg.path, cfg.pathSeparator)
	return Model{
		path:           cfg.path,
		pathSeparator:  cfg.pathSeparator,
		rootSymbol:     cfg.rootSymbol,
		statusbar:      s,
		primaryColor:   cfg.primaryColor,
		secondaryColor: cfg.secondaryColor,
		neutralColor:   cfg.neutralColor,
	}
}

func (m Model) Init() tea.Cmd   { return nil }
func (m *Model) SetSize(w int)  { m.statusbar.SetSize(w) }
func (m *Model) GetHeight() int { return statusbar.Height }
func (m *Model) View() string   { return m.statusbar.View() }

func (m *Model) SetPath(path []string) {
	m.path = path
	m.statusbar.SecondColumn = strings.Join(m.path, m.pathSeparator)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	m.statusbar.SecondColumn = strings.Join(m.path, m.pathSeparator)
	m.statusbar.FourthColumn = ""
	m.statusbar.FourthColumnColors = m.neutralColor

	//nolint // let me use my switches
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			m.statusbar.FourthColumn = "yanked"
			m.statusbar.FourthColumnColors = m.secondaryColor
		}
	}

	return m, nil
}
