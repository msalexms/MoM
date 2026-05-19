package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ams/mom/internal/module/render"
)

// ContainersModule displays running Docker/Podman containers.
type ContainersModule struct {
	Runtime string // "auto", "docker", or "podman"
}

func (m *ContainersModule) Name() string        { return "containers" }
func (m *ContainersModule) Title() string       { return "Containers" }
func (m *ContainersModule) Description() string { return "Running Docker/Podman containers" }

func (m *ContainersModule) Dependencies() []string {
	switch m.Runtime {
	case "docker":
		return []string{"docker"}
	case "podman":
		return []string{"podman"}
	default:
		return []string{"docker", "podman"}
	}
}

func (m *ContainersModule) Available() bool      { return CheckDependency("docker") || CheckDependency("podman") }
func (m *ContainersModule) DefaultEnabled() bool { return false }

func (m *ContainersModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact}
}
func (m *ContainersModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *ContainersModule) Settings() []SettingDef {
	return []SettingDef{
		{Key: "runtime", Label: "Runtime", Type: SettingEnum, Default: "auto", Options: []string{"auto", "docker", "podman"}},
	}
}

func (m *ContainersModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

func (m *ContainersModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	runtime := m.detectRuntime()
	if runtime == "" {
		return "", nil
	}

	cmd := exec.CommandContext(ctx, runtime, "ps", "--format", "{{.Names}}\t{{.Status}}\t{{.Image}}")
	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	r := render.New(opts)
	var sb strings.Builder

	header := fmt.Sprintf("Containers (%s)", runtime)
	sb.WriteString(r.HeaderColor(header, r.Theme().SectionColor("containers")))
	sb.WriteString("\n\n")

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		sb.WriteString("    " + r.Theme().Dim("no running containers"))
		return sb.String(), nil
	}

	if len(lines) > 10 {
		lines = lines[:10]
	}

	th := r.Theme()
	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 3)
		name := parts[0]
		status := ""
		if len(parts) > 1 {
			status = parts[1]
		}

		statusKey := "active"
		if strings.Contains(strings.ToLower(status), "exited") {
			statusKey = "inactive"
		}
		dot := r.StatusDot(statusKey)
		sb.WriteString(fmt.Sprintf("    %s %s  %s\n",
			dot,
			th.Color(fmt.Sprintf("%-18s", truncate(name, 18)), th.Palette.Foreground),
			th.Dim(truncate(status, 24))))
	}

	return sb.String(), nil
}

func (m *ContainersModule) detectRuntime() string {
	switch m.Runtime {
	case "docker":
		if CheckDependency("docker") {
			return "docker"
		}
	case "podman":
		if CheckDependency("podman") {
			return "podman"
		}
	default:
		if CheckDependency("docker") {
			return "docker"
		}
		if CheckDependency("podman") {
			return "podman"
		}
	}
	return ""
}
