package module

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ams/mom/internal/module/render"
)

// SudoModule displays recent sudo activity from auth log.
type SudoModule struct{}

func (m *SudoModule) Name() string           { return "sudo" }
func (m *SudoModule) Title() string          { return "Sudo Activity" }
func (m *SudoModule) Description() string    { return "Recent sudo commands from auth log" }
func (m *SudoModule) Dependencies() []string { return nil }
func (m *SudoModule) DefaultEnabled() bool   { return false }
func (m *SudoModule) Available() bool {
	_, err := os.Stat("/var/log/auth.log")
	return err == nil
}
func (m *SudoModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *SudoModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *SudoModule) Settings() []SettingDef         { return nil }

func (m *SudoModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

type sudoEntry struct {
	user string
	cmd  string
}

func (m *SudoModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	entries := getRecentSudo(5)
	if len(entries) == 0 {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()
	var sb strings.Builder

	switch r.Variant() {
	case render.VariantCompact:
		sb.WriteString(r.Header("Sudo", "sudo"))
		sb.WriteString(fmt.Sprintf("\n    %d recent commands", len(entries)))
	case render.VariantBoxed:
		var c strings.Builder
		for _, e := range entries {
			c.WriteString(fmt.Sprintf("%-10s  %s\n", e.user, th.Dim(truncate(e.cmd, 30))))
		}
		sb.WriteString(render.Indent(r.Box(strings.TrimRight(c.String(), "\n"), "Sudo Activity"), "  "))
	default:
		sb.WriteString(r.Header("Sudo Activity", "sudo"))
		sb.WriteString("\n\n")
		for _, e := range entries {
			sb.WriteString(fmt.Sprintf("    %-10s  %s\n", th.Color(e.user, th.Palette.Warning), th.Dim(truncate(e.cmd, 34))))
		}
	}
	return sb.String(), nil
}

func getRecentSudo(n int) []sudoEntry {
	f, err := os.Open("/var/log/auth.log")
	if err != nil {
		return nil
	}
	defer f.Close()

	var all []sudoEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "sudo") || !strings.Contains(line, "COMMAND=") {
			continue
		}
		user := ""
		cmd := ""
		if idx := strings.Index(line, "USER="); idx >= 0 {
			rest := line[idx+5:]
			if sp := strings.IndexByte(rest, ' '); sp >= 0 {
				user = rest[:sp]
			}
		}
		if idx := strings.Index(line, "COMMAND="); idx >= 0 {
			cmd = line[idx+8:]
		}
		if user != "" && cmd != "" {
			all = append(all, sudoEntry{user, cmd})
		}
	}

	if len(all) > n {
		all = all[len(all)-n:]
	}
	return all
}
