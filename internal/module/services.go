package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ServicesModule displays the status of common systemd services.
type ServicesModule struct{}

func (m *ServicesModule) Name() string        { return "services" }
func (m *ServicesModule) Title() string       { return "Services" }
func (m *ServicesModule) Description() string { return "Status of systemd services" }
func (m *ServicesModule) Dependencies() []string { return []string{"systemctl"} }
func (m *ServicesModule) Available() bool     { return CheckDependency("systemctl") }
func (m *ServicesModule) DefaultEnabled() bool { return false }

// monitoredServices is the default list of services to check.
var monitoredServices = []string{
	"sshd",
	"nginx",
	"docker",
	"ufw",
	"cron",
	"fail2ban",
	"postgresql",
	"mysql",
	"redis",
}

func (m *ServicesModule) Generate(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var sb strings.Builder
	sb.WriteString("┌─ Services ───────────────────────────┐\n")

	found := false
	for _, svc := range monitoredServices {
		status := getServiceStatus(ctx, svc)
		if status == "" {
			continue // Service doesn't exist
		}
		found = true

		icon := "●"
		switch status {
		case "active":
			icon = "\033[32m●\033[0m" // Green
		case "inactive":
			icon = "\033[31m●\033[0m" // Red
		case "failed":
			icon = "\033[33m●\033[0m" // Yellow
		}

		entry := fmt.Sprintf("%s %s: %s", icon, svc, status)
		// Account for ANSI escape codes in length calculation
		displayLen := len(fmt.Sprintf("● %s: %s", svc, status))
		pad := 37 - displayLen
		if pad < 0 {
			pad = 0
		}
		sb.WriteString(fmt.Sprintf("│ %s%s │\n", entry, strings.Repeat(" ", pad)))
	}

	if !found {
		sb.WriteString("│ No monitored services found           │\n")
	}

	sb.WriteString("└───────────────────────────────────────┘")
	return sb.String(), nil
}

func getServiceStatus(ctx context.Context, service string) string {
	cmd := exec.CommandContext(ctx, "systemctl", "is-active", service)
	output, _ := cmd.Output()
	status := strings.TrimSpace(string(output))

	// "inactive" means it exists but is stopped
	// empty or "unknown" means it doesn't exist
	if status == "" || status == "unknown" {
		// Check if service exists at all
		checkCmd := exec.CommandContext(ctx, "systemctl", "cat", service)
		if err := checkCmd.Run(); err != nil {
			return "" // Service doesn't exist
		}
		return "inactive"
	}

	return status
}
