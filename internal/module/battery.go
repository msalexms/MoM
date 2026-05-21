package module

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/msalexms/MoM/internal/module/render"
)

// BatteryModule displays battery status on laptops.
type BatteryModule struct{}

func (m *BatteryModule) Name() string           { return "battery" }
func (m *BatteryModule) Title() string          { return "Battery" }
func (m *BatteryModule) Description() string    { return "Battery level and charging status" }
func (m *BatteryModule) Dependencies() []string { return nil }
func (m *BatteryModule) DefaultEnabled() bool   { return false }

func (m *BatteryModule) Available() bool {
	_, err := os.Stat("/sys/class/power_supply/BAT0/capacity")
	return err == nil
}

func (m *BatteryModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *BatteryModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *BatteryModule) Settings() []SettingDef         { return nil }

func (m *BatteryModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

func (m *BatteryModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	capacity := readFileInt("/sys/class/power_supply/BAT0/capacity")
	if capacity < 0 {
		return "", nil
	}

	statusRaw := strings.TrimSpace(readFileStr("/sys/class/power_supply/BAT0/status"))
	charging := statusRaw == "Charging" || statusRaw == "Full"

	// Time to empty/full
	timeLeft := getBatteryTime()

	r := render.New(opts)
	th := r.Theme()

	pct := float64(capacity)
	statusLabel := "discharging"
	statusColor := th.PercentColor(100 - pct) // invert: low battery = danger
	if charging {
		statusLabel = "charging"
		statusColor = th.Palette.Success
	}
	if statusRaw == "Full" {
		statusLabel = "full"
		statusColor = th.Palette.Success
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("%-10s  %s", th.Color("level", th.Palette.Warning), r.ProgressBar(pct, 24, fmt.Sprintf("%d%%", capacity))))
	lines = append(lines, fmt.Sprintf("%-10s  %s", th.Color("status", th.Palette.Warning), th.Color(statusLabel, statusColor)))
	if timeLeft != "" {
		lines = append(lines, fmt.Sprintf("%-10s  %s", th.Color("remaining", th.Palette.Warning), timeLeft))
	}

	compact := fmt.Sprintf("%d%% %s", capacity, th.Color(statusLabel, statusColor))
	if timeLeft != "" {
		compact += " (" + timeLeft + ")"
	}

	return r.Section("Battery", "battery", compact, lines), nil
}

func readFileInt(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return -1
	}
	var v int
	fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &v)
	return v
}

func readFileStr(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func getBatteryTime() string {
	// energy_now and power_now in µWh and µW
	energyNow := readFileInt("/sys/class/power_supply/BAT0/energy_now")
	powerNow := readFileInt("/sys/class/power_supply/BAT0/power_now")
	if energyNow <= 0 || powerNow <= 0 {
		return ""
	}
	hours := float64(energyNow) / float64(powerNow)
	d := time.Duration(hours * float64(time.Hour))
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
