package tui

import "github.com/charmbracelet/lipgloss"

const (
	ColorBlack         = "0"
	ColorRed           = "1"
	ColorGreen         = "2"
	ColorYellow        = "3"
	ColorBlue          = "4"
	ColorMagenta       = "5"
	ColorCyan          = "6"
	ColorWhite         = "7"
	ColorBrightBlack   = "8"
	ColorBrightRed     = "9"
	ColorBrightGreen   = "10"
	ColorBrightYellow  = "11"
	ColorBrightBlue    = "12"
	ColorBrightMagenta = "13"
	ColorBrightCyan    = "14"
	ColorBrightWhite   = "15"

	ColorAlert = lipgloss.Color(ColorRed)
	ColorWarn  = lipgloss.Color(ColorYellow)
	ColorLight = lipgloss.Color(ColorWhite)
	ColorDark  = lipgloss.Color(ColorBlack)

	ColorBackground = ColorDark
	ColorForeground = ColorLight
	ColorHighlight  = ColorBrightBlack

	DateFormat = "02 Jan 06 15:04"
)
