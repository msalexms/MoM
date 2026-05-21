package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	embeds "github.com/msalexms/MoM/embed"
	"github.com/msalexms/MoM/internal/distro"
	"github.com/msalexms/MoM/internal/module/render"
)

// LogoModule displays the distro's ASCII logo.
type LogoModule struct {
	Distro distro.Info
}

func (m *LogoModule) Name() string           { return "logo" }
func (m *LogoModule) Title() string          { return "Distro Logo" }
func (m *LogoModule) Description() string    { return "ASCII art logo of your Linux distribution" }
func (m *LogoModule) Dependencies() []string { return nil }
func (m *LogoModule) Available() bool        { return true }
func (m *LogoModule) DefaultEnabled() bool   { return true }

func (m *LogoModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantMinimal}
}
func (m *LogoModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *LogoModule) Settings() []SettingDef         { return nil }

func (m *LogoModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

func (m *LogoModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	if opts.Variant == render.VariantMinimal {
		return fmt.Sprintf("  [%s]", m.Distro.Name), nil
	}

	if CheckDependency("fastfetch") {
		result, err := m.fromFastfetch(ctx)
		if err == nil && result != "" {
			return result, nil
		}
	}
	return m.fromEmbedded()
}

func (m *LogoModule) fromFastfetch(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "fastfetch", "--logo-only", "--logo-padding-top", "0")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(output), "\n"), nil
}

func (m *LogoModule) fromEmbedded() (string, error) {
	filename := m.logoFilename()
	data, err := embeds.LogosFS.ReadFile("logos/" + filename)
	if err != nil {
		data, err = embeds.LogosFS.ReadFile("logos/default.txt")
		if err != nil {
			return fmt.Sprintf("  [%s]", m.Distro.Name), nil
		}
	}
	return strings.TrimRight(string(data), "\n"), nil
}

func (m *LogoModule) logoFilename() string {
	switch m.Distro.Family {
	case distro.FamilyDebian:
		if strings.Contains(strings.ToLower(m.Distro.ID), "ubuntu") {
			return "ubuntu.txt"
		}
		return "debian.txt"
	case distro.FamilyRHEL:
		return "fedora.txt"
	case distro.FamilyArch:
		return "arch.txt"
	case distro.FamilySUSE:
		return "opensuse.txt"
	default:
		return "default.txt"
	}
}
