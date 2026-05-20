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

func (m *ContainersModule) Available() bool {
	return CheckDependency("docker") || CheckDependency("podman")
}
func (m *ContainersModule) DefaultEnabled() bool { return false }

func (m *ContainersModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
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
	th := r.Theme()

	rawLines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(rawLines) == 0 || (len(rawLines) == 1 && rawLines[0] == "") {
		rawLines = nil
	}
	if len(rawLines) > 10 {
		rawLines = rawLines[:10]
	}

	header := fmt.Sprintf("Containers (%s)", runtime)
	headerColor := th.SectionColor("containers")

	// Build content lines
	var lines []string
	if len(rawLines) == 0 {
		lines = append(lines, th.Dim("no running containers"))
	} else {
		for _, line := range rawLines {
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
			lines = append(lines, fmt.Sprintf("%s %s  %s", dot,
				th.Color(fmt.Sprintf("%-18s", truncate(name, 18)), th.Palette.Foreground),
				th.Dim(truncate(status, 24))))
		}
	}

	compact := fmt.Sprintf("%s %d running", r.Icon("docker"), len(rawLines))
	return r.SectionCustom(header, headerColor, compact, lines), nil
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
