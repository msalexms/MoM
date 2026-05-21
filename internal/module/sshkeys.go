package module

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/msalexms/MoM/internal/module/render"
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

	agentStatus := "inactive"
	agentColor := th.Palette.Danger
	if agentRunning {
		agentStatus = "active"
		agentColor = th.Palette.Success
	}

	lines := []string{
		fmt.Sprintf("%-12s  %s", th.Color("agent", th.Palette.Warning), th.Color(agentStatus, agentColor)),
		fmt.Sprintf("%-12s  %d", th.Color("auth keys", th.Palette.Warning), authorizedKeys),
	}

	compact := fmt.Sprintf("agent: %s  keys: %d", th.Color(agentStatus, agentColor), authorizedKeys)
	return r.Section("SSH Keys", "sshkeys", compact, lines), nil
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
