package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ams/mom/internal/module/render"
)

// TimersModule displays upcoming systemd timers.
type TimersModule struct{}

func (m *TimersModule) Name() string           { return "timers" }
func (m *TimersModule) Title() string          { return "Systemd Timers" }
func (m *TimersModule) Description() string    { return "Next scheduled systemd timer activations" }
func (m *TimersModule) Dependencies() []string { return []string{"systemctl"} }
func (m *TimersModule) Available() bool        { return CheckDependency("systemctl") }
func (m *TimersModule) DefaultEnabled() bool   { return false }
func (m *TimersModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *TimersModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *TimersModule) Settings() []SettingDef         { return nil }

func (m *TimersModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

func (m *TimersModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "systemctl", "list-timers", "--no-pager", "--no-legend", "-n", "5")
	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	rawLines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(rawLines) == 0 || rawLines[0] == "" {
		return "", nil
	}

	type timer struct {
		next string
		unit string
	}
	var timers []timer
	for _, line := range rawLines {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		// Fields: NEXT(3 fields: day date time) LEFT LAST(3) PASSED UNIT ACTIVATES
		next := strings.Join(fields[:3], " ")
		unit := fields[len(fields)-1]
		unit = strings.TrimSuffix(unit, ".timer")
		timers = append(timers, timer{next: next, unit: unit})
		if len(timers) >= 5 {
			break
		}
	}

	if len(timers) == 0 {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()

	var lines []string
	for _, t := range timers {
		lines = append(lines, fmt.Sprintf("%-20s  %s", th.Color(t.unit, th.Palette.Warning), th.Dim(t.next)))
	}

	compact := fmt.Sprintf("%d scheduled", len(timers))
	return r.Section("Timers", "timers", compact, lines), nil
}
