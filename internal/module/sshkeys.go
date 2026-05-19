package module

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ams/mom/internal/module/render"
)

// SSHKeysModule displays loaded SSH keys and agent status.
type SSHKeysModule struct{}

func (m *SSHKeysModule) Name() string           { return "sshkeys" }
func (m *SSHKeysModule) Title() string          { return "SSH Keys" }
func (m *SSHKeysModule) Description() string    { return "SSH agent status and authorized keys count" }
func (m *SSHKeysModule) Dependencies() []string { return nil }
func (m *SSHKeysModule) Available() bool        { return true }
func (m *SSHKeysModule) DefaultEnabled() bool   { return false }

func (m *SSHKeysModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *SSHKeysModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *SSHKeysModule) Settings() []SettingDef         { return nil }

func (m *SSHKeysModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

func (m *SSHKeysModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	agentRunning := os.Getenv("SSH_AUTH_SOCK") != ""
	authorizedKeys := countAuthorizedKeys()

	r := render.New(opts)
	th := r.Theme()
	var sb strings.Builder

	agentStatus := "inactive"
	agentColor := th.Palette.Danger
	if agentRunning {
		agentStatus = "active"
		agentColor = th.Palette.Success
	}

	switch r.Variant() {
	case render.VariantCompact:
		sb.WriteString(r.Header("SSH", "sshkeys"))
		sb.WriteString(fmt.Sprintf("\n    agent: %s  keys: %d",
			th.Color(agentStatus, agentColor), authorizedKeys))

	case render.VariantBoxed:
		var content strings.Builder
		content.WriteString(fmt.Sprintf("%-12s  %s\n", "agent", th.Color(agentStatus, agentColor)))
		content.WriteString(fmt.Sprintf("%-12s  %d", "auth keys", authorizedKeys))
		sb.WriteString(render.Indent(r.Box(content.String(), "SSH Keys"), "  "))

	case render.VariantPowerline:
		sb.WriteString(r.Header("SSH Keys", "sshkeys"))
		sb.WriteString("\n\n")
		sb.WriteString(fmt.Sprintf("    %s %-10s %s\n",
			th.Color("▌", agentColor), th.Color("agent", th.Palette.Warning),
			th.Color(agentStatus, agentColor)))
		sb.WriteString(fmt.Sprintf("    %s %-10s %d",
			th.Color("▌", th.Palette.Accent), th.Color("auth keys", th.Palette.Warning),
			authorizedKeys))

	case render.VariantCards:
		var content strings.Builder
		content.WriteString(fmt.Sprintf("  %-12s  %s\n", "agent", th.Color(agentStatus, agentColor)))
		content.WriteString(fmt.Sprintf("  %-12s  %d", "auth keys", authorizedKeys))
		sb.WriteString(render.Indent(r.Card(content.String(), "SSH Keys"), "  "))

	default:
		sb.WriteString(r.Header("SSH Keys", "sshkeys"))
		sb.WriteString("\n\n")
		sb.WriteString(r.KeyValue("agent", th.Color(agentStatus, agentColor)) + "\n")
		sb.WriteString(r.KeyValue("auth keys", fmt.Sprintf("%d", authorizedKeys)))
	}

	return sb.String(), nil
}

func countAuthorizedKeys() int {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0
	}
	data, err := os.ReadFile(home + "/.ssh/authorized_keys")
	if err != nil {
		return 0
	}
	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			count++
		}
	}
	return count
}
