package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
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

func (m *ContainersModule) Generate(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	runtime := m.detectRuntime()
	if runtime == "" {
		return "", nil
	}

	cmd := exec.CommandContext(ctx, runtime, "ps", "--format", "{{.Names}}\t{{.Status}}")
	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return fmt.Sprintf("┌─ Containers (%s) ─────────────────────┐\n│ No running containers                 │\n└───────────────────────────────────────┘", runtime), nil
	}

	// Limit to 10 containers
	if len(lines) > 10 {
		lines = lines[:10]
	}

	var sb strings.Builder
	header := fmt.Sprintf("Containers (%s)", runtime)
	sb.WriteString(fmt.Sprintf("┌─ %-36s ┐\n", header))
	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 2)
		name := parts[0]
		status := ""
		if len(parts) > 1 {
			status = parts[1]
		}
		entry := fmt.Sprintf("%s: %s", truncate(name, 15), truncate(status, 18))
		sb.WriteString(fmt.Sprintf("│ %-37s │\n", entry))
	}
	sb.WriteString("└───────────────────────────────────────┘")

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
