package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) viewHelp() string {
	var sb strings.Builder

	sb.WriteString(viewTitleStyle("#00BFFF").Render("  [?] Keyboard Shortcuts") + "\n")
	sb.WriteString(viewSeparator() + "\n\n")

	helpItems := []struct{ key, desc string }{
		{"↑/k, ↓/j", "Navigate up/down"},
		{"Enter", "Select / Confirm"},
		{"Space", "Toggle module on/off"},
		{"a", "Enable all available modules"},
		{"d", "Disable all modules"},
		{"s / Ctrl+S", "Save & Apply MOTD"},
		{"/", "Filter / Search"},
		{"Esc", "Go back / Cancel"},
		{"?", "Toggle this help screen"},
		{"q / Ctrl+C", "Quit"},
	}

	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00BFFF")).Width(14)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))

	for _, h := range helpItems {
		sb.WriteString(fmt.Sprintf("  %s %s\n", keyStyle.Render(h.key), descStyle.Render(h.desc)))
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  Press ? or Esc to close"))

	return sb.String()
}
