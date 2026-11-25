// Package tui implements the interactive interface of mom with Bubble Tea.
package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/msalexms/MoM/internal/backup"
	"github.com/msalexms/MoM/internal/config"
	"github.com/msalexms/MoM/internal/distro"
	"github.com/msalexms/MoM/internal/generator"
	"github.com/msalexms/MoM/internal/i18n"
	"github.com/msalexms/MoM/internal/module"
	"github.com/msalexms/MoM/internal/permission"
	"github.com/msalexms/MoM/internal/template"
	c "github.com/msalexms/MoM/internal/tui/components"
	"github.com/msalexms/MoM/internal/tui/keys"
)

type screen int

const (
	scrHome screen = iota
	scrModules
	scrOrder
	scrSettings       // configurable modules
	scrModuleSettings // settings of one module
	scrList           // editor of a list setting
	scrServices       // systemd service picker
	scrTheme
	scrTemplates
	scrProfiles
	scrPreview
	scrBackups
	scrHelp
	scrModuleHelp
)

var screenTitles = map[screen]string{
	scrModules:        "Modules",
	scrOrder:          "Order",
	scrSettings:       "Module settings",
	scrModuleSettings: "Settings",
	scrList:           "List",
	scrServices:       "Services",
	scrTheme:          "Appearance",
	scrTemplates:      "Templates",
	scrProfiles:       "Profiles",
	scrPreview:        "Preview",
	scrBackups:        "Backups",
	scrHelp:           "Help",
	scrModuleHelp:     "Module help",
}

// inputPurpose says what the shared text input is editing.
type inputPurpose int

const (
	inputNone inputPurpose = iota
	inputSetting
	inputListItem
	inputProfileName
)

// Model is the Bubble Tea model.
type Model struct {
	screen  screen
	cursors map[screen]int // each screen keeps its own cursor
	width   int
	height  int

	unsaved     bool
	confirmQuit bool
	confirmDel  string // profile awaiting a second "d"
	busy        string // non-empty while a background job runs
	flash       string
	flashErr    bool

	reg     *module.Registry
	cfg     *config.Config
	gen     *generator.Generator
	writer  *generator.Writer
	backups *backup.Manager
	distro  distro.Info
	keys    keys.KeyMap

	templates  []*template.Template
	profiles   []string
	backupList []backup.Backup

	preview        viewport.Model
	previewLoading bool

	input   textinput.Model
	purpose inputPurpose

	settingsMod module.Module  // module whose settings are open
	listSetting module.Setting // list setting being edited
	services    []string
	svcFilter   string
	helpReturn  screen
	fromPreview bool // opened from the preview: Esc goes back to a fresh preview

	modFilter    string // search text of the module list
	modFiltering bool   // the search box has the keyboard

	pending func(context.Context) result // privileged job waiting for sudo

	newRegistry    func() *module.Registry // fresh modules, to render backups
	backupPreviews map[string]backupPreview
	currentBackup  string // path of the backup that is installed now
}

// result is the outcome of a background job.
type result struct {
	text     string
	err      error
	saved    bool           // the config was persisted
	restored *config.Config // config brought back from a backup
	reload   bool           // backups changed: reload the list
}

type (
	resultMsg  result
	previewMsg struct {
		content string
		err     error
	}
	authMsg          struct{ err error }
	backupPreviewMsg struct {
		path string
		backupPreview
	}
)

// NewModel creates the TUI model.
// newRegistry builds a fresh set of modules; it is used to render backups
// without touching the live modules.
func NewModel(reg *module.Registry, cfg *config.Config, gen *generator.Generator, w *generator.Writer, bm *backup.Manager, di distro.Info, newRegistry func() *module.Registry) Model {
	templates, _ := template.BuiltinTemplates() // an empty list is handled by the view

	in := textinput.New()
	in.Prompt = "› "
	in.PromptStyle = c.Accent
	in.TextStyle = c.Text
	in.Cursor.Style = c.Brand
	in.CharLimit = 200

	return Model{
		cursors:   map[screen]int{},
		reg:       reg,
		cfg:       cfg,
		gen:       gen,
		writer:    w,
		backups:   bm,
		distro:    di,
		keys:      keys.DefaultKeyMap(),
		templates: templates,
		input:     in,

		newRegistry:    newRegistry,
		backupPreviews: map[string]backupPreview{},
	}
}

// Init sets the terminal window title.
func (m Model) Init() tea.Cmd {
	return tea.SetWindowTitle(T("mom · MOTD manager · [AMS] tech"))
}

func (m Model) cursor() int      { return m.cursors[m.screen] }
func (m *Model) setCursor(i int) { m.cursors[m.screen] = i }

// moveCursor handles up/down over n items and reports whether the key was
// a movement key.
func (m *Model) moveCursor(msg tea.KeyMsg, n int) bool {
	cur := m.cursor()
	switch {
	case key.Matches(msg, m.keys.Up):
		m.setCursor(max(0, cur-1))
	case key.Matches(msg, m.keys.Down):
		m.setCursor(min(max(0, n-1), cur+1))
	case msg.String() == "home" || msg.String() == "g":
		m.setCursor(0)
	case msg.String() == "end" || msg.String() == "G":
		m.setCursor(max(0, n-1))
	default:
		return false
	}
	return true
}

func (m *Model) ok(format string, args ...any) {
	m.flash, m.flashErr = fmt.Sprintf(format, args...), false
}

func (m *Model) fail(err error) {
	m.flash, m.flashErr = err.Error(), true
}

// changed marks the config as modified and refreshes the modules.
func (m *Model) changed() {
	m.unsaved = true
	m.reg.SyncConfig(m.cfg)
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.preview.Width, m.preview.Height = m.contentWidth(), m.listHeight(1)
		return m, nil

	case previewMsg:
		m.previewLoading = false
		switch {
		case msg.err != nil:
			m.preview.SetContent(c.Bad.Render(T("Could not generate: ") + msg.err.Error()))
		case msg.content == "":
			m.preview.SetContent(c.Muted.Render(T("No enabled module has anything to show.")))
		default:
			m.preview.SetContent(msg.content)
		}
		return m, nil

	case authMsg:
		job := m.pending
		m.pending = nil
		if msg.err != nil || job == nil {
			m.busy = ""
			m.fail(errors.New(T("sudo was not authorized: the MOTD was not changed")))
			return m, nil
		}
		return m, runJob(job)

	case backupPreviewMsg:
		m.backupPreviews[msg.path] = msg.backupPreview
		return m, nil

	case resultMsg:
		m.busy = ""
		if msg.err != nil {
			m.fail(msg.err)
			return m, nil
		}
		if msg.restored != nil {
			*m.cfg = *msg.restored
			i18n.Set(i18n.Lang(m.cfg.Mode.Language))
			m.reg.SyncConfig(m.cfg)
		}
		if msg.saved {
			m.unsaved = false
		}
		m.ok("%s", msg.text)
		if msg.reload || m.screen == scrBackups {
			m.reloadBackups()
			return m, m.loadBackupPreview()
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	if m.purpose != inputNone {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if m.busy != "" {
		return m, nil
	}
	m.flash = ""

	if m.purpose != inputNone {
		return m.updateInput(msg)
	}
	if m.screen == scrServices {
		return m.updateServices(msg)
	}
	if m.screen == scrModules && (m.modFiltering || (m.modFilter != "" && key.Matches(msg, m.keys.Back))) {
		return m.updateModuleSearch(msg)
	}

	if m.confirmQuit {
		m.confirmQuit = false
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
	}

	switch {
	case key.Matches(msg, m.keys.Quit):
		if m.unsaved {
			m.confirmQuit = true
			m.flash, m.flashErr = T("There are changes not installed yet. Press q again to quit without saving, or s to install them."), true
			return m, nil
		}
		return m, tea.Quit
	case key.Matches(msg, m.keys.Save) && m.screen != scrProfiles:
		return m.save()
	case key.Matches(msg, m.keys.Help):
		return m.toggleHelp(), nil
	case key.Matches(msg, m.keys.Back):
		return m.back()
	}

	switch m.screen {
	case scrHome:
		return m.updateHome(msg)
	case scrModules:
		return m.updateModules(msg)
	case scrOrder:
		return m.updateOrder(msg)
	case scrSettings:
		return m.updateSettings(msg)
	case scrModuleSettings:
		return m.updateModuleSettings(msg)
	case scrList:
		return m.updateList(msg)
	case scrTheme:
		return m.updateTheme(msg)
	case scrTemplates:
		return m.updateTemplates(msg)
	case scrProfiles:
		return m.updateProfiles(msg)
	case scrPreview:
		return m.updatePreview(msg)
	case scrBackups:
		return m.updateBackups(msg)
	}
	return m, nil
}

// parent is the screen Esc returns to.
func (m Model) parent() screen {
	switch m.screen {
	case scrModuleSettings:
		return scrSettings
	case scrList, scrServices:
		return scrModuleSettings
	case scrModuleHelp:
		return scrModules
	case scrHelp:
		return m.helpReturn
	case scrModules, scrOrder, scrSettings, scrTheme:
		if m.fromPreview {
			return scrPreview
		}
	}
	return scrHome
}

// back returns to the parent screen. Going back to the preview renders it
// again, so changes made on the way are visible at once.
func (m Model) back() (Model, tea.Cmd) {
	target := m.parent()
	if target == scrPreview || target == scrHome {
		m.fromPreview = false
	}
	if target == scrPreview {
		return m.open(scrPreview)
	}
	m.screen = target
	return m, nil
}

func (m Model) toggleHelp() Model {
	switch m.screen {
	case scrHelp, scrModuleHelp:
		m, _ = m.back()
		return m
	case scrModules:
		m.screen = scrModuleHelp
	default:
		m.helpReturn = m.screen
		m.screen = scrHelp
	}
	return m
}

// open switches to a screen, loading whatever data it shows.
func (m Model) open(s screen) (Model, tea.Cmd) {
	m.screen = s
	switch s {
	case scrProfiles:
		m.profiles = config.ListProfiles()
		m.setCursor(min(m.cursor(), len(m.profiles)))
	case scrBackups:
		m.reloadBackups()
		m.setCursor(0)
		return m, m.loadBackupPreview()
	case scrPreview:
		m.previewLoading = true
		m.preview = viewport.New(m.contentWidth(), m.listHeight(1))
		gen := m.gen
		return m, func() tea.Msg {
			out, err := gen.Generate(context.Background())
			return previewMsg{content: out, err: err}
		}
	}
	return m, nil
}

// --- Privileged jobs -------------------------------------------------------

// save generates and installs the MOTD and persists the config.
func (m Model) save() (tea.Model, tea.Cmd) {
	cfg, gen, w := m.cfg, m.gen, m.writer
	return m.privileged(T("Generating and installing the MOTD…"), func(ctx context.Context) result {
		if err := generator.Apply(ctx, cfg, gen, w); err != nil {
			return result{err: err}
		}
		return result{text: T("MOTD installed: it will show live data at every login."), saved: true}
	})
}

// privileged runs job in the background, asking for the sudo password
// first when needed. The prompt runs through tea.ExecProcess, which
// suspends the interface and gives sudo the real terminal.
func (m Model) privileged(busy string, job func(context.Context) result) (tea.Model, tea.Cmd) {
	m.busy = busy
	if permission.NeedsAuth(context.Background()) {
		m.pending = job
		return m, tea.ExecProcess(permission.AuthCommand(), func(err error) tea.Msg { return authMsg{err} })
	}
	return m, runJob(job)
}

func runJob(job func(context.Context) result) tea.Cmd {
	return func() tea.Msg { return resultMsg(job(context.Background())) }
}

// --- Text input --------------------------------------------------------------

func (m Model) startInput(p inputPurpose, value, placeholder string) (Model, tea.Cmd) {
	m.purpose = p
	m.input.SetValue(value)
	m.input.Placeholder = placeholder
	m.input.CursorEnd()
	return m, m.input.Focus()
}

func (m Model) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.purpose = inputNone
		m.input.Blur()
		return m, nil
	case tea.KeyEnter:
		value := m.input.Value()
		p := m.purpose
		m.purpose = inputNone
		m.input.Blur()
		return m.submitInput(p, value)
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) submitInput(p inputPurpose, value string) (tea.Model, tea.Cmd) {
	switch p {
	case inputSetting:
		settings := m.moduleSettings()
		if i := m.cursors[scrModuleSettings]; i < len(settings) {
			if err := settings[i].Set(m.cfg, value); err != nil {
				m.fail(err)
				return m, nil
			}
			m.changed()
		}
	case inputListItem:
		value = strings.TrimSpace(value)
		if value != "" {
			items := append(m.listSetting.Values(m.cfg), value)
			if err := m.listSetting.Set(m.cfg, strings.Join(items, ",")); err != nil {
				m.fail(err)
				return m, nil
			}
			m.changed()
			m.setCursor(len(m.listSetting.Values(m.cfg)) - 1)
		}
	case inputProfileName:
		return m.saveProfile(value)
	}
	return m, nil
}

// --- Layout ------------------------------------------------------------------

const maxContentWidth = 96

// contentWidth is the usable width inside the page margins.
func (m Model) contentWidth() int {
	if m.width == 0 {
		return 80
	}
	return max(20, min(m.width-4, maxContentWidth))
}

// listHeight is the number of rows left for a list once the frame and
// `extra` lines of the view itself are accounted for.
func (m Model) listHeight(extra int) int {
	if m.height == 0 {
		return 20
	}
	// header (2) + footer (3) + top/bottom margin (2)
	return max(3, m.height-7-extra)
}

// View renders the frame: header, current screen, status and key hints.
func (m Model) View() string {
	w := m.contentWidth()

	var body string
	switch m.screen {
	case scrHome:
		body = m.viewHome()
	case scrModules:
		body = m.viewModules()
	case scrOrder:
		body = m.viewOrder()
	case scrSettings:
		body = m.viewSettings()
	case scrModuleSettings:
		body = m.viewModuleSettings()
	case scrList:
		body = m.viewList()
	case scrServices:
		body = m.viewServices()
	case scrTheme:
		body = m.viewTheme()
	case scrTemplates:
		body = m.viewTemplates()
	case scrProfiles:
		body = m.viewProfiles()
	case scrPreview:
		body = m.viewPreview()
	case scrBackups:
		body = m.viewBackups()
	case scrHelp:
		body = m.viewHelp()
	case scrModuleHelp:
		body = m.viewModuleHelp()
	}

	header := m.viewHeader(w)
	footer := m.viewFooter(w)

	bodyH := 0
	if m.height > 0 {
		bodyH = max(1, m.height-2-lipgloss.Height(header)-lipgloss.Height(footer))
	}
	body = lipgloss.NewStyle().Width(w).Height(bodyH).MaxHeight(max(bodyH, 1)).Render(body)
	if bodyH == 0 {
		body = lipgloss.NewStyle().Width(w).Render(body)
	}

	page := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
	if m.width == 0 {
		return page
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, page)
}

func (m Model) viewHeader(w int) string {
	left := c.BrandMark() + " " + c.Strong.Render("mom")
	if m.screen != scrHome {
		title := T(screenTitles[m.screen])
		if m.screen == scrList || m.screen == scrServices {
			title = T(m.listSetting.Label)
		}
		if mod := m.settingsModTitle(); mod != "" && mod != title &&
			(m.screen == scrModuleSettings || m.screen == scrList || m.screen == scrServices || m.screen == scrModuleHelp) {
			title = mod + " · " + title
		}
		left += c.Line.Render("  ›  ") + c.Text.Render(title)
	}
	right := ""
	if m.unsaved {
		right = c.Yellow.Render(T("● changes not installed"))
	}
	gap := max(1, w-lipgloss.Width(left)-lipgloss.Width(right))
	return left + strings.Repeat(" ", gap) + right + "\n" + c.Rule(w)
}

func (m Model) settingsModTitle() string {
	if m.screen == scrModuleHelp {
		if mod := m.helpModule(); mod != nil {
			return T(mod.Title())
		}
		return ""
	}
	if m.settingsMod != nil {
		return T(m.settingsMod.Title())
	}
	return ""
}

func (m Model) viewFooter(w int) string {
	status := ""
	switch {
	case m.busy != "":
		status = c.Brand.Render("◌ " + m.busy)
	case m.flash != "" && m.flashErr:
		status = c.Bad.Render("✗ " + m.flash)
	case m.flash != "":
		status = c.OK.Render("✓ " + m.flash)
	}
	status = lipgloss.NewStyle().Width(w).Render(status)
	return c.Rule(w) + "\n" + status + "\n" + m.hints()
}

func (m Model) hints() string {
	if m.purpose != inputNone {
		return c.Hints("Enter", T("accept"), "Esc", T("cancel"))
	}
	nav := []string{"↑↓", T("move")}
	common := []string{"Esc", T("back"), "?", T("help")}
	var h []string
	switch m.screen {
	case scrHome:
		h = append(nav, "Enter", T("open"), "s", T("install"), "?", T("help"), "q", T("quit"))
	case scrModules:
		if m.modFiltering {
			h = []string{"↑↓", T("move"), T("Space"), T("toggle"), "Enter", T("done"), "Esc", T("clear")}
		} else {
			h = append(nav, T("Space"), T("toggle"), "/", T("search"), "a/d", T("all/none"), "?", T("what it does"), "Esc", T("back"))
		}
	case scrOrder:
		h = append(nav, T("Shift+↑↓"), T("move"), "Esc", T("back"))
	case scrSettings:
		h = append(append(nav, "Enter", T("configure")), common...)
	case scrModuleSettings:
		h = append(append(nav, "Enter", T("change")), common...)
	case scrList:
		h = append(append(nav, "a", T("add"), "x", T("remove")), common...)
	case scrServices:
		h = append(nav, T("Space"), T("select"), T("type"), T("filter"), "Esc", T("back"))
	case scrTheme:
		h = append(append(nav, "Enter", T("choose")), common...)
	case scrTemplates:
		h = append(append(nav, "Enter", T("apply")), common...)
	case scrProfiles:
		h = append(append(nav, "Enter", T("load/save"), "d", T("delete")), common...)
	case scrPreview:
		h = []string{"↑↓", T("scroll"), "m/o/c/t", T("edit"), "r", T("refresh"), "s", T("install"), "Esc", T("back")}
	case scrBackups:
		h = append(append(nav, "Enter", T("restore")), common...)
	default:
		h = []string{"Esc", T("back"), "q", T("quit")}
	}
	return c.Hints(h...)
}
