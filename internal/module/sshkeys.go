package module

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/msalexms/MoM/internal/i18n"

	"github.com/msalexms/MoM/internal/module/render"
)

// SSHKeysModule displays loaded SSH keys and agent status.
type SSHKeysModule struct{}

func (m *SSHKeysModule) Name() string           { return "sshkeys" }
func (m *SSHKeysModule) Title() string          { return "SSH keys" }
func (m *SSHKeysModule) Description() string    { return "SSH agent status and authorized keys" }
func (m *SSHKeysModule) Dependencies() []string { return nil }
func (m *SSHKeysModule) Available() bool        { return true }

func (m *SSHKeysModule) Generate(ctx context.Context, opts render.Options) (string, error) {
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
		r.Field("agent", th.Color(agentStatus, agentColor)),
		r.Field("authorized", strconv.Itoa(authorizedKeys)),
	}

	compact := fmt.Sprintf(i18n.T("agent: %s  keys: %d"), th.Color(agentStatus, agentColor), authorizedKeys)
	return r.Section("SSH keys", "sshkeys", compact, lines), nil
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
