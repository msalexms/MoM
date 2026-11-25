package tui

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/msalexms/MoM/internal/backup"
	"github.com/msalexms/MoM/internal/config"
	"github.com/msalexms/MoM/internal/distro"
	"github.com/msalexms/MoM/internal/generator"
	"github.com/msalexms/MoM/internal/i18n"
	"github.com/msalexms/MoM/internal/module"
)

func newTestModel(t *testing.T) Model {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	reg := module.NewRegistry()
	reg.RegisterAll(
		&module.SystemModule{},
		&module.WeatherModule{},
		&module.GitStatusModule{},
		&module.ServicesModule{},
		&module.QuoteModule{},
	)
	cfg := config.Defaults()
	reg.SyncConfig(cfg)
	bm := backup.NewManager(filepath.Join(t.TempDir(), "backups"))

	m := NewModel(reg, cfg, generator.NewGenerator(reg, cfg), generator.NewWriter(bm, distro.FamilyDebian), bm,
		distro.Info{Name: "Debian 13", Family: distro.FamilyDebian}, func() *module.Registry { return reg })
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	return next.(Model)
}

// press feeds keys to the model; a key is a rune string ("m"), a name
// ("enter", "esc", "space", "up", "down", "shift+down", "backspace") or
// literal text prefixed with "type:".
func press(t *testing.T, m Model, keys ...string) Model {
	t.Helper()
	for _, k := range keys {
		var msgs []tea.KeyMsg
		switch {
		case strings.HasPrefix(k, "type:"):
			for _, r := range strings.TrimPrefix(k, "type:") {
				msgs = append(msgs, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			}
		case k == "enter":
			msgs = append(msgs, tea.KeyMsg{Type: tea.KeyEnter})
		case k == "esc":
			msgs = append(msgs, tea.KeyMsg{Type: tea.KeyEsc})
		case k == "space":
			msgs = append(msgs, tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
		case k == "up":
			msgs = append(msgs, tea.KeyMsg{Type: tea.KeyUp})
		case k == "down":
			msgs = append(msgs, tea.KeyMsg{Type: tea.KeyDown})
		case k == "shift+down":
			msgs = append(msgs, tea.KeyMsg{Type: tea.KeyShiftDown})
		case k == "backspace":
			msgs = append(msgs, tea.KeyMsg{Type: tea.KeyBackspace})
		case k == "ctrl+c":
			msgs = append(msgs, tea.KeyMsg{Type: tea.KeyCtrlC})
		default:
			msgs = append(msgs, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
		}
		for _, msg := range msgs {
			next, _ := m.Update(msg)
			m = next.(Model)
			_ = m.View() // every intermediate state must render
		}
	}
	return m
}

func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func TestEveryHomeEntryOpensAndReturns(t *testing.T) {
	m := newTestModel(t)
	for _, e := range homeMenu {
		if e.shortcut == "s" {
			continue // installs the MOTD; needs root
		}
		m = press(t, m, e.shortcut)
		if m.screen == scrHome {
			t.Errorf("shortcut %q did not leave the home screen", e.shortcut)
		}
		if !strings.Contains(m.View(), screenTitles[m.screen]) {
			t.Errorf("screen %q: header missing its title", e.label)
		}
		m = press(t, m, "esc")
		if m.screen != scrHome {
			t.Errorf("esc from %q went to %v, want home", e.label, m.screen)
		}
	}
}

func TestToggleModuleMarksUnsaved(t *testing.T) {
	m := press(t, newTestModel(t), "m", "down", "space") // weather
	if !m.cfg.IsModuleEnabled("weather") {
		t.Fatal("weather not enabled")
	}
	if !m.unsaved || !strings.Contains(m.View(), "changes not installed") {
		t.Error("unsaved indicator missing")
	}
	m = press(t, m, "d")
	if len(m.cfg.EnabledModuleNames()) != 0 {
		t.Errorf("d should disable everything, got %v", m.cfg.EnabledModuleNames())
	}
	m = press(t, m, "a")
	if len(m.cfg.EnabledModuleNames()) == 0 {
		t.Error("a should enable the available modules")
	}
}

func TestModuleHelpFollowsCursor(t *testing.T) {
	m := press(t, newTestModel(t), "m", "down", "?")
	if m.screen != scrModuleHelp || !strings.Contains(m.View(), "Weather") {
		t.Fatal("module help should show the highlighted module")
	}
	if m = press(t, m, "esc"); m.screen != scrModules {
		t.Error("esc should return to the module list")
	}
}

func TestReorder(t *testing.T) {
	m := newTestModel(t)
	before := m.moduleOrder()
	m = press(t, m, "o", "shift+down")
	after := m.cfg.Mode.ModuleOrder
	if len(after) == 0 || after[0] != before[1] || after[1] != before[0] {
		t.Errorf("first two modules not swapped: %v → %v", before[:2], after[:2])
	}
	if m.cursor() != 1 {
		t.Error("cursor should follow the moved module")
	}
}

func TestEditSettings(t *testing.T) {
	m := press(t, newTestModel(t), "c")
	mods := m.configurableModules()
	idx := slices.IndexFunc(mods, func(x module.Module) bool { return x.Name() == "weather" })
	for range idx {
		m = press(t, m, "down")
	}
	m = press(t, m, "enter") // open weather settings

	// city: free text
	m = press(t, m, "enter", "type:Madrid", "enter")
	if got := m.cfg.Modules.WeatherConfig.City; got != "Madrid" {
		t.Errorf("city = %q", got)
	}
	// units: enum cycles
	m = press(t, m, "down", "enter")
	if got := m.cfg.Modules.WeatherConfig.Units; got != "imperial" {
		t.Errorf("units = %q, want imperial", got)
	}
	// esc while typing cancels without changing the value
	m = press(t, m, "up", "enter", "type:XX", "esc")
	if got := m.cfg.Modules.WeatherConfig.City; got != "Madrid" {
		t.Errorf("cancelled edit changed city to %q", got)
	}
	if m.screen != scrModuleSettings {
		t.Error("esc in the input should stay on the settings screen")
	}
}

func TestListSettingAddRemove(t *testing.T) {
	m := press(t, newTestModel(t), "c")
	mods := m.configurableModules()
	idx := slices.IndexFunc(mods, func(x module.Module) bool { return x.Name() == "git" })
	for range idx {
		m = press(t, m, "down")
	}
	m = press(t, m, "enter", "enter") // git → paths list
	if m.screen != scrList {
		t.Fatalf("screen = %v, want list editor", m.screen)
	}
	m = press(t, m, "a", "type:~/work", "enter", "a", "type:/srv/code", "enter")
	if got := m.cfg.Modules.GitConfig.Paths; !slices.Equal(got, []string{"~/work", "/srv/code"}) {
		t.Fatalf("paths = %v", got)
	}
	m = press(t, m, "up", "x")
	if got := m.cfg.Modules.GitConfig.Paths; !slices.Equal(got, []string{"/srv/code"}) {
		t.Errorf("after remove paths = %v", got)
	}
}

func TestIntSettingRejectsGarbage(t *testing.T) {
	m := press(t, newTestModel(t), "c")
	mods := m.configurableModules()
	idx := slices.IndexFunc(mods, func(x module.Module) bool { return x.Name() == "git" })
	for range idx {
		m = press(t, m, "down")
	}
	m = press(t, m, "enter", "down", "enter", "backspace", "type:abc", "enter")
	if m.cfg.Modules.GitConfig.MaxRepos != 5 {
		t.Errorf("max_repos changed to %d", m.cfg.Modules.GitConfig.MaxRepos)
	}
	if !m.flashErr {
		t.Error("expected an error message")
	}
}

func TestServicePickerFilterAndToggle(t *testing.T) {
	m := newTestModel(t)
	m.services = []string{"cron", "nginx", "ssh"}
	m.screen = scrServices
	m.settingsMod, _ = m.reg.Get("services")
	m.listSetting = m.moduleSettings()[0]

	m = press(t, m, "type:ngi")
	if got := m.filteredServices(); !slices.Equal(got, []string{"nginx"}) {
		t.Fatalf("filter = %v", got)
	}
	m = press(t, m, "space")
	if got := m.cfg.Modules.ServicesConfig.Services; !slices.Equal(got, []string{"nginx"}) {
		t.Errorf("services = %v", got)
	}
	m = press(t, m, "esc") // clears the filter
	if m.screen != scrServices || m.svcFilter != "" {
		t.Error("first esc should only clear the filter")
	}
	m = press(t, m, "q") // letters filter, they do not quit
	if m.svcFilter != "q" {
		t.Error("q should be typed into the filter")
	}
}

func TestThemeAndVariant(t *testing.T) {
	m := press(t, newTestModel(t), "t", "down", "down", "down", "enter") // skip the two languages
	if m.cfg.ThemeID() == config.DefaultTheme {
		t.Error("theme did not change")
	}
	m = press(t, m, "G", "enter") // last row: the last variant
	if m.cfg.GlobalVariant() != string(styleVariants[len(styleVariants)-1].id) {
		t.Errorf("variant = %q", m.cfg.GlobalVariant())
	}
}

func TestApplyTemplate(t *testing.T) {
	m := newTestModel(t)
	if len(m.templates) == 0 {
		t.Skip("no templates embedded")
	}
	m = press(t, m, "p", "enter")
	if m.screen != scrHome || !m.unsaved {
		t.Error("applying a template should return home with unsaved changes")
	}
}

func TestProfilesSaveLoadDelete(t *testing.T) {
	m := newTestModel(t)
	m.cfg.Modules.WeatherConfig.City = "Lugo"
	m = press(t, m, "f", "enter", "type:mi perfil", "enter")
	if !slices.Equal(m.profiles, []string{"mi-perfil"}) {
		t.Fatalf("profiles = %v", m.profiles)
	}

	m.cfg.Modules.WeatherConfig.City = "Otra"
	m = press(t, m, "down", "enter")
	if m.cfg.Modules.WeatherConfig.City != "Lugo" {
		t.Errorf("profile not loaded, city = %q", m.cfg.Modules.WeatherConfig.City)
	}

	m = press(t, m, "f", "down", "d")
	if len(m.profiles) != 1 {
		t.Fatal("first d must only ask for confirmation")
	}
	m = press(t, m, "d")
	if len(m.profiles) != 0 {
		t.Errorf("profile not deleted: %v", m.profiles)
	}
}

func TestQuitAsksWhenUnsaved(t *testing.T) {
	m := press(t, newTestModel(t), "m", "space", "esc")
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if isQuit(cmd) {
		t.Fatal("q with unsaved changes should ask first")
	}
	_, cmd = next.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if !isQuit(cmd) {
		t.Error("second q should quit")
	}
}

func TestCtrlCQuitsWhileTyping(t *testing.T) {
	m := press(t, newTestModel(t), "f", "enter") // profile name input
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if !isQuit(cmd) {
		t.Error("ctrl+c must always quit")
	}
}

func TestTinyTerminalDoesNotPanic(t *testing.T) {
	m := newTestModel(t)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 30, Height: 8})
	m = next.(Model)
	for _, e := range homeMenu {
		if e.shortcut != "s" {
			m = press(t, m, e.shortcut, "down", "down", "esc")
		}
	}
}

func TestLanguageSwitchesInterface(t *testing.T) {
	t.Cleanup(func() { i18n.Set(i18n.EN) })
	m := newTestModel(t)
	if !strings.Contains(m.View(), "Modules") {
		t.Fatal("English should be the default")
	}
	m = press(t, m, "t", "down", "enter") // Español
	if m.cfg.Mode.Language != "es" || i18n.Current() != i18n.ES {
		t.Fatalf("language = %q", m.cfg.Mode.Language)
	}
	if v := m.View(); !strings.Contains(v, "Apariencia") || !strings.Contains(v, "IDIOMA") {
		t.Error("interface did not switch to Spanish")
	}
	saved, err := config.Load()
	if err != nil || saved.Mode.Language != "es" {
		t.Errorf("language not saved right away: %v %q", err, saved.Mode.Language)
	}
}

func TestModuleSearch(t *testing.T) {
	m := press(t, newTestModel(t), "m", "/", "type:git")
	if got := m.visibleModules(); len(got) != 1 || got[0].Name() != "git" {
		t.Fatalf("search git = %d modules", len(got))
	}
	m = press(t, m, "space")
	if !m.cfg.IsModuleEnabled("git") {
		t.Error("space should toggle the found module while searching")
	}
	m = press(t, m, "q") // typed into the search, does not quit
	if m.modFilter != "gitq" {
		t.Errorf("filter = %q", m.modFilter)
	}
	m = press(t, m, "esc")
	if m.modFilter != "" || m.screen != scrModules {
		t.Error("esc should clear the search and stay on the list")
	}
	m = press(t, m, "esc")
	if m.screen != scrHome {
		t.Error("second esc should go back")
	}
}

func TestRootModulesAreFlagged(t *testing.T) {
	if !module.NeedsRoot(&module.FirewallModule{}) || module.NeedsRoot(&module.SystemModule{}) {
		t.Error("NeedsRoot")
	}
}

func TestPreviewJumpsAndComesBack(t *testing.T) {
	m := press(t, newTestModel(t), "v", "m")
	if m.screen != scrModules {
		t.Fatalf("m in the preview should open Modules, got %v", m.screen)
	}
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	if m.screen != scrPreview || !m.previewLoading || cmd == nil {
		t.Error("esc should return to a regenerating preview")
	}
	m = press(t, m, "esc")
	if m.screen != scrHome {
		t.Error("esc from the preview should go home")
	}
}
