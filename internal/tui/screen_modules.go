package tui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/msalexms/MoM/internal/config"
	"github.com/msalexms/MoM/internal/distro"
	"github.com/msalexms/MoM/internal/module"
	c "github.com/msalexms/MoM/internal/tui/components"
)

// --- Module selection ----------------------------------------------------------

// visibleModules is the module list narrowed by the search text, which
// matches the title, the name and the description.
func (m Model) visibleModules() []module.Module {
	all := m.reg.Ordered()
	q := strings.ToLower(strings.TrimSpace(m.modFilter))
	if q == "" {
		return all
	}
	var out []module.Module
	for _, mod := range all {
		text := strings.ToLower(T(mod.Title()) + " " + mod.Name() + " " + T(mod.Description()))
		if strings.Contains(text, q) {
			out = append(out, mod)
		}
	}
	return out
}

func (m Model) toggleModule(mods []module.Module) Model {
	if i := m.cursor(); i < len(mods) {
		name := mods[i].Name()
		m.cfg.SetModuleEnabled(name, !m.cfg.IsModuleEnabled(name))
		m.changed()
	}
	return m
}

func (m Model) updateModules(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	mods := m.visibleModules()
	if m.moveCursor(msg, len(mods)) {
		return m, nil
	}
	switch {
	case msg.String() == "/":
		m.modFiltering = true
		m.setCursor(0)
	case key.Matches(msg, m.keys.Space), key.Matches(msg, m.keys.Enter):
		m = m.toggleModule(mods)
	case key.Matches(msg, m.keys.AllOn):
		n := 0
		for _, mod := range mods {
			if mod.Available() {
				m.cfg.SetModuleEnabled(mod.Name(), true)
				n++
			}
		}
		m.changed()
		m.ok(T("Enabled the %d modules available on this machine."), n)
	case key.Matches(msg, m.keys.AllOff):
		names := config.KnownModuleNames()
		if m.modFilter != "" {
			names = nil
			for _, mod := range mods {
				names = append(names, mod.Name())
			}
		}
		for _, name := range names {
			m.cfg.SetModuleEnabled(name, false)
		}
		m.changed()
		m.ok("%s", T("Modules disabled."))
	}
	return m, nil
}

// updateModuleSearch owns the keyboard while the search box is active:
// letters are typed, arrows move and Space still toggles.
func (m Model) updateModuleSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	mods := m.visibleModules()
	switch msg.Type {
	case tea.KeyEsc:
		m.modFilter, m.modFiltering = "", false
		m.setCursor(0)
	case tea.KeyEnter:
		m.modFiltering = false
	case tea.KeyUp:
		m.setCursor(max(0, m.cursor()-1))
	case tea.KeyDown:
		m.setCursor(min(max(0, len(mods)-1), m.cursor()+1))
	case tea.KeySpace:
		m = m.toggleModule(mods)
	case tea.KeyBackspace:
		if r := []rune(m.modFilter); len(r) > 0 {
			m.modFilter = string(r[:len(r)-1])
			m.setCursor(0)
		}
	case tea.KeyRunes:
		m.modFilter += string(msg.Runes)
		m.setCursor(0)
	}
	return m, nil
}

// rootNote explains the "root" tag. On Debian the login hook runs as root,
// so those modules work there; elsewhere it runs as the user.
func (m Model) rootNote() string {
	if m.distro.Family == distro.FamilyDebian {
		return T("root: reads data only root can see; it works at login, which runs as root here.")
	}
	return T("root: reads data only root can see; at login it shows nothing unless your user can read it.")
}

func (m Model) viewModules() string {
	mods := m.visibleModules()
	var sb strings.Builder

	switch {
	case m.modFiltering || m.modFilter != "":
		cursor := ""
		if m.modFiltering {
			cursor = c.Brand.Render("▏")
		}
		sb.WriteString(c.Muted.Render(T("Search: ")) + c.Accent.Render(m.modFilter) + cursor +
			c.Muted.Render(Tf("   %d of %d", len(mods), len(m.reg.Ordered()))) + "\n")
	default:
		sb.WriteString(c.Muted.Render(Tf("%d of %d enabled · / to search", len(m.cfg.EnabledModuleNames()), len(mods))) + "\n")
	}
	sb.WriteString(c.Yellow.Render("root") + c.Muted.Render(strings.TrimPrefix(m.rootNote(), "root")) + "\n\n")

	if len(mods) == 0 {
		sb.WriteString(c.Muted.Render(T("No module matches the search.")) + "\n")
		return sb.String()
	}
	start, end := c.Window(len(mods), m.cursor(), m.listHeight(4))
	for i := start; i < end; i++ {
		mod := mods[i]
		active := i == m.cursor()
		avail := mod.Available()

		tag := "     "
		if module.NeedsRoot(mod) {
			tag = c.Yellow.Render("root ")
		}
		desc := c.Muted.Render(T(mod.Description()))
		if !avail {
			if deps := mod.Dependencies(); len(deps) > 0 {
				desc = c.Bad.Render(T("needs ") + strings.Join(deps, T(" or ")))
			} else {
				desc = c.Muted.Render(T("not available on this machine"))
			}
		}
		sb.WriteString(c.Cursor(active) + c.Check(m.cfg.IsModuleEnabled(mod.Name())) + " " +
			c.Pad(c.Item(T(mod.Title()), active, avail), 22) + tag + desc + "\n")
	}
	if info := c.ScrollInfo(start, end, len(mods)); info != "" {
		sb.WriteString("\n" + info)
	}
	return sb.String()
}

// --- Order -------------------------------------------------------------------

// moduleOrder returns the configured order, completed with any registered
// module missing from it so every module can be moved.
func (m Model) moduleOrder() []string {
	order := slices.Clone(m.cfg.Mode.ModuleOrder)
	if len(order) == 0 {
		order = config.DefaultModuleOrder()
	}
	for _, mod := range m.reg.Ordered() {
		if !slices.Contains(order, mod.Name()) {
			order = append(order, mod.Name())
		}
	}
	return order
}

func (m Model) updateOrder(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	order := m.moduleOrder()
	cur := m.cursor()
	to := -1
	switch {
	case key.Matches(msg, m.keys.MoveUp):
		to = cur - 1
	case key.Matches(msg, m.keys.MoveDown):
		to = cur + 1
	default:
		m.moveCursor(msg, len(order))
		return m, nil
	}
	if to >= 0 && to < len(order) {
		order[cur], order[to] = order[to], order[cur]
		m.cfg.Mode.ModuleOrder = order
		m.setCursor(to)
		m.changed()
	}
	return m, nil
}

func (m Model) viewOrder() string {
	order := m.moduleOrder()
	var sb strings.Builder
	sb.WriteString(c.Muted.Render(T("Sections appear in the MOTD from top to bottom. Only enabled ones (●) count.")) + "\n\n")

	start, end := c.Window(len(order), m.cursor(), m.listHeight(3))
	for i := start; i < end; i++ {
		name := order[i]
		title := name
		if mod, ok := m.reg.Get(name); ok {
			title = T(mod.Title())
		}
		active := i == m.cursor()
		enabled := m.cfg.IsModuleEnabled(name)
		row := c.Cursor(active) + c.Muted.Render(fmt.Sprintf("%2d ", i+1)) + c.Dot(enabled) + " " + c.Item(title, active, enabled)
		if active {
			row += c.Muted.Render(T("   Shift+↑↓ to move"))
		}
		sb.WriteString(row + "\n")
	}
	if info := c.ScrollInfo(start, end, len(order)); info != "" {
		sb.WriteString("\n" + info)
	}
	return sb.String()
}

// --- Module help -------------------------------------------------------------

func (m Model) helpModule() module.Module {
	mods := m.visibleModules()
	if i := m.cursors[scrModules]; i < len(mods) {
		return mods[i]
	}
	return nil
}

func (m Model) viewModuleHelp() string {
	mod := m.helpModule()
	if mod == nil {
		return ""
	}
	row := func(label, value string) string {
		return c.Pad(c.Muted.Render(label), 14) + value + "\n"
	}

	var sb strings.Builder
	sb.WriteString(c.Strong.Render(T(mod.Title())) + "\n" + c.Text2.Render(T(mod.Description())) + "\n\n")

	state := c.OK.Render(T("available on this machine"))
	if !mod.Available() {
		state = c.Bad.Render(T("not available on this machine"))
	}
	sb.WriteString(row(T("Status"), state))
	sb.WriteString(row(T("Enabled"), map[bool]string{true: c.OK.Render(T("yes")), false: c.Text.Render(T("no"))}[m.cfg.IsModuleEnabled(mod.Name())]))
	deps := c.Text.Render(T("nothing"))
	if d := mod.Dependencies(); len(d) > 0 {
		deps = c.Text.Render(strings.Join(d, T(" or ")))
	}
	sb.WriteString(row(T("Needs"), deps))
	sb.WriteString(row(T("Name"), c.Text.Render(mod.Name())))
	if module.NeedsRoot(mod) {
		sb.WriteString(row(T("Permissions"), c.Yellow.Render(strings.TrimPrefix(m.rootNote(), "root: "))))
	}

	if cm, ok := mod.(module.Configurable); ok {
		sb.WriteString("\n" + c.Label.Render(T("SETTINGS")) + c.Muted.Render(T("  (in Module settings)")) + "\n")
		for _, s := range cm.Settings() {
			sb.WriteString(row(T(s.Label), c.Text.Render(displayValue(s, m.cfg))))
			sb.WriteString(c.Pad("", 14) + c.Muted.Render(T(s.Description)) + "\n")
		}
	} else {
		sb.WriteString("\n" + c.Muted.Render(T("No settings: it works as is.")) + "\n")
	}
	return sb.String()
}
