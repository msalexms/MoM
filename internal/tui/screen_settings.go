package tui

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/msalexms/MoM/internal/config"
	"github.com/msalexms/MoM/internal/module"
	c "github.com/msalexms/MoM/internal/tui/components"
)

// displayValue renders a setting's current value for lists and help.
func displayValue(s module.Setting, cfg *config.Config) string {
	v := s.Get(cfg)
	switch s.Type {
	case module.SettingBool:
		if v == "true" {
			return T("yes")
		}
		return T("no")
	case module.SettingList:
		n := len(s.Values(cfg))
		if n == 0 {
			return T("(default)")
		}
		return fmt.Sprintf(T("%d: %s"), n, v)
	}
	if v == "" {
		return T("(empty)")
	}
	return v
}

// --- Hub: modules with settings ------------------------------------------------

func (m Model) configurableModules() []module.Module {
	var out []module.Module
	for _, mod := range m.reg.Ordered() {
		if _, ok := mod.(module.Configurable); ok {
			out = append(out, mod)
		}
	}
	return out
}

func (m Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	mods := m.configurableModules()
	if m.moveCursor(msg, len(mods)) {
		return m, nil
	}
	if key.Matches(msg, m.keys.Enter) && len(mods) > 0 {
		m.settingsMod = mods[m.cursor()]
		m.screen = scrModuleSettings
		m.cursors[scrModuleSettings] = 0
	}
	return m, nil
}

func (m Model) viewSettings() string {
	var sb strings.Builder
	sb.WriteString(c.Muted.Render(T("Modules with their own options. Changes take effect when you install (s).")) + "\n\n")
	for i, mod := range m.configurableModules() {
		active := i == m.cursor()
		enabled := m.cfg.IsModuleEnabled(mod.Name())
		var parts []string
		for _, s := range mod.(module.Configurable).Settings() {
			parts = append(parts, T(s.Label)+": "+displayValue(s, m.cfg))
		}
		sb.WriteString(c.Cursor(active) + c.Dot(enabled) + " " + c.Pad(c.Item(T(mod.Title()), active, true), 22) +
			c.Muted.Render(c.Truncate(strings.Join(parts, " · "), max(10, m.contentWidth()-28))) + "\n")
	}
	return sb.String()
}

// --- Settings of one module --------------------------------------------------

func (m Model) moduleSettings() []module.Setting {
	if cm, ok := m.settingsMod.(module.Configurable); ok {
		return cm.Settings()
	}
	return nil
}

func (m Model) updateModuleSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	settings := m.moduleSettings()
	if m.moveCursor(msg, len(settings)) || len(settings) == 0 {
		return m, nil
	}
	if !key.Matches(msg, m.keys.Enter) && !key.Matches(msg, m.keys.Space) {
		return m, nil
	}

	s := settings[m.cursor()]
	switch s.Type {
	case module.SettingBool, module.SettingEnum:
		if err := s.Set(m.cfg, s.Next(m.cfg)); err != nil {
			m.fail(err)
			return m, nil
		}
		m.changed()
	case module.SettingList:
		m.listSetting = s
		if m.settingsMod.Name() == "services" {
			m.screen = scrServices
			m.svcFilter = ""
			m.cursors[scrServices] = 0
			svcs, err := module.ListSystemServices(context.Background())
			if err != nil {
				m.fail(fmt.Errorf(T("could not list the services: %w"), err))
			}
			m.services = svcs
			return m, nil
		}
		m.screen = scrList
		m.cursors[scrList] = 0
	default:
		return m.startInput(inputSetting, s.Get(m.cfg), T(s.Label))
	}
	return m, nil
}

func (m Model) viewModuleSettings() string {
	settings := m.moduleSettings()
	var sb strings.Builder
	if !m.cfg.IsModuleEnabled(m.settingsMod.Name()) {
		sb.WriteString(c.Yellow.Render(T("This module is disabled: enable it in Modules to see it in the MOTD.")) + "\n\n")
	}
	for i, s := range settings {
		active := i == m.cursor()
		value := c.Text2.Render(displayValue(s, m.cfg))
		if m.purpose == inputSetting && active {
			value = m.input.View()
		}
		sb.WriteString(c.Cursor(active) + c.Pad(c.Item(T(s.Label), active, true), 20) + value + "\n")
	}
	if len(settings) > 0 {
		s := settings[m.cursor()]
		sb.WriteString("\n" + c.Muted.Render(T(s.Description)) + "\n")
		switch s.Type {
		case module.SettingEnum:
			sb.WriteString(c.Muted.Render(T("Options: ")+strings.Join(s.Options, ", ")+T(" — Enter moves to the next one.")) + "\n")
		case module.SettingBool:
			sb.WriteString(c.Muted.Render(T("Enter turns it on or off.")) + "\n")
		case module.SettingList:
			sb.WriteString(c.Muted.Render(T("Enter opens the list.")) + "\n")
		}
	}
	return sb.String()
}

// --- List editor -------------------------------------------------------------

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := m.listSetting.Values(m.cfg)
	if m.moveCursor(msg, len(items)) {
		return m, nil
	}
	switch {
	case key.Matches(msg, m.keys.Add):
		return m.startInput(inputListItem, "", T("new item"))
	case key.Matches(msg, m.keys.Remove) && len(items) > 0:
		i := m.cursor()
		items = slices.Delete(items, i, i+1)
		if err := m.listSetting.Set(m.cfg, strings.Join(items, ",")); err != nil {
			m.fail(err)
			return m, nil
		}
		m.setCursor(min(i, max(0, len(items)-1)))
		m.changed()
	}
	return m, nil
}

func (m Model) viewList() string {
	items := m.listSetting.Values(m.cfg)
	checkPaths := m.listSetting.Key == "paths"

	var sb strings.Builder
	sb.WriteString(c.Text2.Render(T(m.listSetting.Label)) + c.Muted.Render(" — "+T(m.listSetting.Description)) + "\n\n")
	if len(items) == 0 {
		sb.WriteString(c.Muted.Render(T("Empty list: the defaults are used. Press a to add.")) + "\n")
	}
	for i, it := range items {
		active := i == m.cursor() && m.purpose == inputNone
		row := c.Cursor(active) + c.Item(it, active, true)
		if checkPaths {
			if _, err := os.Stat(module.ExpandHomePath(it)); err == nil {
				row += "  " + c.OK.Render("✓")
			} else {
				row += "  " + c.Bad.Render(T("✗ does not exist"))
			}
		}
		sb.WriteString(row + "\n")
	}
	if m.purpose == inputListItem {
		sb.WriteString("\n" + m.input.View() + "\n")
	}
	return sb.String()
}

// --- Service picker ----------------------------------------------------------

func (m Model) filteredServices() []string {
	if m.svcFilter == "" {
		return m.services
	}
	f := strings.ToLower(m.svcFilter)
	var out []string
	for _, s := range m.services {
		if strings.Contains(strings.ToLower(s), f) {
			out = append(out, s)
		}
	}
	return out
}

// updateServices owns every key: letters feed the filter, so the global
// shortcuts (q, s, ?) are not available here.
func (m Model) updateServices(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	list := m.filteredServices()
	switch msg.Type {
	case tea.KeyUp:
		m.setCursor(max(0, m.cursor()-1))
		return m, nil
	case tea.KeyDown:
		m.setCursor(min(max(0, len(list)-1), m.cursor()+1))
		return m, nil
	case tea.KeyEsc:
		if m.svcFilter != "" {
			m.svcFilter = ""
			m.setCursor(0)
			return m, nil
		}
		return m.back()
	case tea.KeyBackspace:
		if r := []rune(m.svcFilter); len(r) > 0 {
			m.svcFilter = string(r[:len(r)-1])
			m.setCursor(0)
		}
		return m, nil
	case tea.KeySpace, tea.KeyEnter:
		if len(list) > 0 {
			m.toggleService(list[m.cursor()])
		}
		return m, nil
	case tea.KeyRunes:
		m.svcFilter += string(msg.Runes)
		m.setCursor(0)
	}
	return m, nil
}

func (m *Model) toggleService(svc string) {
	selected := m.listSetting.Values(m.cfg)
	if i := slices.Index(selected, svc); i >= 0 {
		selected = slices.Delete(selected, i, i+1)
	} else {
		selected = append(selected, svc)
	}
	if err := m.listSetting.Set(m.cfg, strings.Join(selected, ",")); err != nil {
		m.fail(err)
		return
	}
	m.changed()
}

func (m Model) viewServices() string {
	selected := m.listSetting.Values(m.cfg)
	list := m.filteredServices()

	var sb strings.Builder
	filter := c.Muted.Render(T("type to filter"))
	if m.svcFilter != "" {
		filter = c.Accent.Render(m.svcFilter) + c.Brand.Render("▏")
	}
	sb.WriteString(c.Muted.Render(T("Filter: ")) + filter + c.Muted.Render(fmt.Sprintf(T("    %d selected"), len(selected))) + "\n")
	if len(selected) == 0 {
		sb.WriteString(c.Muted.Render(T("With none selected, the usual ones are watched (ssh, nginx, docker…).")) + "\n")
	} else {
		sb.WriteString(c.Muted.Render(c.Truncate(strings.Join(selected, ", "), m.contentWidth())) + "\n")
	}
	sb.WriteString("\n")

	if len(list) == 0 {
		sb.WriteString(c.Muted.Render(T("No service matches (is systemd available?).")) + "\n")
		return sb.String()
	}
	start, end := c.Window(len(list), m.cursor(), m.listHeight(5))
	for i := start; i < end; i++ {
		active := i == m.cursor()
		sb.WriteString(c.Cursor(active) + c.Check(slices.Contains(selected, list[i])) + " " + c.Item(list[i], active, true) + "\n")
	}
	if info := c.ScrollInfo(start, end, len(list)); info != "" {
		sb.WriteString("\n" + info)
	}
	return sb.String()
}
