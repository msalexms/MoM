package module

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/msalexms/MoM/internal/i18n"

	"github.com/msalexms/MoM/internal/module/render"
)

// LastBootModule displays last boot/shutdown times and system age.
type LastBootModule struct{}

func (m *LastBootModule) Name() string           { return "boot" }
func (m *LastBootModule) Title() string          { return "Boot" }
func (m *LastBootModule) Description() string    { return "Last boot and system age" }
func (m *LastBootModule) Dependencies() []string { return nil }
func (m *LastBootModule) Available() bool        { return true }

func (m *LastBootModule) Generate(ctx context.Context, opts render.Options) (string, error) {
	bootTime := getBootTime()
	installAge := getInstallAge()

	if bootTime == "" {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()

	var lines []string
	lines = append(lines, r.Field("last boot", bootTime))
	if installAge != "" {
		lines = append(lines, r.Field("installed", th.Color(installAge, th.Palette.Success)))
	}

	compact := fmt.Sprintf(i18n.T("booted %s"), bootTime)
	if installAge != "" {
		compact += fmt.Sprintf(i18n.T("  installed %s ago"), installAge)
	}

	return r.Section("Boot", "boot", compact, lines), nil
}

func getBootTime() string {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return ""
	}
	var seconds float64
	fmt.Sscanf(string(data), "%f", &seconds)
	bootAt := time.Now().Add(-time.Duration(seconds) * time.Second)
	return bootAt.Format("2006-01-02 15:04")
}

func getInstallAge() string {
	// Approximate the install date from the oldest of a few system files.
	// ModTime is a heuristic: it changes if the file is touched, but for
	// machine-id/installer dirs it is stable enough.
	paths := []string{"/etc/machine-id", "/var/log/installer", "/root"}
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		age := time.Since(info.ModTime())
		days := int(age.Hours() / 24)
		if days > 365 {
			years := days / 365
			months := (days % 365) / 30
			return fmt.Sprintf("%dy %dm", years, months)
		}
		if days > 30 {
			return fmt.Sprintf("%dm %dd", days/30, days%30)
		}
		return fmt.Sprintf("%dd", days)
	}
	return ""
}
