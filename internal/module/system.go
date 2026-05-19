package module

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"strings"
	"time"
)

// SystemModule displays basic system information.
type SystemModule struct{}

func (m *SystemModule) Name() string        { return "system" }
func (m *SystemModule) Title() string       { return "System Information" }
func (m *SystemModule) Description() string { return "Hostname, kernel, uptime, shell, user" }
func (m *SystemModule) Dependencies() []string { return nil }
func (m *SystemModule) Available() bool     { return true }
func (m *SystemModule) DefaultEnabled() bool { return true }

func (m *SystemModule) Generate(ctx context.Context) (string, error) {
	hostname, _ := os.Hostname()
	kernel := readKernel()
	uptime := readUptime()
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "unknown"
	}
	username := "unknown"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}

	var sb strings.Builder
	sb.WriteString("┌─ System ─────────────────────────────┐\n")
	sb.WriteString(fmt.Sprintf("│ Hostname: %-27s │\n", hostname))
	sb.WriteString(fmt.Sprintf("│ Kernel:   %-27s │\n", truncate(kernel, 27)))
	sb.WriteString(fmt.Sprintf("│ Uptime:   %-27s │\n", uptime))
	sb.WriteString(fmt.Sprintf("│ Shell:    %-27s │\n", shell))
	sb.WriteString(fmt.Sprintf("│ User:     %-27s │\n", username))
	sb.WriteString("└───────────────────────────────────────┘")

	return sb.String(), nil
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
