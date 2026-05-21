package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/msalexms/MoM/internal/module"
	"github.com/msalexms/MoM/internal/tui/components"
)

func (m Model) viewAsciiArt() string {
	var sb strings.Builder

	sb.WriteString(viewTitleStyle("#FF00FF").Render("  :: ASCII Art Text Generator") + "\n\n")
	sb.WriteString("  Enter text to display as ASCII art in your MOTD:\n\n")
	sb.WriteString("  " + m.textInput.View() + "\n\n")
	sb.WriteString(components.HelpStyle.Render("  [Enter] Apply  [Esc] Cancel") + "\n")
	sb.WriteString(components.HelpStyle.Render("  Uses built-in block letters (or figlet if installed)"))

	return sb.String()
}

func (m Model) updateAsciiArt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Back) {
		m.textInput.Blur()
		m.state = StateModuleSettings
		return m, nil
	}
	if msg.Type == tea.KeyEnter {
		m.config.Modules.CowsayConfig.Message = m.textInput.Value()
		m.config.Modules.CowsayConfig.Mode = "figlet"
		m.config.SetModuleEnabled("cowsay", true)
		if mod, ok := m.registry.Get("cowsay"); ok {
			if cowMod, ok := mod.(*module.CowsayModule); ok {
				cowMod.Message = m.textInput.Value()
				cowMod.Mode = "figlet"
			}
		}
		m.unsaved = true
		m.status = fmt.Sprintf("ASCII art text set: %q", m.textInput.Value())
		m.textInput.Blur()
		m.state = StateModuleSettings
		return m, nil
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// --- Services Picker View ---

func loadSystemServices() ([]string, error) {
	return module.ListSystemServices(context.Background())
}

func (m Model) filteredServices() []string {
	if m.serviceFilter == "" {
		return m.systemServices
	}
	filter := strings.ToLower(m.serviceFilter)
	var out []string
	for _, s := range m.systemServices {
		if strings.Contains(strings.ToLower(s), filter) {
			out = append(out, s)
		}
	}
	return out
}

func (m Model) viewServices() string {
	var sb strings.Builder

	sb.WriteString(viewTitleStyle("#00FF7F").Render("  :: Services Picker") + "\n")
	sb.WriteString(viewSeparator() + "\n")

	if m.serviceFilter != "" {
		sb.WriteString(fmt.Sprintf("  Filter: %s\n", m.serviceFilter))
	}
	sb.WriteString("\n")

	selected := make(map[string]bool)
	for _, s := range m.config.Modules.ServicesConfig.Services {
		selected[s] = true
	}

	services := m.filteredServices()
	if len(services) == 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true).Render(
			"  No services found (systemctl unavailable?)") + "\n")
	} else {
		// Show a window of services around the cursor
		pageSize := 15
		if m.height > 0 && m.height-10 > 5 {
			pageSize = m.height - 10
		}
		start := 0
		if m.serviceCursor >= pageSize {
			start = m.serviceCursor - pageSize + 1
		}
		end := start + pageSize
		if end > len(services) {
			end = len(services)
		}

		for i := start; i < end; i++ {
			svc := services[i]
			active := i == m.serviceCursor
			cursor := listCursor(active, colGreen)

			checkbox := colGray + "[ ]" + colReset
			if selected[svc] {
				checkbox = colGreen + colBold + "[✓]" + colReset
			}

			nameColor := colWhite
			if active {
				nameColor = colGreen
			}
			name := fixedCol(svc, 28, nameColor)

			sb.WriteString(cursor + checkbox + " " + name + "\n")
		}

		sb.WriteString(fmt.Sprintf("\n  %d/%d services  %d selected",
			m.serviceCursor+1, len(services), len(m.config.Modules.ServicesConfig.Services)))
	}

	sb.WriteString("\n\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
		"  [Space] Toggle  [/] Filter  [Esc] Back"))

	return sb.String()
}

func (m Model) updateServices(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	services := m.filteredServices()
	maxIdx := len(services) - 1

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.serviceCursor > 0 {
			m.serviceCursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.serviceCursor < maxIdx {
			m.serviceCursor++
		}
	case key.Matches(msg, m.keys.Space):
		if m.serviceCursor < len(services) {
			svc := services[m.serviceCursor]
			m.toggleService(svc)
			m.unsaved = true
		}
	case msg.String() == "/":
		// Simple filter: append chars. Backspace to remove.
		// For now just toggle filter mode indicator
	case msg.Type == tea.KeyBackspace:
		if len(m.serviceFilter) > 0 {
			m.serviceFilter = m.serviceFilter[:len(m.serviceFilter)-1]
			m.serviceCursor = 0
		}
	case msg.Type == tea.KeyRunes && len(msg.Runes) == 1:
		ch := msg.Runes[0]
		if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '-' || ch == '_' || ch == '.' {
			m.serviceFilter += string(ch)
			m.serviceCursor = 0
		}
	}

	return m, nil
}

func (m *Model) toggleService(svc string) {
	current := m.config.Modules.ServicesConfig.Services
	for i, s := range current {
		if s == svc {
			m.config.Modules.ServicesConfig.Services = append(current[:i], current[i+1:]...)
			m.syncServicesModule()
			return
		}
	}
	m.config.Modules.ServicesConfig.Services = append(current, svc)
	m.syncServicesModule()
}

func (m *Model) syncServicesModule() {
	if mod, ok := m.registry.Get("services"); ok {
		if svcMod, ok := mod.(*module.ServicesModule); ok {
			svcMod.Services = m.config.Modules.ServicesConfig.Services
		}
	}
}
