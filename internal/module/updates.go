package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ams/mom/internal/distro"
)

// UpdatesModule displays the number of pending package updates.
type UpdatesModule struct {
	Distro     distro.Family
	IncludeAUR bool
}

func (m *UpdatesModule) Name() string        { return "updates" }
func (m *UpdatesModule) Title() string       { return "Pending Updates" }
func (m *UpdatesModule) Description() string { return "Number of packages pending update" }

func (m *UpdatesModule) Dependencies() []string {
	switch m.Distro {
	case distro.FamilyDebian:
		return []string{"apt"}
	case distro.FamilyRHEL:
		return []string{"dnf"}
	case distro.FamilyArch:
		return []string{"pacman"}
	case distro.FamilySUSE:
		return []string{"zypper"}
	default:
		return nil
	}
}

func (m *UpdatesModule) Available() bool {
	return CheckDependency("apt") || CheckDependency("dnf") ||
		CheckDependency("pacman") || CheckDependency("zypper")
}

func (m *UpdatesModule) DefaultEnabled() bool { return false }

func (m *UpdatesModule) Generate(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	count := m.getUpdateCount(ctx)

	var sb strings.Builder
	sb.WriteString("┌─ Updates ────────────────────────────┐\n")
	if count == 0 {
		sb.WriteString("│ System is up to date                  │\n")
	} else {
		msg := fmt.Sprintf("%d package(s) can be updated", count)
		sb.WriteString(fmt.Sprintf("│ %-37s │\n", msg))
	}
	sb.WriteString("└───────────────────────────────────────┘")

	return sb.String(), nil
}

func (m *UpdatesModule) getUpdateCount(ctx context.Context) int {
	switch m.Distro {
	case distro.FamilyDebian:
		return m.countApt(ctx)
	case distro.FamilyRHEL:
		return m.countDnf(ctx)
	case distro.FamilyArch:
		return m.countPacman(ctx)
	case distro.FamilySUSE:
		return m.countZypper(ctx)
	default:
		// Try to auto-detect
		if CheckDependency("apt") {
			return m.countApt(ctx)
		}
		if CheckDependency("dnf") {
			return m.countDnf(ctx)
		}
		if CheckDependency("pacman") {
			return m.countPacman(ctx)
		}
		if CheckDependency("zypper") {
			return m.countZypper(ctx)
		}
		return 0
	}
}

func (m *UpdatesModule) countApt(ctx context.Context) int {
	cmd := exec.CommandContext(ctx, "apt", "list", "--upgradable")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	// First line is "Listing..." header
	if len(lines) <= 1 {
		return 0
	}
	return len(lines) - 1
}

func (m *UpdatesModule) countDnf(ctx context.Context) int {
	cmd := exec.CommandContext(ctx, "dnf", "check-update", "-q")
	output, _ := cmd.Output() // dnf returns exit code 100 when updates available
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

func (m *UpdatesModule) countPacman(ctx context.Context) int {
	// Try checkupdates first (safer, doesn't need root)
	if CheckDependency("checkupdates") {
		cmd := exec.CommandContext(ctx, "checkupdates")
		output, _ := cmd.Output()
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) == 1 && lines[0] == "" {
			return 0
		}
		count := len(lines)

		// Add AUR updates if configured
		if m.IncludeAUR {
			count += m.countAUR(ctx)
		}
		return count
	}

	// Fallback to pacman -Qu
	cmd := exec.CommandContext(ctx, "pacman", "-Qu")
	output, _ := cmd.Output()
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0
	}
	return len(lines)
}

func (m *UpdatesModule) countAUR(ctx context.Context) int {
	var cmd *exec.Cmd
	if CheckDependency("yay") {
		cmd = exec.CommandContext(ctx, "yay", "-Qua")
	} else if CheckDependency("paru") {
		cmd = exec.CommandContext(ctx, "paru", "-Qua")
	} else {
		return 0
	}
	output, _ := cmd.Output()
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0
	}
	return len(lines)
}

func (m *UpdatesModule) countZypper(ctx context.Context) int {
	cmd := exec.CommandContext(ctx, "zypper", "list-updates")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	lines := strings.Split(string(output), "\n")
	count := 0
	inTable := false
	for _, line := range lines {
		if strings.HasPrefix(line, "---") {
			inTable = true
			continue
		}
		if inTable && strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}
