package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ams/mom/internal/module/render"
)

// PortsModule displays listening TCP ports.
type PortsModule struct{}

func (m *PortsModule) Name() string           { return "ports" }
func (m *PortsModule) Title() string          { return "Listening Ports" }
func (m *PortsModule) Description() string    { return "TCP ports currently in LISTEN state" }
func (m *PortsModule) Dependencies() []string { return []string{"ss"} }
func (m *PortsModule) Available() bool        { return CheckDependency("ss") }
func (m *PortsModule) DefaultEnabled() bool   { return false }

func (m *PortsModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *PortsModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *PortsModule) Settings() []SettingDef         { return nil }

func (m *PortsModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

func (m *PortsModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ss", "-tlnp")
	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	type portInfo struct {
		port    string
		process string
	}

	var ports []portInfo
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines[1:] { // skip header
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		local := fields[3]
		// Extract port from address like *:22 or 0.0.0.0:80
		port := local
		if idx := strings.LastIndex(local, ":"); idx >= 0 {
			port = local[idx+1:]
		}
		proc := ""
		if len(fields) >= 6 {
			proc = extractProcessName(fields[5])
		}
		ports = append(ports, portInfo{port, proc})
	}

	if len(ports) == 0 {
		return "", nil
	}
	if len(ports) > 8 {
		ports = ports[:8]
	}

	r := render.New(opts)
	th := r.Theme()
	var sb strings.Builder

	switch r.Variant() {
	case render.VariantCompact:
		sb.WriteString(r.Header("Ports", "ports"))
		sb.WriteString("\n    ")
		var ps []string
		for _, p := range ports {
			ps = append(ps, p.port)
		}
		sb.WriteString(strings.Join(ps, th.Color(" · ", th.Palette.Subtle)))

	case render.VariantBoxed:
		var content strings.Builder
		for _, p := range ports {
			content.WriteString(fmt.Sprintf("%-6s  %s\n", p.port, th.Dim(p.process)))
		}
		sb.WriteString(render.Indent(r.Box(strings.TrimRight(content.String(), "\n"), "Ports"), "  "))

	case render.VariantPowerline:
		sb.WriteString(r.Header("Ports", "ports"))
		sb.WriteString("\n\n")
		for _, p := range ports {
			sb.WriteString(fmt.Sprintf("    %s %-8s %s\n",
				th.Color("▌", th.Palette.Accent),
				th.Color(p.port, th.Palette.Warning),
				th.Dim(p.process)))
		}

	case render.VariantCards:
		var content strings.Builder
		for _, p := range ports {
			content.WriteString(fmt.Sprintf("  %-6s  %s\n", p.port, th.Dim(p.process)))
		}
		sb.WriteString(render.Indent(r.Card(strings.TrimRight(content.String(), "\n"), "Ports"), "  "))

	default:
		sb.WriteString(r.Header("Ports", "ports"))
		sb.WriteString("\n\n")
		for _, p := range ports {
			sb.WriteString(fmt.Sprintf("    %-6s  %s\n", th.Color(p.port, th.Palette.Success), th.Dim(p.process)))
		}
	}

	return sb.String(), nil
}

func extractProcessName(field string) string {
	// Format: users:(("sshd",pid=1234,fd=3))
	if idx := strings.Index(field, "((\""); idx >= 0 {
		rest := field[idx+3:]
		if end := strings.Index(rest, "\""); end >= 0 {
			return rest[:end]
		}
	}
	return ""
}
