package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ams/mom/internal/template"
)

func (m Model) viewTemplates() string {
	var sb strings.Builder

	sb.WriteString(viewTitleStyle("#FF00FF").Render("  :: Apply Template") + "\n")
	sb.WriteString(viewSeparator() + "\n\n")

	for i, tmpl := range m.templates {
		active := i == m.cursor
		cursor := listCursor(active, colMagenta)

		nameColor := colWhite
		if active {
			nameColor = colMagenta
		}
		name := fixedCol(tmpl.Name, 14, nameColor)
		desc := dimText(tmpl.Description)

		sb.WriteString(cursor + name + " " + desc + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
		"  [Enter] Apply  [Esc] Back"))

	return sb.String()
}

func (m Model) updateTemplates(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxIdx := len(m.templates) - 1

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.cursor < maxIdx {
			m.cursor++
		}
	case key.Matches(msg, m.keys.Enter):
		if m.cursor < len(m.templates) {
			tmpl := m.templates[m.cursor]
			template.Apply(tmpl, m.config)
			m.unsaved = true
			m.status = fmt.Sprintf("Template '%s' applied!", tmpl.Name)
			m.state = StateDashboard
			m.cursor = 0
		}
	}

	return m, nil
}
