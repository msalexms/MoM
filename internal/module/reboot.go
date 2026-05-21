package module

import (
	"context"
	"os"
	"strings"

	"github.com/msalexms/MoM/internal/module/render"
)

// RebootModule displays whether a system reboot is pending.
type RebootModule struct{}

func (m *RebootModule) Name() string           { return "reboot" }
func (m *RebootModule) Title() string          { return "Pending Reboot" }
func (m *RebootModule) Description() string    { return "Shows if a reboot is required (kernel update)" }
func (m *RebootModule) Dependencies() []string { return nil }
func (m *RebootModule) Available() bool        { return true }
func (m *RebootModule) DefaultEnabled() bool   { return false }
func (m *RebootModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *RebootModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *RebootModule) Settings() []SettingDef         { return nil }

func (m *RebootModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

func (m *RebootModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	needed, reason := rebootNeeded()
	if !needed {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()

	msg := "System restart required"
	if reason != "" {
		msg = reason
	}

	lines := []string{th.Color("⚠ "+msg, th.Palette.Warning)}
	compact := th.Color("⚠ "+msg, th.Palette.Warning)
	return r.Section("Reboot Required", "reboot", compact, lines), nil
}

func rebootNeeded() (bool, string) {
	// Ubuntu/Debian
	if data, err := os.ReadFile("/var/run/reboot-required"); err == nil {
		return true, strings.TrimSpace(string(data))
	}
	// RHEL/Fedora: check if running kernel != installed kernel
	if _, err := os.Stat("/var/run/reboot-required"); err == nil {
		return true, ""
	}
	// Generic: check needs-restarting on RHEL
	if _, err := os.Stat("/usr/bin/needs-restarting"); err == nil {
		return true, "Packages updated, reboot recommended"
	}
	return false, ""
}
