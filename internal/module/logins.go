package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/msalexms/MoM/internal/i18n"

	"github.com/msalexms/MoM/internal/module/render"
)

// LoginsModule displays recent logins and active SSH sessions.
type LoginsModule struct{}

func (m *LoginsModule) Name() string           { return "logins" }
func (m *LoginsModule) Title() string          { return "Logins" }
func (m *LoginsModule) Description() string    { return "Recent logins and active SSH sessions" }
func (m *LoginsModule) Dependencies() []string { return []string{"last"} }
func (m *LoginsModule) Available() bool        { return CheckDependency("last") }

func (m *LoginsModule) Generate(ctx context.Context, opts render.Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	r := render.New(opts)
	th := r.Theme()

	var lines []string
	sessions := 0

	if CheckDependency("who") {
		cmd := exec.CommandContext(ctx, "who")
		output, err := cmd.Output()
		if err == nil {
			rawLines := strings.Split(strings.TrimSpace(string(output)), "\n")
			if len(rawLines) > 0 && rawLines[0] != "" {
				sessions = len(rawLines)
			}
			lines = append(lines, r.Field("active", fmt.Sprintf(i18n.T("%d session(s)"), sessions)))
			lines = append(lines, "")
		}
	}

	cmd := exec.CommandContext(ctx, "last", "-5", "-w")
	output, err := cmd.Output()
	if err == nil {
		rawLines := strings.Split(strings.TrimSpace(string(output)), "\n")
		lines = append(lines, th.Color(i18n.T("recent"), th.Palette.Warning))
		count := 0
		for _, line := range rawLines {
			if line == "" || strings.HasPrefix(line, "wtmp") || strings.HasPrefix(line, "reboot") {
				continue
			}
			if count >= 4 {
				break
			}
			lines = append(lines, "  "+th.Dim(truncate(strings.TrimSpace(line), 44)))
			count++
		}
	}

	if len(lines) == 0 {
		return "", nil
	}

	compact := fmt.Sprintf(i18n.T("%d session(s)"), sessions)
	return r.Section("Sessions", "logins", compact, lines), nil
}
