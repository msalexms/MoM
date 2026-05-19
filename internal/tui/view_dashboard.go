package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ams/mom/internal/tui/components"
)

var dashboardItems = []struct {
	icon  string
	label string
}{
	{" + ", "Select Modules"},
	{" ↕ ", "Reorder Modules"},
	{" # ", "Apply Template"},
	{" T ", "Theme & Style"},
	{" S ", "Services Picker"},
	{" A ", "ASCII Art Text"},
	{" > ", "Preview MOTD"},
	{" ~ ", "Auto-Detect Modules"},
	{" ! ", "Full-Auto Setup"},
	{" W ", "Save & Apply"},
	{" P ", "Profiles"},
	{" R ", "Rollback"},
	{" Q ", "Quit"},
}

func (m Model) viewDashboard() string {
	var sb strings.Builder

	header := `
 ███╗   ███╗ ██████╗ ███╗   ███╗
 ████╗ ████║██╔═══██╗████╗ ████║
 ██╔████╔██║██║   ██║██╔████╔██║
 ██║╚██╔╝██║██║   ██║██║╚██╔╝██║
 ██║ ╚═╝ ██║╚██████╔╝██║ ╚═╝ ██║
 ╚═╝     ╚═╝ ╚═════╝ ╚═╝     ╚═╝`

	sb.WriteString(logoStyle.Render(header) + "\n")
	sb.WriteString(subtitleStyle.Render("  Message Of the Day Manager") + "\n\n")

	// Info box
	distroText := fmt.Sprintf(" %s  |  %s", m.distroInfo.Name, m.distroInfo.Family)
	enabled := m.config.EnabledModuleNames()
	themeID := m.config.ThemeID()
	variant := m.config.GlobalVariant()
	modulesText := fmt.Sprintf("  Modules: %s  Theme: %s  Style: %s",
		moduleCountStyle.Render(fmt.Sprintf("%d active", len(enabled))),
		themeID, variant)

	infoBox := distroBoxStyle.Render(distroText + "\n" + modulesText)
	sb.WriteString(infoBox + "\n\n")

	// Menu items
	for i, item := range dashboardItems {
		active := i == m.cursor
		cursor := listCursor(active, colCyan)

		iconColor := colGray
		labelColor := colWhite
		if active {
			iconColor = colCyan
			labelColor = colCyan
		}
		icon := col("["+item.icon+"]", iconColor)
		label := col(item.label, labelColor)

		sb.WriteString(cursor + icon + " " + label + "\n")
	}
	// Messages
	if m.errMsg != "" {
		sb.WriteString("\n" + components.ErrorStyle.Render("  ✗ "+m.errMsg))
	}
	if m.status != "" {
		sb.WriteString("\n" + components.SuccessStyle.Render("  ✓ "+m.status))
	}

	return sb.String()
}

func (m Model) updateDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(dashboardItems)-1 {
			m.cursor++
		}
	case key.Matches(msg, m.keys.Enter):
		return m.handleDashboardSelect()
	}
	return m, nil
}

func (m Model) handleDashboardSelect() (tea.Model, tea.Cmd) {
	m.errMsg = ""
	m.status = ""

	switch m.cursor {
	case 0: // Select Modules
		m.state = StateModules
		m.cursor = 0
	case 1: // Reorder Modules
		m.state = StateOrder
		m.cursor = 0
	case 2: // Apply Template
		m.state = StateTemplates
		m.cursor = 0
	case 3: // Theme & Style
		m.state = StateTheme
		m.cursor = 0
	case 4: // Services Picker
		m.state = StateServices
		m.serviceCursor = 0
		m.serviceFilter = ""
		svcs, err := loadSystemServices()
		if err == nil {
			m.systemServices = svcs
		}
	case 5: // ASCII Art Text
		m.state = StateAsciiArt
		m.textInput.Focus()
		return m, m.textInput.Cursor.BlinkCmd()
	case 6: // Preview MOTD
		m.state = StatePreview
		m.cursor = 0
		result, err := m.generator.Generate(context.Background())
		if err != nil {
			m.errMsg = err.Error()
		} else if result == "" {
			m.previewText = "(No modules enabled — nothing to preview)"
		} else {
			m.previewText = result
		}
	case 7: // Auto-Detect
		m.autoDetect()
	case 8: // Full-Auto
		m.fullAuto()
	case 9: // Save & Apply
		m.saveAndApply()
	case 10: // Profiles
		m.state = StateProfiles
		m.cursor = 0
	case 11: // Rollback
		m.state = StateRollback
		m.cursor = 0
		backups, _ := m.backupMgr.List()
		m.backups = backups
	case 12: // Quit
		return m, tea.Quit
	}

	return m, nil
}
