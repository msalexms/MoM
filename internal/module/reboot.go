package module

import (
	"context"
	"os"
	"strings"

	"github.com/msalexms/MoM/internal/i18n"

	"github.com/msalexms/MoM/internal/module/render"
)

// RebootModule displays whether a system reboot is pending.
type RebootModule struct{}

func (m *RebootModule) Name() string  { return "reboot" }
func (m *RebootModule) Title() string { return "Reboot required" }
func (m *RebootModule) Description() string {
	return "Shows when a reboot is needed (updated kernel)"
}
func (m *RebootModule) Dependencies() []string { return nil }
func (m *RebootModule) Available() bool        { return true }

func (m *RebootModule) Generate(ctx context.Context, opts render.Options) (string, error) {
	needed, reason := rebootNeeded()
	if !needed {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()

	msg := i18n.T("System restart required")
	if reason != "" {
		msg = reason
	}

	lines := []string{th.Color("⚠ "+msg, th.Palette.Warning)}
	compact := th.Color("⚠ "+msg, th.Palette.Warning)
	return r.Section("Reboot required", "reboot", compact, lines), nil
}

func rebootNeeded() (bool, string) {
	// Ubuntu/Debian: presence of this file means a reboot is required and
	// it usually carries a short reason message.
	if data, err := os.ReadFile("/var/run/reboot-required"); err == nil {
		return true, strings.TrimSpace(string(data))
	}
	// Generic: needs-restarting exists (RHEL-family tooling)
	if _, err := os.Stat("/usr/bin/needs-restarting"); err == nil {
		return true, i18n.T("Packages updated, reboot recommended")
	}
	return false, ""
}
