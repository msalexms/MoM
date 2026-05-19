package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ams/mom/internal/module/render"
	"github.com/ams/mom/internal/theme"
)

func (m Model) viewTheme() string {
	var sb strings.Builder

	sb.WriteString(viewTitleStyle("#FF00FF").Render("  :: Theme Selector") + "\n")
	sb.WriteString(viewSeparator() + "\n\n")

	themes := theme.All()
	currentID := m.config.ThemeID()

	for i, th := range themes {
		cursor := "  "
		if i == m.cursor {
			cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF00FF")).Bold(true).Render("▸ ")
		}

		nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

		if i == m.cursor {
			nameStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF00FF"))
		}

		active := ""
		if th.ID == currentID {
			active = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Render(" ●")
		}

		sb.WriteString(fmt.Sprintf("%s%-14s %s%s\n", cursor,
			nameStyle.Render(th.Name),
			descStyle.Render(th.Description),
			active))
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
		"  [Enter] Select  [Esc] Back"))

	return sb.String()
}

func (m Model) updateTheme(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	themes := theme.All()
	maxIdx := len(themes) - 1

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
		if m.cursor < len(themes) {
			selected := themes[m.cursor]
			m.config.Mode.Theme = selected.ID
			m.generator.RenderOpts = render.Options{
				Theme:   selected,
				Variant: render.Variant(m.config.GlobalVariant()),
			}
			m.unsaved = true
			m.status = fmt.Sprintf("Theme set to '%s'", selected.Name)
			m.state = StateDashboard
			m.cursor = 0
		}
	}

	return m, nil
}
