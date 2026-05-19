package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) viewModules() string {
	var sb strings.Builder

	sb.WriteString(viewTitleStyle("#00BFFF").Render("  :: Module Selection") + "\n")
	sb.WriteString(viewSeparator() + "\n\n")

	allModules := m.registry.Ordered()
	for i, mod := range allModules {
		active := i == m.cursor
		cursor := listCursor(active, colCyan)

		enabled := m.config.IsModuleEnabled(mod.Name())
		available := mod.Available()

		// Checkbox: [✓] or [ ] — always 3 chars
		checkbox := colGray + "[ ]" + colReset
		if enabled {
			checkbox = colGreen + colBold + "[✓]" + colReset
		}

		// Name: padded to 20 chars, then colored
		nameColor := colWhite
		if !available {
			nameColor = colGray
		}
		if active {
			nameColor = colCyan
		}
		name := fixedCol(mod.Title(), 20, nameColor)

		// Description
		var desc string
		if !available {
			deps := strings.Join(mod.Dependencies(), ", ")
			desc = colRed + "[needs: " + deps + "]" + colReset
		} else {
			desc = dimText(mod.Description())
		}

		sb.WriteString(cursor + checkbox + " " + name + " " + desc + "\n")
	}

	sb.WriteString("\n  " + dimText("[Space] Toggle  [a] All on  [d] All off  [?] Help  [Esc] Back"))

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
	case key.Matches(msg, m.keys.Help):
		// Show module help
		m.state = StateModuleHelp
		return m, nil
	}

	return m, nil
}
