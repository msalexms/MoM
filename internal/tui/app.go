// Package tui implements the Bubble Tea terminal user interface for mom.
package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ams/mom/internal/backup"
	"github.com/ams/mom/internal/config"
	"github.com/ams/mom/internal/distro"
	"github.com/ams/mom/internal/generator"
	"github.com/ams/mom/internal/module"
	"github.com/ams/mom/internal/template"
	"github.com/ams/mom/internal/tui/components"
	"github.com/ams/mom/internal/tui/keys"
)

// AppState represents the current view state.
type AppState int

const (
	StateDashboard AppState = iota
	StateModules
	StateTemplates
	StatePreview
	StateHelp
	StateRollback
)

// Model is the main Bubble Tea model.
type Model struct {
	state      AppState
	width      int
	height     int
	cursor     int
	unsaved    bool
	status     string
	errMsg     string

	// Dependencies
	registry   *module.Registry
	config     *config.Config
	generator  *generator.Generator
	writer     *generator.Writer
	backupMgr  *backup.Manager
	distroInfo distro.Info
	keys       keys.KeyMap

	// Sub-state data
	templates    []*template.Template
	backups      []backup.Backup
	previewText  string
}

// NewModel creates a new TUI model with all dependencies.
func NewModel(
	reg *module.Registry,
	cfg *config.Config,
	gen *generator.Generator,
	w *generator.Writer,
	bm *backup.Manager,
	di distro.Info,
) Model {
	templates, _ := template.BuiltinTemplates()

	return Model{
		state:      StateDashboard,
		registry:   reg,
		config:     cfg,
		generator:  gen,
		writer:     w,
		backupMgr:  bm,
		distroInfo: di,
		keys:       keys.DefaultKeyMap(),
		templates:  templates,
	}
}

// Init is the Bubble Tea init function.
func (m Model) Init() tea.Cmd {
	return tea.SetWindowTitle("mom — MOTD Manager")
}

// Update handles messages and user input.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Global shortcuts
		if key.Matches(msg, m.keys.Quit) && m.state == StateDashboard {
			return m, tea.Quit
		}
		if key.Matches(msg, m.keys.Help) {
			if m.state == StateHelp {
				m.state = StateDashboard
			} else {
				m.state = StateHelp
			}
			m.cursor = 0
			return m, nil
		}
		if key.Matches(msg, m.keys.Back) {
			if m.state != StateDashboard {
				m.state = StateDashboard
				m.cursor = 0
				m.errMsg = ""
			}
			return m, nil
		}

		// Delegate to view-specific handlers
		switch m.state {
		case StateDashboard:
			return m.updateDashboard(msg)
		case StateModules:
			return m.updateModules(msg)
		case StateTemplates:
			return m.updateTemplates(msg)
		case StatePreview:
			return m.updatePreview(msg)
		case StateRollback:
			return m.updateRollback(msg)
		}
	}

	return m, nil
}

// View renders the current state.
func (m Model) View() string {
	var content string

	switch m.state {
	case StateDashboard:
		content = m.viewDashboard()
	case StateModules:
		content = m.viewModules()
	case StateTemplates:
		content = m.viewTemplates()
	case StatePreview:
		content = m.viewPreview()
	case StateHelp:
		content = m.viewHelp()
	case StateRollback:
		content = m.viewRollback()
	}

	// Status bar at the bottom
	statusBar := m.viewStatusBar()

	return content + "\n" + statusBar
}

// --- Dashboard ---

var dashboardItems = []string{
	"Select Modules",
	"Apply Template",
	"Preview MOTD",
	"Auto-Detect Modules",
	"Full-Auto Setup",
	"Save & Apply",
	"Rollback",
	"Quit",
}

func (m Model) viewDashboard() string {
	var sb strings.Builder

	// Header
	title := components.TitleStyle.Render("mom — Message Of the day Manager")
	sb.WriteString(title + "\n\n")

	// Distro info
	distroLine := fmt.Sprintf("  Distro: %s (%s)", m.distroInfo.Name, m.distroInfo.Family)
	sb.WriteString(components.InfoStyle.Render(distroLine) + "\n")

	// Active modules summary
	enabled := m.config.EnabledModuleNames()
	moduleLine := fmt.Sprintf("  Active modules: %d/12 [%s]", len(enabled), strings.Join(enabled, ", "))
	sb.WriteString(components.HelpStyle.Render(moduleLine) + "\n\n")

	// Menu
	for i, item := range dashboardItems {
		cursor := "  "
		style := components.MenuItemStyle
		if i == m.cursor {
			cursor = "▸ "
			style = components.ActiveMenuItemStyle
		}
		sb.WriteString(cursor + style.Render(item) + "\n")
	}

	// Error message
	if m.errMsg != "" {
		sb.WriteString("\n" + components.ErrorStyle.Render("  ✗ "+m.errMsg))
	}

	// Success message
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
	case 1: // Apply Template
		m.state = StateTemplates
		m.cursor = 0
	case 2: // Preview MOTD
		m.state = StatePreview
		m.cursor = 0
		// Generate preview
		result, err := m.generator.Generate(context.Background())
		if err != nil {
			m.errMsg = err.Error()
		} else if result == "" {
			m.previewText = "(No modules enabled — nothing to preview)"
		} else {
			m.previewText = result
		}
	case 3: // Auto-Detect
		m.autoDetect()
	case 4: // Full-Auto
		m.fullAuto()
	case 5: // Save & Apply
		m.saveAndApply()
	case 6: // Rollback
		m.state = StateRollback
		m.cursor = 0
		backups, _ := m.backupMgr.List()
		m.backups = backups
	case 7: // Quit
		return m, tea.Quit
	}

	return m, nil
}

// --- Modules View ---

func (m Model) viewModules() string {
	var sb strings.Builder

	sb.WriteString(components.HeadingStyle.Render("  Module Selection") + "\n\n")

	allModules := m.registry.Ordered()
	for i, mod := range allModules {
		cursor := "  "
		if i == m.cursor {
			cursor = "▸ "
		}

		enabled := m.config.IsModuleEnabled(mod.Name())
		available := mod.Available()

		checkbox := "[ ]"
		style := lipgloss.NewStyle()
		if enabled {
			checkbox = components.CheckboxChecked.Render("[✓]")
		} else {
			checkbox = components.CheckboxUnchecked.Render("[ ]")
		}

		name := mod.Title()
		desc := mod.Description()

		if !available {
			style = components.DisabledStyle
			deps := strings.Join(mod.Dependencies(), ", ")
			desc = fmt.Sprintf("[missing: %s]", deps)
		}

		line := fmt.Sprintf("%s %s — %s", checkbox, name, desc)
		sb.WriteString(cursor + style.Render(line) + "\n")
	}

	sb.WriteString("\n" + components.HelpStyle.Render("  [space] toggle  [a] all on  [d] all off  [esc] back"))

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

// --- Templates View ---

func (m Model) viewTemplates() string {
	var sb strings.Builder

	sb.WriteString(components.HeadingStyle.Render("  Apply Template") + "\n\n")

	for i, tmpl := range m.templates {
		cursor := "  "
		if i == m.cursor {
			cursor = "▸ "
		}
		style := components.MenuItemStyle
		if i == m.cursor {
			style = components.ActiveMenuItemStyle
		}

		line := fmt.Sprintf("%s — %s", tmpl.Name, tmpl.Description)
		sb.WriteString(cursor + style.Render(line) + "\n")
	}

	sb.WriteString("\n" + components.HelpStyle.Render("  [enter] apply  [esc] back"))

	return sb.String()
}

func (m Model) updateTemplates(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxIdx := len(m.templates) - 1

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
		if m.cursor < len(m.templates) {
			tmpl := m.templates[m.cursor]
			template.Apply(tmpl, m.config)
			m.unsaved = true
			m.status = fmt.Sprintf("Template '%s' applied", tmpl.Name)
			m.state = StateDashboard
			m.cursor = 0
		}
	}

	return m, nil
}

// --- Preview View ---

func (m Model) viewPreview() string {
	var sb strings.Builder

	sb.WriteString(components.HeadingStyle.Render("  MOTD Preview") + "\n\n")

	if m.previewText == "" {
		sb.WriteString("  (Generating...)\n")
	} else {
		// Indent preview content
		for _, line := range strings.Split(m.previewText, "\n") {
			sb.WriteString("  " + line + "\n")
		}
	}

	sb.WriteString("\n" + components.HelpStyle.Render("  [esc] back"))

	return sb.String()
}

func (m Model) updatePreview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Preview is read-only, just handle back
	return m, nil
}

// --- Rollback View ---

func (m Model) viewRollback() string {
	var sb strings.Builder

	sb.WriteString(components.HeadingStyle.Render("  Rollback") + "\n\n")

	if len(m.backups) == 0 {
		sb.WriteString("  No backups available\n")
	} else {
		for i, b := range m.backups {
			cursor := "  "
			if i == m.cursor {
				cursor = "▸ "
			}
			label := b.Timestamp.Format("2006-01-02 15:04:05")
			if b.IsOriginal {
				label += " [ORIGINAL]"
			}
			if b.Distro != "" {
				label += fmt.Sprintf(" (%s)", b.Distro)
			}
			style := components.MenuItemStyle
			if i == m.cursor {
				style = components.ActiveMenuItemStyle
			}
			sb.WriteString(cursor + style.Render(label) + "\n")
		}
	}

	sb.WriteString("\n" + components.HelpStyle.Render("  [enter] restore  [esc] back"))

	return sb.String()
}

func (m Model) updateRollback(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxIdx := len(m.backups) - 1

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
		if m.cursor < len(m.backups) {
			b := m.backups[m.cursor]
			paths := distro.GetPaths(m.distroInfo.Family)
			err := m.backupMgr.Restore(context.Background(), &b, paths.MotdFile)
			if err != nil {
				m.errMsg = "Rollback failed: " + err.Error()
			} else {
				m.status = "Rollback successful"
			}
			m.state = StateDashboard
			m.cursor = 0
		}
	}

	return m, nil
}

// --- Help View ---

func (m Model) viewHelp() string {
	var sb strings.Builder

	sb.WriteString(components.HeadingStyle.Render("  Help — Keyboard Shortcuts") + "\n\n")

	helpItems := []struct{ key, desc string }{
		{"↑/k, ↓/j", "Navigate up/down"},
		{"Enter", "Select / Confirm"},
		{"Space", "Toggle module on/off"},
		{"a", "Enable all available modules"},
		{"d", "Disable all modules"},
		{"s / Ctrl+S", "Save & Apply MOTD"},
		{"Esc", "Go back / Cancel"},
		{"?", "Toggle this help screen"},
		{"q / Ctrl+C", "Quit"},
	}

	for _, h := range helpItems {
		keyStyle := lipgloss.NewStyle().Bold(true).Width(14)
		sb.WriteString(fmt.Sprintf("  %s %s\n", keyStyle.Render(h.key), h.desc))
	}

	sb.WriteString("\n" + components.HelpStyle.Render("  Press ? or Esc to close"))

	return sb.String()
}

// --- Status Bar ---

func (m Model) viewStatusBar() string {
	var parts []string

	if m.unsaved {
		parts = append(parts, components.UnsavedStyle.Render("[unsaved]"))
	}

	help := "[↑↓] Navigate  [Enter] Select  [?] Help  [q] Quit"
	switch m.state {
	case StateModules:
		help = "[↑↓] Navigate  [Space] Toggle  [a] All on  [d] All off  [Esc] Back"
	case StateTemplates:
		help = "[↑↓] Navigate  [Enter] Apply  [Esc] Back"
	case StatePreview, StateHelp:
		help = "[Esc] Back"
	case StateRollback:
		help = "[↑↓] Navigate  [Enter] Restore  [Esc] Back"
	}

	parts = append(parts, components.HelpStyle.Render(help))
	return components.StatusBarStyle.Render(strings.Join(parts, "  "))
}

// --- Actions ---

func (m *Model) autoDetect() {
	available := m.registry.Available()
	for _, mod := range available {
		m.config.SetModuleEnabled(mod.Name(), true)
	}
	m.unsaved = true
	m.status = fmt.Sprintf("Auto-detected %d available modules", len(available))
}

func (m *Model) fullAuto() {
	m.autoDetect()
	m.saveAndApply()
}

func (m *Model) saveAndApply() {
	// Save config
	if err := config.Save(m.config); err != nil {
		m.errMsg = "Save failed: " + err.Error()
		return
	}

	// Generate MOTD
	content, err := m.generator.Generate(context.Background())
	if err != nil {
		m.errMsg = "Generate failed: " + err.Error()
		return
	}
	if content == "" {
		m.errMsg = "No modules produced output"
		return
	}

	// Write to system
	modules := m.config.EnabledModuleNames()
	if err := m.writer.Write(context.Background(), content, modules); err != nil {
		m.errMsg = "Write failed: " + err.Error()
		return
	}

	m.unsaved = false
	m.status = "MOTD saved and applied!"
}
