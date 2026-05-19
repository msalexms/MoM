package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) getModuleOrder() []string {
	if len(m.config.Mode.ModuleOrder) > 0 {
		return m.config.Mode.ModuleOrder
	}
	// Default order
	return []string{
		"logo", "system", "resources", "network", "weather",
		"containers", "services", "updates", "logins",
		"calendar", "quote", "cowsay",
	}
}

func (m Model) viewOrder() string {
	var sb strings.Builder

	sb.WriteString(viewTitleStyle("#00BFFF").Render("  :: Reorder Modules") + "\n")
	sb.WriteString(viewSeparator() + "\n\n")

	order := m.getModuleOrder()

	for i, name := range order {
		cursor := "  "
		if i == m.cursor {
			cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF")).Bold(true).Render("▸ ")
		}

		enabled := m.config.IsModuleEnabled(name)
		title := name
		if mod, ok := m.registry.Get(name); ok {
			title = mod.Title()
		}

		numStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
		nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		if !enabled {
			nameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
		}
		if i == m.cursor {
			nameStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00BFFF"))
		}

		indicator := " "
		if enabled {
			indicator = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Render("●")
		}

		sb.WriteString(fmt.Sprintf("%s%s %s %s\n", cursor,
			numStyle.Render(fmt.Sprintf("%2d.", i+1)),
			indicator,
			nameStyle.Render(title)))
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
		"  [K/↑] Move up  [J/↓] Move down  [Esc] Back"))

	return sb.String()
}

func (m Model) updateOrder(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	order := m.getModuleOrder()
	maxIdx := len(order) - 1

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.cursor < maxIdx {
			m.cursor++
		}
	case msg.String() == "K" || msg.String() == "shift+up":
		// Move current item up
		if m.cursor > 0 {
			order[m.cursor], order[m.cursor-1] = order[m.cursor-1], order[m.cursor]
			m.config.Mode.ModuleOrder = order
			m.cursor--
			m.unsaved = true
		}
	case msg.String() == "J" || msg.String() == "shift+down":
		// Move current item down
		if m.cursor < maxIdx {
			order[m.cursor], order[m.cursor+1] = order[m.cursor+1], order[m.cursor]
			m.config.Mode.ModuleOrder = order
			m.cursor++
			m.unsaved = true
		}
	}

	return m, nil
}
