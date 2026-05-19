package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) viewModules() string {
	var sb strings.Builder

	sb.WriteString(viewTitleStyle("#00BFFF").Render("  :: Module Selection") + "\n")
	sb.WriteString(viewSeparator() + "\n\n")

	allModules := m.registry.Ordered()
	for i, mod := range allModules {
		cursor := "  "
		if i == m.cursor {
			cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF")).Bold(true).Render("▸ ")
		}

		enabled := m.config.IsModuleEnabled(mod.Name())
		available := mod.Available()

		var checkbox string
		if enabled {
			checkbox = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Render("[✓]")
		} else {
			checkbox = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render("[ ]")
		}

		nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

		name := nameStyle.Render(mod.Title())
		desc := descStyle.Render(mod.Description())

		if !available {
			name = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render(mod.Title())
			deps := strings.Join(mod.Dependencies(), ", ")
			desc = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Render(fmt.Sprintf("[missing: %s]", deps))
		}

		if i == m.cursor {
			name = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00BFFF")).Render(mod.Title())
		}

		sb.WriteString(fmt.Sprintf("%s%s %s  %s\n", cursor, checkbox, name, desc))
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
		"  [Space] Toggle  [a] All on  [d] All off  [Esc] Back"))

	return sb.String()
}

func (m Model) updateModules(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	allModules := m.registry.Ordered()
	maxIdx := len(allModules) - 1

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.cursor < maxIdx {
			m.cursor++
		}
	case key.Matches(msg, m.keys.Space):
		mod := allModules[m.cursor]
		current := m.config.IsModuleEnabled(mod.Name())
		m.config.SetModuleEnabled(mod.Name(), !current)
		m.unsaved = true
	case key.Matches(msg, m.keys.AllOn):
		for _, mod := range allModules {
			if mod.Available() {
				m.config.SetModuleEnabled(mod.Name(), true)
			}
		}
		m.unsaved = true
	case key.Matches(msg, m.keys.AllOff):
		for _, mod := range allModules {
			m.config.SetModuleEnabled(mod.Name(), false)
		}
		m.unsaved = true
	}

	return m, nil
}
