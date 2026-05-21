package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/msalexms/MoM/internal/module"
)

// viewModuleSettings renders the module settings hub — lists all configurable modules.
func (m Model) viewModuleSettings() string {
	var sb strings.Builder

	sb.WriteString(viewTitleStyle("#FFD700").Render("  :: Module Settings") + "\n")
	sb.WriteString(viewSeparator() + "\n\n")

	mods := m.configurableModules()
	if len(mods) == 0 {
		sb.WriteString(dimText("  No configurable modules found.") + "\n")
	} else {
		for i, mod := range mods {
			active := i == m.settingsCursor
			cursor := listCursor(active, colYellow)

			nameColor := colWhite
			if active {
				nameColor = colYellow
			}
			name := fixedCol(mod.Title(), 20, nameColor)

			cfg := mod.(module.Configurable)
			settingCount := len(cfg.Settings())
			desc := dimText(fmt.Sprintf("%d setting(s)", settingCount))

			sb.WriteString(cursor + name + " " + desc + "\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
		"  [Enter] Configure  [Esc] Back"))

	return sb.String()
}

// updateModuleSettings handles input in the module settings hub.
func (m Model) updateModuleSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	mods := m.configurableModules()
	maxIdx := len(mods) - 1

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.settingsCursor > 0 {
			m.settingsCursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.settingsCursor < maxIdx {
			m.settingsCursor++
		}
	case key.Matches(msg, m.keys.Enter):
		if m.settingsCursor <= maxIdx {
			mod := mods[m.settingsCursor]
			return m.enterModuleConfig(mod)
		}
	}
	return m, nil
}

// enterModuleConfig transitions to the appropriate settings view for a module.
func (m Model) enterModuleConfig(mod module.Module) (tea.Model, tea.Cmd) {
	switch mod.Name() {
	case "services":
		m.state = StateServices
		m.serviceCursor = 0
		m.serviceFilter = ""
		svcs, err := loadSystemServices()
		if err == nil {
			m.systemServices = svcs
		}
	case "git":
		m.state = StateGitPaths
		m.gitPathCursor = 0
		m.gitPathAdding = false
	case "cowsay":
		m.state = StateAsciiArt
		m.textInput.Focus()
		return m, m.textInput.Cursor.BlinkCmd()
	default:
		// Generic settings editor for modules with simple settings
		m.state = StateGenericSettings
		m.genericSettingsMod = mod
		m.genericSettingsCursor = 0
		m.genericSettingsEditing = false
	}
	return m, nil
}

// viewGenericSettings renders a generic settings editor for any configurable module.
func (m Model) viewGenericSettings() string {
	var sb strings.Builder

	mod := m.genericSettingsMod
	if mod == nil {
		return "  No module selected"
	}
	cfg, ok := mod.(module.Configurable)
	if !ok {
		return "  Module has no settings"
	}

	sb.WriteString(viewTitleStyle("#FFD700").Render(fmt.Sprintf("  :: %s — Settings", mod.Title())) + "\n")
	sb.WriteString(viewSeparator() + "\n\n")

	settings := cfg.Settings()
	for i, s := range settings {
		active := i == m.genericSettingsCursor && !m.genericSettingsEditing
		cursor := listCursor(active, colYellow)

		labelColor := colWhite
		if active {
			labelColor = colYellow
		}
		label := fixedCol(s.Label, 20, labelColor)

		value := m.getSettingValue(mod, s)
		valueStr := col(fmt.Sprintf("%v", value), colCyan)

		sb.WriteString(cursor + label + " " + valueStr + "\n")
		if s.Description != "" {
			sb.WriteString("      " + dimText(s.Description) + "\n")
		}
	}

	if m.genericSettingsEditing {
		sb.WriteString("\n  " + m.genericSettingsInput.View() + "\n")
		sb.WriteString("\n")
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
			"  [Enter] Confirm  [Esc] Cancel"))
	} else {
		sb.WriteString("\n")
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
			"  [Enter] Edit  [Esc] Back"))
	}

	return sb.String()
}

// updateGenericSettings handles input in the generic settings editor.
func (m Model) updateGenericSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	mod := m.genericSettingsMod
	if mod == nil {
		m.state = StateModuleSettings
		return m, nil
	}
	cfg, ok := mod.(module.Configurable)
	if !ok {
		m.state = StateModuleSettings
		return m, nil
	}
	settings := cfg.Settings()

	// Editing mode
	if m.genericSettingsEditing {
		if key.Matches(msg, m.keys.Back) {
			m.genericSettingsEditing = false
			m.genericSettingsInput.Blur()
			return m, nil
		}
		if msg.Type == tea.KeyEnter {
			s := settings[m.genericSettingsCursor]
			val := m.genericSettingsInput.Value()
			m.applySettingValue(mod, s, val)
			m.genericSettingsEditing = false
			m.genericSettingsInput.Blur()
			m.unsaved = true
			return m, nil
		}
		var cmd tea.Cmd
		m.genericSettingsInput, cmd = m.genericSettingsInput.Update(msg)
		return m, cmd
	}

	// Navigation mode
	maxIdx := len(settings) - 1
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.genericSettingsCursor > 0 {
			m.genericSettingsCursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.genericSettingsCursor < maxIdx {
			m.genericSettingsCursor++
		}
	case key.Matches(msg, m.keys.Enter):
		if m.genericSettingsCursor <= maxIdx {
			s := settings[m.genericSettingsCursor]
			if s.Type == module.SettingBool {
				// Toggle directly
				m.toggleSettingBool(mod, s)
				m.unsaved = true
			} else if s.Type == module.SettingEnum {
				// Cycle through options
				m.cycleSettingEnum(mod, s)
				m.unsaved = true
			} else {
				// Open text input
				m.genericSettingsEditing = true
				m.genericSettingsInput.SetValue(fmt.Sprintf("%v", m.getSettingValue(mod, s)))
				m.genericSettingsInput.Focus()
				return m, m.genericSettingsInput.Cursor.BlinkCmd()
			}
		}
	}
	return m, nil
}

// configurableModules returns modules that implement Configurable and have settings.
func (m Model) configurableModules() []module.Module {
	var out []module.Module
	for _, mod := range m.registry.Ordered() {
		if cfg, ok := mod.(module.Configurable); ok {
			if len(cfg.Settings()) > 0 {
				out = append(out, mod)
			}
		}
	}
	return out
}

// getSettingValue reads the current value of a setting from config.
func (m Model) getSettingValue(mod module.Module, s module.SettingDef) any {
	switch mod.Name() {
	case "weather":
		switch s.Key {
		case "city":
			return m.config.Modules.WeatherConfig.City
		case "units":
			return m.config.Modules.WeatherConfig.Units
		}
	case "cowsay":
		switch s.Key {
		case "mode":
			return m.config.Modules.CowsayConfig.Mode
		case "message":
			return m.config.Modules.CowsayConfig.Message
		}
	case "resources":
		switch s.Key {
		case "show_temp":
			return m.config.Modules.ResourcesConfig.ShowTemp
		}
	case "containers":
		switch s.Key {
		case "runtime":
			return m.config.Modules.ContainersConfig.Runtime
		}
	case "updates":
		switch s.Key {
		case "include_aur":
			return m.config.Modules.UpdatesConfig.IncludeAUR
		}
	case "git":
		switch s.Key {
		case "paths":
			return strings.Join(m.config.Modules.GitConfig.Paths, ", ")
		case "max_repos":
			return m.config.Modules.GitConfig.MaxRepos
		}
	case "services":
		switch s.Key {
		case "services":
			return strings.Join(m.config.Modules.ServicesConfig.Services, ", ")
		}
	}
	if s.Default != nil {
		return s.Default
	}
	return ""
}

// applySettingValue writes a string value into the config for the given setting.
func (m *Model) applySettingValue(mod module.Module, s module.SettingDef, val string) {
	switch mod.Name() {
	case "weather":
		switch s.Key {
		case "city":
			m.config.Modules.WeatherConfig.City = val
			if wm, ok := mod.(*module.WeatherModule); ok {
				wm.City = val
			}
		case "units":
			m.config.Modules.WeatherConfig.Units = val
			if wm, ok := mod.(*module.WeatherModule); ok {
				wm.Units = val
			}
		}
	case "cowsay":
		switch s.Key {
		case "mode":
			m.config.Modules.CowsayConfig.Mode = val
			if cm, ok := mod.(*module.CowsayModule); ok {
				cm.Mode = val
			}
		case "message":
			m.config.Modules.CowsayConfig.Message = val
			if cm, ok := mod.(*module.CowsayModule); ok {
				cm.Message = val
			}
		}
	case "resources":
		switch s.Key {
		case "show_temp":
			m.config.Modules.ResourcesConfig.ShowTemp = val == "true"
			if rm, ok := mod.(*module.ResourcesModule); ok {
				rm.ShowTemp = val == "true"
			}
		}
	case "containers":
		switch s.Key {
		case "runtime":
			m.config.Modules.ContainersConfig.Runtime = val
			if cm, ok := mod.(*module.ContainersModule); ok {
				cm.Runtime = val
			}
		}
	case "updates":
		switch s.Key {
		case "include_aur":
			m.config.Modules.UpdatesConfig.IncludeAUR = val == "true"
		}
	case "git":
		switch s.Key {
		case "max_repos":
			n := 5
			fmt.Sscanf(val, "%d", &n)
			m.config.Modules.GitConfig.MaxRepos = n
			m.syncGitModule()
		}
	}
}

// toggleSettingBool flips a boolean setting.
func (m *Model) toggleSettingBool(mod module.Module, s module.SettingDef) {
	switch mod.Name() {
	case "resources":
		if s.Key == "show_temp" {
			m.config.Modules.ResourcesConfig.ShowTemp = !m.config.Modules.ResourcesConfig.ShowTemp
			if rm, ok := mod.(*module.ResourcesModule); ok {
				rm.ShowTemp = m.config.Modules.ResourcesConfig.ShowTemp
			}
		}
	case "updates":
		if s.Key == "include_aur" {
			m.config.Modules.UpdatesConfig.IncludeAUR = !m.config.Modules.UpdatesConfig.IncludeAUR
		}
	}
}

// cycleSettingEnum cycles to the next option in an enum setting.
func (m *Model) cycleSettingEnum(mod module.Module, s module.SettingDef) {
	current := fmt.Sprintf("%v", m.getSettingValue(mod, s))
	idx := 0
	for i, opt := range s.Options {
		if opt == current {
			idx = i
			break
		}
	}
	next := s.Options[(idx+1)%len(s.Options)]
	m.applySettingValue(mod, s, next)
}
