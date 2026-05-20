package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ams/mom/internal/module/render"
)

// TmuxModule displays active tmux sessions.
type TmuxModule struct{}

func (m *TmuxModule) Name() string           { return "tmux" }
func (m *TmuxModule) Title() string          { return "Tmux Sessions" }
func (m *TmuxModule) Description() string    { return "Active tmux/screen sessions" }
func (m *TmuxModule) Dependencies() []string { return []string{"tmux"} }
func (m *TmuxModule) Available() bool        { return CheckDependency("tmux") }
func (m *TmuxModule) DefaultEnabled() bool   { return false }
func (m *TmuxModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *TmuxModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *TmuxModule) Settings() []SettingDef         { return nil }

func (m *TmuxModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

type tmuxSession struct {
	name     string
	windows  int
	attached bool
}

func (m *TmuxModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "tmux", "list-sessions", "-F", "#{session_name}\t#{session_windows}\t#{session_attached}")
	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	var sessions []tmuxSession
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}
		var wins int
		fmt.Sscanf(parts[1], "%d", &wins)
		sessions = append(sessions, tmuxSession{parts[0], wins, parts[2] == "1"})
	}
	if len(sessions) == 0 {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()

	var lines []string
	for _, s := range sessions {
		status := th.Dim("detached")
		if s.attached {
			status = th.Color("attached", th.Palette.Success)
		}
		lines = append(lines, fmt.Sprintf("%-14s  %dw  %s", th.Color(s.name, th.Palette.Warning), s.windows, status))
	}

	compact := fmt.Sprintf("%d sessions", len(sessions))
	return r.Section("Tmux Sessions", "tmux", compact, lines), nil
}
