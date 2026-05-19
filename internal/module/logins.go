package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// LoginsModule displays recent logins and active SSH sessions.
type LoginsModule struct{}

func (m *LoginsModule) Name() string        { return "logins" }
func (m *LoginsModule) Title() string       { return "Recent Logins" }
func (m *LoginsModule) Description() string { return "Last logins and active SSH sessions" }
func (m *LoginsModule) Dependencies() []string { return []string{"last", "who"} }
func (m *LoginsModule) Available() bool     { return CheckDependency("last") }
func (m *LoginsModule) DefaultEnabled() bool { return false }

func (m *LoginsModule) Generate(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var sb strings.Builder
	sb.WriteString("┌─ Logins ─────────────────────────────┐\n")

	// Active sessions
	if CheckDependency("who") {
		cmd := exec.CommandContext(ctx, "who")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			if len(lines) > 0 && lines[0] != "" {
				sb.WriteString(fmt.Sprintf("│ Active sessions: %-19d │\n", len(lines)))
			} else {
				sb.WriteString("│ Active sessions: 0                   │\n")
			}
		}
	}

	// Last logins
	cmd := exec.CommandContext(ctx, "last", "-5", "-w")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		sb.WriteString("│                                       │\n")
		sb.WriteString("│ Recent:                               │\n")
		count := 0
		for _, line := range lines {
			if line == "" || strings.HasPrefix(line, "wtmp") || strings.HasPrefix(line, "reboot") {
				continue
			}
			if count >= 3 {
				break
			}
			entry := truncate(strings.TrimSpace(line), 37)
			sb.WriteString(fmt.Sprintf("│ %-37s │\n", entry))
			count++
		}
	}

	sb.WriteString("└───────────────────────────────────────┘")
	return sb.String(), nil
}
