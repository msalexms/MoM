package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/msalexms/MoM/internal/module/render"
)

// LoginsModule displays recent logins and active SSH sessions.
type LoginsModule struct{}

func (m *LoginsModule) Name() string           { return "logins" }
func (m *LoginsModule) Title() string          { return "Logins" }
func (m *LoginsModule) Description() string    { return "Last logins and active SSH sessions" }
func (m *LoginsModule) Dependencies() []string { return []string{"last", "who"} }
func (m *LoginsModule) Available() bool        { return CheckDependency("last") }
func (m *LoginsModule) DefaultEnabled() bool   { return false }

func (m *LoginsModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact}
}
func (m *LoginsModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *LoginsModule) Settings() []SettingDef         { return nil }

func (m *LoginsModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

func (m *LoginsModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	r := render.New(opts)
	th := r.Theme()

	var lines []string

	if CheckDependency("who") {
		cmd := exec.CommandContext(ctx, "who")
		output, err := cmd.Output()
		if err == nil {
			rawLines := strings.Split(strings.TrimSpace(string(output)), "\n")
			count := 0
			if len(rawLines) > 0 && rawLines[0] != "" {
				count = len(rawLines)
			}
			lines = append(lines, fmt.Sprintf("%-10s  %s", th.Color("active", th.Palette.Warning), fmt.Sprintf("%d session(s)", count)))
			lines = append(lines, "")
		}
	}

	cmd := exec.CommandContext(ctx, "last", "-5", "-w")
	output, err := cmd.Output()
	if err == nil {
		rawLines := strings.Split(strings.TrimSpace(string(output)), "\n")
		lines = append(lines, fmt.Sprintf("%-10s", th.Color("recent", th.Palette.Warning)))
		count := 0
		for _, line := range rawLines {
			if line == "" || strings.HasPrefix(line, "wtmp") || strings.HasPrefix(line, "reboot") {
				continue
			}
			if count >= 4 {
				break
			}
			lines = append(lines, "  "+th.Dim(truncate(strings.TrimSpace(line), 44)))
			count++
		}
	}

	if len(lines) == 0 {
		return "", nil
	}

	compact := fmt.Sprintf("%d sessions", len(lines))
	return r.Section("Logins", "logins", compact, lines), nil
}
