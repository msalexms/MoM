// Package tui implements the Bubble Tea terminal user interface for mom.
package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ams/mom/internal/backup"
	"github.com/ams/mom/internal/config"
	"github.com/ams/mom/internal/distro"
	"github.com/ams/mom/internal/generator"
	"github.com/ams/mom/internal/module"
	"github.com/ams/mom/internal/module/render"
	"github.com/ams/mom/internal/template"
	"github.com/ams/mom/internal/theme"
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
	StateAsciiArt
	StateServices
	StateTheme
	StateOrder
	StateProfiles
)

// Model is the main Bubble Tea model.
type Model struct {
	state   AppState
	width   int
	height  int
	cursor  int
	unsaved bool
	status  string
	errMsg  string

	// Dependencies
	registry   *module.Registry
	config     *config.Config
	generator  *generator.Generator
	writer     *generator.Writer
	backupMgr  *backup.Manager
	distroInfo distro.Info
	keys       keys.KeyMap

	// Sub-state data
	templates       []*template.Template
	backups         []backup.Backup
	previewText     string
	textInput       textinput.Model
	systemServices  []string
	serviceFilter   string
	serviceCursor   int
	profileInput    textinput.Model
	profileSaving   bool
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

	ti := textinput.New()
	ti.Placeholder = "Type your text here..."
	ti.CharLimit = 40
	ti.Width = 40

	pi := textinput.New()
	pi.Placeholder = "Profile name..."
	pi.CharLimit = 30
	pi.Width = 30

	// Set the generator's render options from config
	th := theme.MustGet(cfg.ThemeID())
	gen.RenderOpts = render.Options{
		Theme:   th,
		Variant: render.Variant(cfg.GlobalVariant()),
	}

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
		textInput:  ti,
		profileInput: pi,
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
		// Text input mode (ASCII art)
		if m.state == StateAsciiArt && m.textInput.Focused() {
			return m.updateAsciiArt(msg)
		}

		// Profile name input mode
		if m.state == StateProfiles && m.profileSaving {
			return m.updateProfiles(msg)
		}

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
		case StateServices:
			return m.updateServices(msg)
		case StateTheme:
			return m.updateTheme(msg)
		case StateOrder:
			return m.updateOrder(msg)
		case StateProfiles:
			return m.updateProfiles(msg)
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
	case StateAsciiArt:
		content = m.viewAsciiArt()
	case StateServices:
		content = m.viewServices()
	case StateTheme:
		content = m.viewTheme()
	case StateOrder:
		content = m.viewOrder()
	case StateProfiles:
		content = m.viewProfiles()
	}

	return content + "\n\n" + m.viewStatusBar()
}

// --- Status Bar ---

func (m Model) viewStatusBar() string {
	var parts []string

	if m.unsaved {
		parts = append(parts, lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).Bold(true).Render("● unsaved"))
	}

	help := "↑↓ Navigate  Enter Select  ? Help  q Quit"
	switch m.state {
	case StateModules:
		help = "↑↓ Navigate  Space Toggle  a All on  d All off  Esc Back"
	case StateTemplates:
		help = "↑↓ Navigate  Enter Apply  Esc Back"
	case StatePreview, StateHelp:
		help = "Esc Back"
	case StateRollback:
		help = "↑↓ Navigate  Enter Restore  Esc Back"
	case StateAsciiArt:
		help = "Enter Apply  Esc Cancel"
	case StateServices:
		help = "↑↓ Navigate  Space Toggle  / Filter  Esc Back"
	case StateTheme:
		help = "↑↓ Navigate  Enter Select  Esc Back"
	case StateOrder:
		help = "↑↓ Navigate  Shift+↑↓ Move  Esc Back"
	}

	helpRendered := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render(help)
	parts = append(parts, helpRendered)

	return "  " + strings.Join(parts, "  ")
}

// --- Actions ---

func (m *Model) autoDetect() {
	available := m.registry.Available()
	for _, mod := range available {
		m.config.SetModuleEnabled(mod.Name(), true)
	}
	m.unsaved = true
	m.status = fmt.Sprintf("Auto-detected %d available modules!", len(available))
}

func (m *Model) fullAuto() {
	m.autoDetect()
	m.saveAndApply()
}

func (m *Model) saveAndApply() {
	if err := config.Save(m.config); err != nil {
		m.errMsg = "Save failed: " + err.Error()
		return
	}

	content, err := m.generator.Generate(context.Background())
	if err != nil {
		m.errMsg = "Generate failed: " + err.Error()
		return
	}
	if content == "" {
		m.errMsg = "No modules produced output"
		return
	}

	modules := m.config.EnabledModuleNames()
	if err := m.writer.Write(context.Background(), content, modules); err != nil {
		m.errMsg = "Write failed: " + err.Error()
		return
	}

	m.unsaved = false
	m.status = "MOTD saved and applied!"
}

// --- Styles (shared) ---

var (
	logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00BFFF")).
			Bold(true)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)

	distroBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444444")).
			Padding(0, 1)

	moduleCountStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF7F")).
				Bold(true)

	viewTitleStyle = func(color string) lipgloss.Style {
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(color))
	}

	separatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))
)

func viewSeparator() string {
	return separatorStyle.Render("  ─────────────────────────────────────────")
}
