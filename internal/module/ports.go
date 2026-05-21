package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/msalexms/MoM/internal/module/render"
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

	var contentLines []string
	var compactParts []string
	for _, p := range ports {
		contentLines = append(contentLines, fmt.Sprintf("%-6s  %s", th.Color(p.port, th.Palette.Success), th.Dim(p.process)))
		compactParts = append(compactParts, p.port)
	}

	compact := strings.Join(compactParts, th.Color(" · ", th.Palette.Subtle))
	return r.Section("Ports", "ports", compact, contentLines), nil
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
