// Package tui implements the Bubble Tea terminal user interface for mom.
package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

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
	StateModuleHelp
	StateGitPaths
	StateModuleSettings
	StateGenericSettings
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
	templates      []*template.Template
	backups        []backup.Backup
	previewText    string
	viewport       viewport.Model
	textInput      textinput.Model
	systemServices []string
	serviceFilter  string
	serviceCursor  int
	profileInput   textinput.Model
	profileSaving  bool

	// Git paths editor state
	gitPathInput   textinput.Model
	gitPathCursor  int
	gitPathAdding  bool

	// Module settings hub state
	settingsCursor int

	// Generic settings editor state
	genericSettingsMod     module.Module
	genericSettingsCursor  int
	genericSettingsEditing bool
	genericSettingsInput   textinput.Model
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

	gpi := textinput.New()
	gpi.Placeholder = "e.g. ~/work or /srv/projects"
	gpi.CharLimit = 120
	gpi.Width = 50

	gsi := textinput.New()
	gsi.Placeholder = "Enter value..."
	gsi.CharLimit = 120
	gsi.Width = 40

	// Set the generator's render options from config
	th := theme.MustGet(cfg.ThemeID())
	gen.RenderOpts = render.Options{
		Theme:   th,
		Variant: render.Variant(cfg.GlobalVariant()),
	}

	return Model{
		state:                StateDashboard,
		registry:             reg,
		config:               cfg,
		generator:            gen,
		writer:               w,
		backupMgr:            bm,
		distroInfo:           di,
		keys:                 keys.DefaultKeyMap(),
		templates:            templates,
		textInput:            ti,
		profileInput:         pi,
		gitPathInput:         gpi,
		genericSettingsInput: gsi,
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
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 6
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

		// Git path input mode
		if m.state == StateGitPaths && m.gitPathAdding {
			return m.updateGitPaths(msg)
		}

		// Generic settings input mode
		if m.state == StateGenericSettings && m.genericSettingsEditing {
			return m.updateGenericSettings(msg)
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
			if m.state == StateModuleHelp {
				m.state = StateModules
				return m, nil
			}
			if m.state == StateGenericSettings {
				m.state = StateModuleSettings
				return m, nil
			}
			if m.state == StateServices || m.state == StateGitPaths || m.state == StateAsciiArt {
				m.state = StateModuleSettings
				return m, nil
			}
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
		case StateGitPaths:
			return m.updateGitPaths(msg)
		case StateModuleSettings:
			return m.updateModuleSettings(msg)
		case StateGenericSettings:
			return m.updateGenericSettings(msg)
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
	case StateModuleHelp:
		content = m.viewModuleHelp()
	case StateGitPaths:
		content = m.viewGitPaths()
	case StateModuleSettings:
		content = m.viewModuleSettings()
	case StateGenericSettings:
		content = m.viewGenericSettings()
	}

	page := content + "\n\n" + m.viewStatusBar()

	// Center the content in the terminal.
	// We must normalize line widths first — lipgloss.Place with
	// lipgloss.Center centers each line individually, which causes
	// staggered alignment when lines have different lengths.
	if m.width > 0 && m.height > 0 {
		page = normalizeLineWidths(page)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, page)
	}
	return page
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
	case StateGitPaths:
		help = "↑↓ Navigate  [a] Add  [x] Remove  [+/-] Max repos  Esc Back"
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

// padRight pads a rendered string (which may contain ANSI escapes) to a
// fixed visible width using spaces.
func padRight(s string, width int) string {
	visible := lipgloss.Width(s)
	if visible >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visible)
}

// normalizeLineWidths pads every line in a multiline string to the width
// of the widest line. This prevents lipgloss.Place(..., Center, ...) from
// centering each line individually (which causes staggered alignment).
func normalizeLineWidths(s string) string {
	lines := strings.Split(s, "\n")
	maxW := 0
	for _, line := range lines {
		if w := lipgloss.Width(line); w > maxW {
			maxW = w
		}
	}
	for i, line := range lines {
		lines[i] = padRight(line, maxW)
	}
	return strings.Join(lines, "\n")
}

// listCursor returns a 4-char-wide cursor. Uses plain ANSI to guarantee
// consistent width between active and inactive rows.
func listCursor(active bool, color string) string {
	if active {
		return "  " + color + "▸" + "\033[0m" + " "
	}
	return "    "
}

// col renders text in a color with ANSI escapes.
func col(text, color string) string {
	if color == "" {
		return text
	}
	return color + text + "\033[0m"
}

// fixedCol renders text padded to exactly `width` visible chars, then colored.
// Uses lipgloss.Width for correct Unicode-aware measurement.
func fixedCol(text string, width int, color string) string {
	visible := lipgloss.Width(text)
	if visible > width {
		text = truncateToWidth(text, width)
		visible = lipgloss.Width(text)
	}
	padded := text + strings.Repeat(" ", width-visible)
	return col(padded, color)
}

// truncateToWidth cuts text so its visible width does not exceed `maxWidth`.
// Uses go-runewidth for correct measurement matching lipgloss.
func truncateToWidth(text string, maxWidth int) string {
	var sb strings.Builder
	currentWidth := 0
	for _, r := range text {
		rw := runewidth.RuneWidth(r)
		if currentWidth+rw > maxWidth {
			break
		}
		sb.WriteRune(r)
		currentWidth += rw
	}
	return sb.String()
}

// dimText renders text in gray.
func dimText(s string) string {
	return "\033[90m" + s + "\033[0m"
}

// ANSI color constants for the TUI.
const (
	colCyan    = "\033[96m"
	colMagenta = "\033[95m"
	colGreen   = "\033[92m"
	colYellow  = "\033[93m"
	colRed     = "\033[91m"
	colWhite   = "\033[97m"
	colGray    = "\033[90m"
	colBold    = "\033[1m"
	colReset   = "\033[0m"
)
