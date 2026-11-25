package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/msalexms/MoM/internal/theme"
	c "github.com/msalexms/MoM/internal/tui/components"
)

type menuEntry struct {
	group    string // non-empty starts a new group
	label    string
	hint     string
	shortcut string
	action   func(Model) (tea.Model, tea.Cmd)
}

func goTo(s screen) func(Model) (tea.Model, tea.Cmd) {
	return func(m Model) (tea.Model, tea.Cmd) { return m.open(s) }
}

var homeMenu = []menuEntry{
	{group: "Configure", label: "Modules", hint: "which sections are shown", shortcut: "m", action: goTo(scrModules)},
	{label: "Order", hint: "the order they appear in", shortcut: "o", action: goTo(scrOrder)},
	{label: "Module settings", hint: "city, services, git paths…", shortcut: "c", action: goTo(scrSettings)},
	{label: "Appearance", hint: "language, colors and section style", shortcut: "t", action: goTo(scrTheme)},
	{group: "Start from something", label: "Templates", hint: "ready-made setups", shortcut: "p", action: goTo(scrTemplates)},
	{label: "Profiles", hint: "save and restore your setups", shortcut: "f", action: goTo(scrProfiles)},
	{group: "MOTD", label: "Preview", hint: "how it will look at login", shortcut: "v", action: goTo(scrPreview)},
	{label: "Install", hint: "installs the login hook (asks for sudo)", shortcut: "s", action: func(m Model) (tea.Model, tea.Cmd) { return m.save() }},
	{label: "Backups", hint: "go back to a previous MOTD", shortcut: "b", action: goTo(scrBackups)},
}

func (m Model) updateHome(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.moveCursor(msg, len(homeMenu)) {
		return m, nil
	}
	if key.Matches(msg, m.keys.Enter) {
		return homeMenu[m.cursor()].action(m)
	}
	for i, e := range homeMenu {
		if msg.String() == e.shortcut {
			m.setCursor(i)
			return e.action(m)
		}
	}
	return m, nil
}

func (m Model) viewHome() string {
	var sb strings.Builder

	// Brand block: Thor mark next to the product name.
	text := strings.Join([]string{
		"",
		c.Strong.Render("mom") + c.Muted.Render(T("  ·  MOTD manager")),
		c.Muted.Render(T("Your Linux welcome screen, your way.")),
		c.BrandMark() + c.Muted.Render(" tech · amstech.dev"),
	}, "\n")
	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, c.ThorLogo(), "   ", text))
	sb.WriteString("\n\n")

	// System summary.
	th, _ := theme.Get(m.cfg.ThemeID())
	themeName := m.cfg.ThemeID()
	if th != nil {
		themeName = T(th.Name)
	}
	distroName := m.distro.Name
	if distroName == "" {
		distroName = string(m.distro.Family)
	}
	sep := c.Line.Render("  ·  ")
	sb.WriteString(c.Text2.Render(distroName) + sep +
		c.OK.Render(fmt.Sprintf(T("%d modules enabled"), len(m.cfg.EnabledModuleNames()))) + sep +
		c.Text2.Render(T("theme ")+themeName) + sep +
		c.Text2.Render(T("style ")+m.cfg.GlobalVariant()) + "\n")

	// Menu.
	for i, e := range homeMenu {
		if e.group != "" {
			sb.WriteString("\n" + c.Label.Render(strings.ToUpper(T(e.group))) + "\n")
		}
		active := i == m.cursor()
		sb.WriteString(c.Cursor(active) +
			c.Pad(c.Item(T(e.label), active, true), 22) +
			c.Pad(c.Muted.Render(T(e.hint)), 40) +
			c.Key.Render(e.shortcut) + "\n")
	}
	return sb.String()
}
