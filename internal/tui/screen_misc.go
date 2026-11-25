package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/msalexms/MoM/internal/config"
	"github.com/msalexms/MoM/internal/i18n"
	"github.com/msalexms/MoM/internal/module/render"
	"github.com/msalexms/MoM/internal/template"
	"github.com/msalexms/MoM/internal/theme"
	c "github.com/msalexms/MoM/internal/tui/components"
)

// --- Theme and style ------------------------------------------------------------

// styleVariants are the variants offered in the picker, in display order.
var styleVariants = []struct {
	id   render.Variant
	desc string
}{
	{render.VariantDefault, "title with a rule and key/value pairs"},
	{render.VariantCompact, "one line per module"},
	{render.VariantBoxed, "rounded boxes"},
	{render.VariantPowerline, "powerline blocks with arrows"},
	{render.VariantCards, "double-border cards"},
	{render.VariantMinimal, "just the data, no decoration"},
}

// The appearance screen is one list: languages, then themes, then variants.
func (m Model) themeRows() (themes []*theme.Theme, total int) {
	themes = theme.All()
	return themes, len(i18n.Languages) + len(themes) + len(styleVariants)
}

func (m Model) updateTheme(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	themes, total := m.themeRows()
	if m.moveCursor(msg, total) || !key.Matches(msg, m.keys.Enter) {
		return m, nil
	}
	i := m.cursor()
	nLang := len(i18n.Languages)
	switch {
	case i < nLang:
		return m.setLanguage(i18n.Languages[i].ID)
	case i < nLang+len(themes):
		th := themes[i-nLang]
		m.cfg.Mode.Theme = th.ID
		m.ok(T("Theme: %s."), T(th.Name))
	default:
		v := styleVariants[i-nLang-len(themes)]
		m.cfg.Mode.Variant = string(v.id)
		m.ok(T("Style: %s."), v.id)
	}
	m.changed()
	return m, nil
}

// setLanguage switches the interface at once and stores the choice right
// away, so it survives quitting without installing. The MOTD itself only
// changes language on the next install.
func (m Model) setLanguage(lang i18n.Lang) (tea.Model, tea.Cmd) {
	if string(lang) == m.cfg.Mode.Language {
		return m, nil
	}
	m.cfg.Mode.Language = string(lang)
	i18n.Set(lang)
	if err := config.SaveLanguage(string(lang)); err != nil {
		m.fail(err)
		return m, nil
	}
	m.changed()
	m.ok("%s", T("Language changed. Install (s) to update the MOTD too."))
	return m, nil
}

func (m Model) viewTheme() string {
	themes, _ := m.themeRows()
	cur := m.cursor()
	nLang := len(i18n.Languages)

	// Left column: the three groups, names only.
	var list strings.Builder
	desc := ""
	list.WriteString(c.Label.Render(T("LANGUAGE")) + "\n")
	for i, l := range i18n.Languages {
		list.WriteString(m.choiceRow(i == cur, string(l.ID) == m.cfg.Mode.Language, l.Name))
		if i == cur {
			desc = T("Language of the interface and of the MOTD text.")
		}
	}
	list.WriteString("\n" + c.Label.Render(T("THEME")) + "\n")
	for i, th := range themes {
		list.WriteString(m.choiceRow(nLang+i == cur, th.ID == m.cfg.ThemeID(), T(th.Name)))
		if nLang+i == cur {
			desc = T(th.Description)
		}
	}
	list.WriteString("\n" + c.Label.Render(T("STYLE")) + "\n")
	for j, v := range styleVariants {
		list.WriteString(m.choiceRow(nLang+len(themes)+j == cur, string(v.id) == m.cfg.GlobalVariant(), string(v.id)))
		if nLang+len(themes)+j == cur {
			desc = T(v.desc)
		}
	}

	// Right column: what the highlighted option is, and a live sample.
	previewTheme := theme.MustGet(m.cfg.ThemeID())
	previewVariant := render.Variant(m.cfg.GlobalVariant())
	switch {
	case cur >= nLang && cur < nLang+len(themes):
		previewTheme = themes[cur-nLang]
	case cur >= nLang+len(themes):
		previewVariant = styleVariants[cur-nLang-len(themes)].id
	}
	rightW := max(20, m.contentWidth()-26)
	right := c.Text2.Width(rightW).Render(desc) + "\n\n" +
		c.Label.Render(T("PREVIEW")) + "\n" +
		m.themeSample(render.Options{Theme: previewTheme, Variant: previewVariant})

	return lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(26).Render(list.String()), right)
}

func (m Model) choiceRow(active, current bool, name string) string {
	mark := "  "
	if current {
		mark = c.OK.Render("●") + " "
	}
	return c.Cursor(active) + mark + c.Item(name, active, true) + "\n"
}

// themeSample renders the real system module with the given options.
func (m Model) themeSample(opts render.Options) string {
	mod, ok := m.reg.Get("system")
	if !ok {
		return ""
	}
	out, err := mod.Generate(context.Background(), opts)
	if err != nil {
		return ""
	}
	r := render.New(opts)
	return out + "\n" + "    " + r.Field(T("memory"), r.ProgressBar(62, 20, ""))
}

// --- Templates ---------------------------------------------------------------

func (m Model) updateTemplates(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.moveCursor(msg, len(m.templates)) || !key.Matches(msg, m.keys.Enter) || len(m.templates) == 0 {
		return m, nil
	}
	tmpl := m.templates[m.cursor()]
	template.Apply(tmpl, m.cfg)
	m.changed()
	m.screen = scrHome
	m.ok(T("Template “%s” applied. Check the preview (v) and install it with s."), tmpl.Name)
	return m, nil
}

func (m Model) viewTemplates() string {
	var sb strings.Builder
	sb.WriteString(c.Muted.Render(T("A template changes which modules are enabled. It does not touch your settings or theme.")) + "\n\n")
	if len(m.templates) == 0 {
		return sb.String() + c.Muted.Render(T("No templates available."))
	}
	for i, t := range m.templates {
		active := i == m.cursor()
		n := 0
		for _, on := range t.Modules {
			if on {
				n++
			}
		}
		sb.WriteString(c.Cursor(active) + c.Pad(c.Item(t.Name, active, true), 14) +
			c.Pad(c.Muted.Render(fmt.Sprintf(T("%2d modules"), n)), 13) + c.Text2.Render(T(t.Description)) + "\n")
	}

	t := m.templates[m.cursor()]
	var names []string
	for _, name := range config.KnownModuleNames() {
		if t.Modules[name] {
			if mod, ok := m.reg.Get(name); ok {
				names = append(names, T(mod.Title()))
			}
		}
	}
	sb.WriteString("\n" + c.Label.Render(T("INCLUDES")) + "\n" + c.Muted.Render(strings.Join(names, ", ")) + "\n")
	return sb.String()
}

// --- Profiles ----------------------------------------------------------------

func (m Model) updateProfiles(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	n := len(m.profiles) + 1 // row 0 saves the current config
	if m.moveCursor(msg, n) {
		m.confirmDel = ""
		return m, nil
	}
	cur := m.cursor()
	switch {
	case key.Matches(msg, m.keys.Enter) && cur == 0:
		return m.startInput(inputProfileName, "", T("profile name"))
	case key.Matches(msg, m.keys.Enter):
		name := m.profiles[cur-1]
		loaded, err := config.LoadProfile(name)
		if err != nil {
			m.fail(err)
			return m, nil
		}
		*m.cfg = *loaded
		i18n.Set(i18n.Lang(m.cfg.Mode.Language))
		m.changed()
		m.screen = scrHome
		m.ok(T("Profile “%s” loaded. Install it with s."), name)
	case msg.String() == "d" && cur > 0:
		name := m.profiles[cur-1]
		if m.confirmDel != name {
			m.confirmDel = name
			m.flash, m.flashErr = fmt.Sprintf(T("Press d again to delete “%s”."), name), true
			return m, nil
		}
		m.confirmDel = ""
		if err := config.DeleteProfile(name); err != nil {
			m.fail(err)
			return m, nil
		}
		m.profiles = config.ListProfiles()
		m.setCursor(min(cur, len(m.profiles)))
		m.ok(T("Profile “%s” deleted."), name)
	}
	return m, nil
}

func (m Model) saveProfile(raw string) (tea.Model, tea.Cmd) {
	name := config.SanitizeProfileName(raw)
	if name == "" {
		m.fail(errors.New(T("the name can only contain letters, digits, - and _")))
		return m, nil
	}
	if err := config.SaveProfile(m.cfg, name); err != nil {
		m.fail(err)
		return m, nil
	}
	m.profiles = config.ListProfiles()
	m.ok(T("Profile “%s” saved."), name)
	return m, nil
}

func (m Model) viewProfiles() string {
	var sb strings.Builder
	sb.WriteString(c.Muted.Render(T("A profile stores the whole config: modules, settings, order and theme.")) + "\n\n")

	active := m.cursor() == 0
	sb.WriteString(c.Cursor(active) + c.Yellow.Render("+ ") + c.Item(T("Save the current config as a profile"), active, true) + "\n")
	if m.purpose == inputProfileName {
		sb.WriteString("    " + m.input.View() + "\n")
	}
	if len(m.profiles) == 0 {
		sb.WriteString("\n" + c.Muted.Render(T("No saved profiles yet.")) + "\n")
	}
	for i, p := range m.profiles {
		active := m.cursor() == i+1
		sb.WriteString(c.Cursor(active) + "  " + c.Item(p, active, true) + "\n")
	}
	return sb.String()
}

// --- Preview -----------------------------------------------------------------

// previewJumps are the screens reachable from the preview; Esc there comes
// back to a freshly generated preview.
var previewJumps = map[string]screen{"m": scrModules, "o": scrOrder, "c": scrSettings, "t": scrTheme}

func (m Model) updatePreview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if s, ok := previewJumps[msg.String()]; ok {
		m.fromPreview = true
		return m.open(s)
	}
	if msg.String() == "r" {
		return m.open(scrPreview)
	}
	var cmd tea.Cmd
	m.preview, cmd = m.preview.Update(msg)
	return m, cmd
}

func (m Model) viewPreview() string {
	if m.previewLoading {
		return c.Muted.Render(T("Generating the preview…"))
	}
	info := c.Muted.Render(fmt.Sprintf(T("This is how it will look at login · %3.0f%%"), m.preview.ScrollPercent()*100))
	return info + "\n" + m.preview.View()
}

// --- Help --------------------------------------------------------------------

func (m Model) viewHelp() string {
	rows := [][2]string{
		{"↑↓ · j k", T("move through lists")},
		{"Enter", T("open, choose or change a value")},
		{T("Space"), T("turn on or off")},
		{"Esc", T("go back or cancel")},
		{"s", T("generate and install the MOTD (asks for sudo once)")},
		{"?", T("this help; in Modules, what the highlighted module does")},
		{"q", T("quit (warns about changes not installed)")},
		{"Ctrl+C", T("quit without asking")},
	}
	var sb strings.Builder
	sb.WriteString(c.Label.Render(T("KEYS")) + "\n")
	for _, r := range rows {
		sb.WriteString(c.Pad(c.Key.Render(r[0]), 14) + c.Text.Render(r[1]) + "\n")
	}
	sb.WriteString("\n" + c.Label.Render(T("HOW IT WORKS")) + "\n")
	for _, line := range []string{
		T("Changes are saved to ~/.config/mom/config.toml when you install."),
		T("mom installs a login hook that renders the MOTD with live data at every login."),
		T("Every install is saved: preview it and install it again from Backups."),
		T("Without the interface: mom --print, mom --apply, mom --uninstall (mom --help)."),
	} {
		sb.WriteString(c.Text2.Render(line) + "\n")
	}
	return sb.String()
}
