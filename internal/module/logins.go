package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ams/mom/internal/module/render"
)

// LoginsModule displays recent logins and active SSH sessions.
type LoginsModule struct{}

func (m *LoginsModule) Name() string           { return "logins" }
func (m *LoginsModule) Title() string          { return "Logins" }
func (m *LoginsModule) Description() string    { return "Last logins and active SSH sessions" }
func (m *LoginsModule) Dependencies() []string { return []string{"last", "who"} }
func (m *LoginsModule) Available() bool        { return CheckDependency("last") }
func (m *LoginsModule) DefaultEnabled() bool   { return false }

func (m *LoginsModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact}
}
func (m *LoginsModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *LoginsModule) Settings() []SettingDef         { return nil }

func (m *LoginsModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

func (m *LoginsModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	r := render.New(opts)
	var sb strings.Builder

	sb.WriteString(r.Header("Logins", "logins"))
	sb.WriteString("\n\n")

	if CheckDependency("who") {
		cmd := exec.CommandContext(ctx, "who")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			count := 0
			if len(lines) > 0 && lines[0] != "" {
				count = len(lines)
			}
			sb.WriteString(r.KeyValue("active", fmt.Sprintf("%d session(s)", count)) + "\n\n")
		}
	}

	cmd := exec.CommandContext(ctx, "last", "-5", "-w")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		sb.WriteString(r.KeyValue("recent", "") + "\n")
		count := 0
		th := r.Theme()
		for _, line := range lines {
			if line == "" || strings.HasPrefix(line, "wtmp") || strings.HasPrefix(line, "reboot") {
				continue
			}
			if count >= 4 {
				break
			}
			entry := truncate(strings.TrimSpace(line), 44)
			sb.WriteString("      " + th.Dim(entry) + "\n")
			count++
		}
	}

	return sb.String(), nil
}
