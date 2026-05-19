package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ams/mom/internal/module/render"
)

// ServicesModule displays the status of systemd services.
type ServicesModule struct {
	// Services is the user-selected list of services to monitor.
	// If empty, falls back to defaultMonitoredServices.
	Services []string
}

func (m *ServicesModule) Name() string           { return "services" }
func (m *ServicesModule) Title() string          { return "Services" }
func (m *ServicesModule) Description() string    { return "Status of systemd services" }
func (m *ServicesModule) Dependencies() []string { return []string{"systemctl"} }
func (m *ServicesModule) Available() bool        { return CheckDependency("systemctl") }
func (m *ServicesModule) DefaultEnabled() bool   { return false }

func (m *ServicesModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *ServicesModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *ServicesModule) Settings() []SettingDef {
	return []SettingDef{
		{Key: "services", Label: "Services to monitor", Type: SettingList, Default: []string{},
			Description: "Select which systemd services to show in the MOTD"},
	}
}

var defaultMonitoredServices = []string{
	"sshd", "nginx", "apache2", "docker", "podman", "ufw",
	"firewalld", "cron", "fail2ban", "postgresql", "mysql",
	"mariadb", "redis-server", "redis",
}

func (m *ServicesModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

func (m *ServicesModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	r := render.New(opts)
	th := r.Theme()
	services := m.Services
	if len(services) == 0 {
		services = defaultMonitoredServices
	}

	// Collect active services
	type svcStatus struct {
		name   string
		status string
	}
	var found []svcStatus
	for _, svc := range services {
		status := getServiceStatus(ctx, svc)
		if status == "" {
			continue
		}
		found = append(found, svcStatus{svc, status})
	}

	var sb strings.Builder

	switch r.Variant() {
	case render.VariantCompact:
		sb.WriteString(r.Header("Services", "services"))
		sb.WriteString("\n    ")
		for i, s := range found {
			color, _ := th.Status(s.status)
			sb.WriteString(th.Color(s.name, color))
			if i < len(found)-1 {
				sb.WriteString(th.Color(" · ", th.Palette.Subtle))
			}
		}
		if len(found) == 0 {
			sb.WriteString(th.Dim("none"))
		}

	case render.VariantBoxed:
		var content strings.Builder
		for _, s := range found {
			dot := r.StatusDot(s.status)
			color, label := th.Status(s.status)
			content.WriteString(fmt.Sprintf("%s %-16s  %s\n", dot,
				s.name,
				r.Badge(label, color)))
		}
		if len(found) == 0 {
			content.WriteString(th.Dim("no monitored services found"))
		}
		sb.WriteString(render.Indent(r.Box(strings.TrimRight(content.String(), "\n"), "Services"), "  "))

	case render.VariantPowerline:
		sb.WriteString(r.Header("Services", "services"))
		sb.WriteString("\n\n")
		for _, s := range found {
			sb.WriteString(r.PowerlineStatus(s.name, s.status) + "\n")
		}
		if len(found) == 0 {
			sb.WriteString("    " + th.Dim("no monitored services found"))
		}

	case render.VariantCards:
		var content strings.Builder
		for _, s := range found {
			color, label := th.Status(s.status)
			content.WriteString(fmt.Sprintf("  %-16s  %s\n", s.name, th.Color(label, color)))
		}
		if len(found) == 0 {
			content.WriteString("  " + th.Dim("no monitored services found"))
		}
		sb.WriteString(render.Indent(r.Card(strings.TrimRight(content.String(), "\n"), "Services"), "  "))

	default:
		sb.WriteString(r.Header("Services", "services"))
		sb.WriteString("\n\n")
		for _, s := range found {
			dot := r.StatusDot(s.status)
			color, _ := th.Status(s.status)
			sb.WriteString(fmt.Sprintf("    %s %s  %s\n",
				dot,
				th.Color(fmt.Sprintf("%-16s", s.name), th.Palette.Foreground),
				th.Color(s.status, color)))
		}
		if len(found) == 0 {
			sb.WriteString("    " + th.Dim("no monitored services found"))
		}
	}

	return sb.String(), nil
}

// ListSystemServices enumerates all installed systemd service unit files.
// Returns service names (without .service suffix) suitable for the TUI picker.
func ListSystemServices(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "systemctl", "list-unit-files",
		"--type=service", "--no-legend", "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("listing services: %w", err)
	}

	var services []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := strings.TrimSuffix(fields[0], ".service")
		if name == "" || strings.HasPrefix(name, "@") {
			continue
		}
		services = append(services, name)
	}
	return services, nil
}

func getServiceStatus(ctx context.Context, service string) string {
	cmd := exec.CommandContext(ctx, "systemctl", "is-active", service)
	output, _ := cmd.Output()
	status := strings.TrimSpace(string(output))

	if status == "" || status == "unknown" {
		checkCmd := exec.CommandContext(ctx, "systemctl", "cat", service)
		if err := checkCmd.Run(); err != nil {
			return ""
		}
		return "inactive"
	}
	return status
}
