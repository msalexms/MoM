package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ams/mom/internal/distro"
	"github.com/ams/mom/internal/module/render"
)

// UpdatesModule displays the number of pending package updates.
type UpdatesModule struct {
	Distro     distro.Family
	IncludeAUR bool
}

func (m *UpdatesModule) Name() string         { return "updates" }
func (m *UpdatesModule) Title() string        { return "Updates" }
func (m *UpdatesModule) Description() string  { return "Number of packages pending update" }
func (m *UpdatesModule) DefaultEnabled() bool { return false }

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

func (m *UpdatesModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *UpdatesModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *UpdatesModule) Settings() []SettingDef {
	return []SettingDef{
		{Key: "include_aur", Label: "Include AUR", Type: SettingBool, Default: false},
	}
}

func (m *UpdatesModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

func (m *UpdatesModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	count := m.getUpdateCount(ctx)
	r := render.New(opts)
	th := r.Theme()

	var lines []string
	var compact string
	if count == 0 {
		lines = append(lines, th.Color("+ system up to date", th.Palette.Success))
		compact = r.Icon("check") + " " + th.Color("up to date", th.Palette.Success)
	} else {
		color := th.Palette.Warning
		if count > 50 {
			color = th.Palette.Danger
		}
		lines = append(lines, th.Color(fmt.Sprintf("! %d packages pending", count), color))
		compact = r.Icon("update") + " " + th.Color(fmt.Sprintf("%d pending", count), color)
	}

	return r.Section("Updates", "updates", compact, lines), nil
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
	if len(lines) <= 1 {
		return 0
	}
	return len(lines) - 1
}

func (m *UpdatesModule) countDnf(ctx context.Context) int {
	cmd := exec.CommandContext(ctx, "dnf", "check-update", "-q")
	output, _ := cmd.Output()
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
	if CheckDependency("checkupdates") {
		cmd := exec.CommandContext(ctx, "checkupdates")
		output, _ := cmd.Output()
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) == 1 && lines[0] == "" {
			return 0
		}
		count := len(lines)
		if m.IncludeAUR {
			count += m.countAUR(ctx)
		}
		return count
	}
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
