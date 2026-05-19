package components

import "github.com/charmbracelet/lipgloss"

// Theme defines the application color scheme and styles.
var (
	// Colors
	ColorCyan    = lipgloss.Color("#00BFFF")
	ColorMagenta = lipgloss.Color("#FF00FF")
	ColorGreen   = lipgloss.Color("#00FF7F")
	ColorRed     = lipgloss.Color("#FF4444")
	ColorYellow  = lipgloss.Color("#FFD700")
	ColorGray    = lipgloss.Color("#666666")
	ColorWhite   = lipgloss.Color("#FFFFFF")
	ColorBlue    = lipgloss.Color("#4169E1")

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan).
			PaddingLeft(1).
			PaddingRight(1)

	HeadingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorMagenta)

	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorBlue)

	DisabledStyle = lipgloss.NewStyle().
			Foreground(ColorGray)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorGreen).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(ColorYellow)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorGray)

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorCyan)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorGray).
			PaddingLeft(1)

	MenuItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	ActiveMenuItemStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(ColorCyan).
				Bold(true)

	CheckboxChecked = lipgloss.NewStyle().
			Foreground(ColorGreen).
			Bold(true)

	CheckboxUnchecked = lipgloss.NewStyle().
				Foreground(ColorGray)

	UnsavedStyle = lipgloss.NewStyle().
			Foreground(ColorYellow).
			Bold(true)
)
