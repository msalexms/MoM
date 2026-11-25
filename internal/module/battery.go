package module

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/msalexms/MoM/internal/i18n"

	"github.com/msalexms/MoM/internal/module/render"
)

// BatteryModule displays battery status on laptops.
type BatteryModule struct{}

func (m *BatteryModule) Name() string           { return "battery" }
func (m *BatteryModule) Title() string          { return "Battery" }
func (m *BatteryModule) Description() string    { return "Battery level and charging state" }
func (m *BatteryModule) Dependencies() []string { return nil }
func (m *BatteryModule) Available() bool {
	_, err := os.Stat(batteryDir() + "/capacity")
	return err == nil
}

func (m *BatteryModule) Generate(ctx context.Context, opts render.Options) (string, error) {
	bat := batteryDir()
	capacity := readFileInt(bat + "/capacity")
	if capacity < 0 {
		return "", nil
	}

	statusRaw := strings.TrimSpace(readFileStr(bat + "/status"))
	charging := statusRaw == "Charging" || statusRaw == "Full"

	// Time to empty/full
	timeLeft := getBatteryTime(bat)

	r := render.New(opts)
	th := r.Theme()

	pct := float64(capacity)
	statusLabel := i18n.T("discharging")
	statusColor := th.PercentColor(100 - pct) // invert: low battery = danger
	if charging {
		statusLabel = i18n.T("charging")
		statusColor = th.Palette.Success
	}
	if statusRaw == "Full" {
		statusLabel = i18n.T("full")
		statusColor = th.Palette.Success
	}

	var lines []string
	lines = append(lines, r.Field("level", r.ProgressBar(pct, 24, fmt.Sprintf("%d%%", capacity))))
	lines = append(lines, r.Field("status", th.Color(statusLabel, statusColor)))
	if timeLeft != "" {
		lines = append(lines, r.Field("remaining", timeLeft))
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

// batteryDir returns the sysfs directory of the first battery that
// exposes a capacity file (BAT0, BAT1, …).
func batteryDir() string {
	entries, err := os.ReadDir("/sys/class/power_supply")
	if err != nil {
		return "/sys/class/power_supply/BAT0"
	}
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "BAT") {
			continue
		}
		dir := "/sys/class/power_supply/" + e.Name()
		if _, err := os.Stat(dir + "/capacity"); err == nil {
			return dir
		}
	}
	return "/sys/class/power_supply/BAT0"
}

func getBatteryTime(bat string) string {
	// energy_now and power_now in µWh and µW
	energyNow := readFileInt(bat + "/energy_now")
	powerNow := readFileInt(bat + "/power_now")
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
