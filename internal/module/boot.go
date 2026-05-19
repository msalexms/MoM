package module

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ams/mom/internal/module/render"
)

// LastBootModule displays last boot/shutdown times and system age.
type LastBootModule struct{}

func (m *LastBootModule) Name() string           { return "boot" }
func (m *LastBootModule) Title() string          { return "Boot Info" }
func (m *LastBootModule) Description() string    { return "Last boot time, system install age" }
func (m *LastBootModule) Dependencies() []string { return nil }
func (m *LastBootModule) Available() bool        { return true }
func (m *LastBootModule) DefaultEnabled() bool   { return false }

func (m *LastBootModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *LastBootModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *LastBootModule) Settings() []SettingDef         { return nil }

func (m *LastBootModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

func (m *LastBootModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	bootTime := getBootTime()
	installAge := getInstallAge()

	if bootTime == "" {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()
	var sb strings.Builder

	switch r.Variant() {
	case render.VariantCompact:
		sb.WriteString(r.Header("Boot", "boot"))
		sb.WriteString(fmt.Sprintf("\n    booted %s", bootTime))
		if installAge != "" {
			sb.WriteString(fmt.Sprintf("  age %s", installAge))
		}

	case render.VariantBoxed:
		var content strings.Builder
		content.WriteString(fmt.Sprintf("%-10s  %s", "last boot", bootTime))
		if installAge != "" {
			content.WriteString(fmt.Sprintf("\n%-10s  %s", "sys age", th.Color(installAge, th.Palette.Success)))
		}
		sb.WriteString(render.Indent(r.Box(content.String(), "Boot Info"), "  "))

	case render.VariantPowerline:
		sb.WriteString(r.Header("Boot Info", "boot"))
		sb.WriteString("\n\n")
		sb.WriteString(fmt.Sprintf("    %s %-10s %s\n",
			th.Color("▌", th.Palette.Accent), th.Color("boot", th.Palette.Warning), bootTime))
		if installAge != "" {
			sb.WriteString(fmt.Sprintf("    %s %-10s %s",
				th.Color("▌", th.Palette.Accent), th.Color("age", th.Palette.Warning),
				th.Color(installAge, th.Palette.Success)))
		}

	case render.VariantCards:
		var content strings.Builder
		content.WriteString(fmt.Sprintf("  %-10s  %s", "last boot", bootTime))
		if installAge != "" {
			content.WriteString(fmt.Sprintf("\n  %-10s  %s", "sys age", th.Color(installAge, th.Palette.Success)))
		}
		sb.WriteString(render.Indent(r.Card(content.String(), "Boot Info"), "  "))

	default:
		sb.WriteString(r.Header("Boot Info", "boot"))
		sb.WriteString("\n\n")
		sb.WriteString(r.KeyValue("last boot", bootTime))
		if installAge != "" {
			sb.WriteString("\n" + r.KeyValue("sys age", installAge))
		}
	}

	return sb.String(), nil
}

func getBootTime() string {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return ""
	}
	var seconds float64
	fmt.Sscanf(string(data), "%f", &seconds)
	bootAt := time.Now().Add(-time.Duration(seconds) * time.Second)
	return bootAt.Format("2006-01-02 15:04")
}

func getInstallAge() string {
	// Use filesystem creation time of /etc/machine-id or /var/log/installer
	paths := []string{"/etc/machine-id", "/var/log/installer", "/root"}
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		age := time.Since(info.ModTime())
		days := int(age.Hours() / 24)
		if days > 365 {
			years := days / 365
			months := (days % 365) / 30
			return fmt.Sprintf("%dy %dm", years, months)
		}
		if days > 30 {
			return fmt.Sprintf("%dm %dd", days/30, days%30)
		}
		return fmt.Sprintf("%dd", days)
	}
	return ""
}
