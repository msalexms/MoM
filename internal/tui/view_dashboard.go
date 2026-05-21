package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/msalexms/MoM/internal/tui/components"
)

var dashboardItems = []struct {
	icon  string
	label string
}{
	{" + ", "Select Modules"},
	{" ↕ ", "Reorder Modules"},
	{" # ", "Apply Template"},
	{" T ", "Theme & Style"},
	{" ⚙ ", "Module Settings"},
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

	header := ` ███╗   ███╗ ██████╗ ███╗   ███╗
 ████╗ ████║██╔═══██╗████╗ ████║
 ██╔████╔██║██║   ██║██╔████╔██║
 ██║╚██╔╝██║██║   ██║██║╚██╔╝██║
 ██║ ╚═╝ ██║╚██████╔╝██║ ╚═╝ ██║
 ╚═╝     ╚═╝ ╚═════╝ ╚═╝     ╚═╝`

	// Info box: center both lines within the box using the wider line as inner width.
	distroText := fmt.Sprintf("%s  |  %s", m.distroInfo.Name, m.distroInfo.Family)
	enabled := m.config.EnabledModuleNames()
	themeID := m.config.ThemeID()
	variant := m.config.GlobalVariant()
	modulesText := fmt.Sprintf("Modules: %s  Theme: %s  Style: %s",
		moduleCountStyle.Render(fmt.Sprintf("%d active", len(enabled))),
		themeID, variant)

	innerW := lipgloss.Width(distroText)
	if w := lipgloss.Width(modulesText); w > innerW {
		innerW = w
	}
	lineStyle := lipgloss.NewStyle().Width(innerW).Align(lipgloss.Center)
	infoBox := distroBoxStyle.Render(lineStyle.Render(distroText) + "\n" + lineStyle.Render(modulesText))

	// blockW is the width of the widest element (infobox or menu items).
	// The logo and subtitle are centered within this width.
	blockW := lipgloss.Width(infoBox)
	for _, item := range dashboardItems {
		// cursor(4) + icon(5) + space(1) + label
		if w := 4 + 5 + 1 + lipgloss.Width(item.label); w > blockW {
			blockW = w
		}
	}

	sb.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00BFFF")).
		Bold(true).
		Width(blockW).
		Align(lipgloss.Center).
		Render(header) + "\n")
	sb.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true).
		Width(blockW).
		Align(lipgloss.Center).
		Render("Message Of the Day Manager") + "\n\n")
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
		icon := fixedCol("["+item.icon+"]", 5, iconColor)
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
	case 4: // Module Settings
		m.state = StateModuleSettings
		m.settingsCursor = 0
	case 5: // Preview MOTD
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
		m.viewport = viewport.New(m.width-4, m.height-6)
		m.viewport.SetContent(trimTrailingSpaces(m.previewText))
	case 6: // Auto-Detect
		m.autoDetect()
	case 7: // Full-Auto
		m.fullAuto()
	case 8: // Save & Apply
		m.saveAndApply()
	case 9: // Profiles
		m.state = StateProfiles
		m.cursor = 0
	case 10: // Rollback
		m.state = StateRollback
		m.cursor = 0
		backups, _ := m.backupMgr.List()
		m.backups = backups
	case 11: // Quit
		return m, tea.Quit
	}

	return m, nil
}
