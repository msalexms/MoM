package module

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/msalexms/MoM/internal/i18n"

	"github.com/msalexms/MoM/internal/module/render"
)

// SystemModule displays basic system information.
type SystemModule struct{}

func (m *SystemModule) Name() string           { return "system" }
func (m *SystemModule) Title() string          { return "System" }
func (m *SystemModule) Description() string    { return "Host, kernel, uptime, shell and user" }
func (m *SystemModule) Dependencies() []string { return nil }
func (m *SystemModule) Available() bool        { return true }

func (m *SystemModule) Generate(ctx context.Context, opts render.Options) (string, error) {
	r := render.New(opts)
	th := r.Theme()
	hostname, _ := os.Hostname()
	kernel := readKernel()
	uptime := readUptime()
	shell := trimShellPath(os.Getenv("SHELL"))
	if shell == "" {
		shell = i18n.T("unknown")
	}
	username := i18n.T("unknown")
	if u, err := user.Current(); err == nil {
		username = u.Username
	}

	var sb strings.Builder

	// System's minimal variant is a one-liner, not a typical section
	if r.Variant() == render.VariantMinimal {
		sb.WriteString(fmt.Sprintf("  %s@%s  %s  up %s", username, hostname, kernel, uptime))
		return sb.String(), nil
	}

	lines := []string{
		r.Field("host", hostname),
		r.Field("kernel", kernel),
		r.Field("uptime", th.Color(uptime, th.Palette.Success)),
		r.Field("shell", shell),
		r.Field("user", th.Color(username, th.Palette.Secondary)),
	}

	compact := fmt.Sprintf("%s %s@%s %s %s %s up %s",
		r.Icon("user"), username, hostname,
		th.Color("│", th.Palette.Subtle),
		kernel,
		th.Color("│", th.Palette.Subtle),
		th.Color(uptime, th.Palette.Success))

	return r.Section("System", "system", compact, lines), nil
}

func trimShellPath(shell string) string {
	if idx := strings.LastIndex(shell, "/"); idx >= 0 {
		return shell[idx+1:]
	}
	return shell
}

func readKernel() string {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return "unknown"
	}
	parts := strings.Fields(string(data))
	if len(parts) >= 3 {
		return parts[0] + " " + parts[2]
	}
	return strings.TrimSpace(string(data))
}

func readUptime() string {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "unknown"
	}
	var seconds float64
	fmt.Sscanf(string(data), "%f", &seconds)

	d := time.Duration(seconds) * time.Second
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
