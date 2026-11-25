package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/msalexms/MoM/internal/i18n"

	"github.com/msalexms/MoM/internal/config"
	"github.com/msalexms/MoM/internal/module/render"
)

// ContainersModule displays running Docker/Podman containers.
type ContainersModule struct {
	Runtime string // "auto", "docker", or "podman"
}

func (m *ContainersModule) Name() string  { return "containers" }
func (m *ContainersModule) Title() string { return "Containers" }

// UpdateFromConfig implements ConfigSync.
func (m *ContainersModule) UpdateFromConfig(cfg *config.Config) {
	m.Runtime = cfg.Modules.ContainersConfig.Runtime
}
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
func (m *ContainersModule) Settings() []Setting {
	return []Setting{
		enumSetting("runtime", "Engine", "auto uses docker when installed, otherwise podman",
			[]string{"auto", "docker", "podman"},
			func(c *config.Config) *string { return &c.Modules.ContainersConfig.Runtime }),
	}
}

func (m *ContainersModule) Generate(ctx context.Context, opts render.Options) (string, error) {
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

	header := fmt.Sprintf(i18n.T("Containers (%s)"), runtime)
	headerColor := th.SectionColor("containers")

	// Build content lines
	var lines []string
	if len(rawLines) == 0 {
		lines = append(lines, th.Dim(i18n.T("no running containers")))
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

	compact := fmt.Sprintf(i18n.T("%s %d running"), r.Icon("docker"), len(rawLines))
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
