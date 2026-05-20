package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ams/mom/internal/config"
)

func (m Model) getModuleOrder() []string {
	// Start with the configured order (or default)
	base := m.config.Mode.ModuleOrder
	if len(base) == 0 {
		base = config.DefaultModuleOrder()
	}

	// Ensure all registered modules are included (auto-add new ones at the end)
	seen := make(map[string]bool)
	for _, name := range base {
		seen[name] = true
	}
	for _, mod := range m.registry.Ordered() {
		if !seen[mod.Name()] {
			base = append(base, mod.Name())
		}
	}
	return base
}

func (m Model) viewOrder() string {
	var sb strings.Builder

	sb.WriteString(viewTitleStyle("#00BFFF").Render("  :: Reorder Modules") + "\n")
	sb.WriteString(viewSeparator() + "\n\n")

	order := m.getModuleOrder()

	for i, name := range order {
		active := i == m.cursor
		cursor := listCursor(active, colCyan)

		enabled := m.config.IsModuleEnabled(name)
		title := name
		if mod, ok := m.registry.Get(name); ok {
			title = mod.Title()
		}

		num := fixedCol(fmt.Sprintf("%2d.", i+1), 3, colGray)

		indicatorColor := colGray
		if enabled {
			indicatorColor = colGreen
		}
		if active {
			indicatorColor = colCyan
		}
		indicator := fixedCol("●", 1, indicatorColor)
		if !enabled {
			indicator = fixedCol("○", 1, colGray)
		}

		nameColor := colGray
		if enabled {
			nameColor = colWhite
		}
		if active {
			nameColor = colCyan
		}
		name := fixedCol(title, 20, nameColor)

		sb.WriteString(cursor + num + " " + indicator + " " + name + "\n")
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
