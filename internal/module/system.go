package module

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/ams/mom/internal/module/render"
)

// SystemModule displays basic system information.
type SystemModule struct{}

func (m *SystemModule) Name() string           { return "system" }
func (m *SystemModule) Title() string          { return "System" }
func (m *SystemModule) Description() string    { return "Hostname, kernel, uptime, shell, user" }
func (m *SystemModule) Dependencies() []string { return nil }
func (m *SystemModule) Available() bool        { return true }
func (m *SystemModule) DefaultEnabled() bool   { return true }

func (m *SystemModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *SystemModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *SystemModule) Settings() []SettingDef         { return nil }

func (m *SystemModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

func (m *SystemModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	r := render.New(opts)
	th := r.Theme()
	hostname, _ := os.Hostname()
	kernel := readKernel()
	uptime := readUptime()
	shell := trimShellPath(os.Getenv("SHELL"))
	if shell == "" {
		shell = "unknown"
	}
	username := "unknown"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}

	var sb strings.Builder

	switch r.Variant() {
	case render.VariantCompact:
		sb.WriteString(r.Header("System", "system"))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("    %s %s@%s %s %s %s up %s",
			r.Icon("user"), username, hostname,
			th.Color("│", th.Palette.Subtle),
			kernel,
			th.Color("│", th.Palette.Subtle),
			th.Color(uptime, th.Palette.Success)))

	case render.VariantBoxed:
		var content strings.Builder
		content.WriteString(fmt.Sprintf("%-4s %-6s  %s\n", r.Icon("host"), "host", hostname))
		content.WriteString(fmt.Sprintf("%-4s %-6s  %s\n", r.Icon("kernel"), "kern", kernel))
		content.WriteString(fmt.Sprintf("%-4s %-6s  %s\n", r.Icon("uptime"), "up", th.Color(uptime, th.Palette.Success)))
		content.WriteString(fmt.Sprintf("%-4s %-6s  %s\n", r.Icon("shell"), "sh", shell))
		content.WriteString(fmt.Sprintf("%-4s %-6s  %s", r.Icon("user"), "user", th.Color(username, th.Palette.Secondary)))
		sb.WriteString(render.Indent(r.Box(content.String(), "System"), "  "))

	case render.VariantPowerline:
		sb.WriteString(r.Header("System", "system"))
		sb.WriteString("\n\n")
		sb.WriteString(r.PowerlineRow([][]string{{"host", hostname}, {"kern", kernel}}) + "\n")
		sb.WriteString(r.PowerlineRow([][]string{{"up", uptime}, {"shell", shell}, {"user", username}}))

	case render.VariantCards:
		var content strings.Builder
		content.WriteString(fmt.Sprintf("  %-8s  %s\n", "host", hostname))
		content.WriteString(fmt.Sprintf("  %-8s  %s\n", "kernel", kernel))
		content.WriteString(fmt.Sprintf("  %-8s  %s\n", "uptime", th.Color(uptime, th.Palette.Success)))
		content.WriteString(fmt.Sprintf("  %-8s  %s\n", "shell", shell))
		content.WriteString(fmt.Sprintf("  %-8s  %s", "user", th.Color(username, th.Palette.Secondary)))
		sb.WriteString(render.Indent(r.Card(content.String(), "System"), "  "))

	case render.VariantMinimal:
		sb.WriteString(fmt.Sprintf("  %s@%s  %s  up %s", username, hostname, kernel, uptime))

	default:
		sb.WriteString(r.Header("System", "system"))
		sb.WriteString("\n\n")
		sb.WriteString(r.KeyValue("host", hostname) + "\n")
		sb.WriteString(r.KeyValue("kernel", kernel) + "\n")
		sb.WriteString(r.KeyValue("uptime", uptime) + "\n")
		sb.WriteString(r.KeyValue("shell", shell) + "\n")
		sb.WriteString(r.KeyValue("user", username))
	}

	return sb.String(), nil
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
