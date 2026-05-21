package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) viewPreview() string {
	var sb strings.Builder

	sb.WriteString(viewTitleStyle("#00FF7F").Render("  :: MOTD Preview") + "\n")
	sb.WriteString(viewSeparator() + "\n\n")

	sb.WriteString(m.viewport.View() + "\n\n")

	pct := int(m.viewport.ScrollPercent() * 100)
	footer := fmt.Sprintf("  [↑/↓/j/k] Scroll  [Esc] Back  (%d%%)", pct)
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(footer))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, sb.String())
}

func (m Model) updatePreview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func trimTrailingSpaces(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " ")
	}
	return strings.Join(lines, "\n")
}
