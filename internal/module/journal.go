package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ams/mom/internal/module/render"
)

// JournalModule displays recent error-level journal entries.
type JournalModule struct{}

func (m *JournalModule) Name() string           { return "journal" }
func (m *JournalModule) Title() string          { return "Journal Errors" }
func (m *JournalModule) Description() string    { return "Error-level journal entries from last 24h" }
func (m *JournalModule) Dependencies() []string { return []string{"journalctl"} }
func (m *JournalModule) Available() bool        { return CheckDependency("journalctl") }
func (m *JournalModule) DefaultEnabled() bool   { return false }
func (m *JournalModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *JournalModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *JournalModule) Settings() []SettingDef         { return nil }

func (m *JournalModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

func (m *JournalModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "journalctl", "-p", "err", "--since", "24h ago", "--no-pager", "-n", "5", "-o", "short")
	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || lines[0] == "" || strings.Contains(lines[0], "No entries") {
		return "", nil
	}
	if len(lines) > 5 {
		lines = lines[:5]
	}

	r := render.New(opts)
	th := r.Theme()
	var sb strings.Builder

	switch r.Variant() {
	case render.VariantCompact:
		sb.WriteString(r.Header("Journal", "journal"))
		sb.WriteString(fmt.Sprintf("\n    %s %d errors in 24h", th.Color("⚠", th.Palette.Danger), len(lines)))
	case render.VariantBoxed:
		var c strings.Builder
		for _, l := range lines {
			c.WriteString(truncate(l, 44) + "\n")
		}
		sb.WriteString(render.Indent(r.Box(strings.TrimRight(c.String(), "\n"), "Journal Errors"), "  "))
	default:
		sb.WriteString(r.Header("Journal Errors", "journal"))
		sb.WriteString("\n\n")
		for _, l := range lines {
			sb.WriteString("    " + th.Color(truncate(l, 50), th.Palette.Danger) + "\n")
		}
	}
	return sb.String(), nil
}
