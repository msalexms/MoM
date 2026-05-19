package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) viewPreview() string {
	var sb strings.Builder

	sb.WriteString(viewTitleStyle("#00FF7F").Render("  :: MOTD Preview") + "\n")
	sb.WriteString(viewSeparator() + "\n\n")

	if m.previewText == "" {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true).Render(
			"  (No output generated)") + "\n")
	} else {
		previewBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#333333")).
			Padding(1, 2)
		sb.WriteString(previewBox.Render(m.previewText) + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  [Esc] Back"))

	return sb.String()
}

func (m Model) updatePreview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}
