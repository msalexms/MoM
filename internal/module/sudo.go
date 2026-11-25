package module

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/msalexms/MoM/internal/module/render"
)

// SudoModule displays recent sudo activity from auth log.
type SudoModule struct{}

func (m *SudoModule) Name() string           { return "sudo" }
func (m *SudoModule) Title() string          { return "Sudo activity" }
func (m *SudoModule) Description() string    { return "Recent sudo commands from the auth log" }
func (m *SudoModule) Dependencies() []string { return nil }
func (m *SudoModule) Available() bool {
	_, err := os.Stat("/var/log/auth.log")
	return err == nil
}

type sudoEntry struct {
	user string
	cmd  string
}

func (m *SudoModule) Generate(ctx context.Context, opts render.Options) (string, error) {
	entries := getRecentSudo(5)
	if len(entries) == 0 {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()

	var lines []string
	for _, e := range entries {
		lines = append(lines, fmt.Sprintf("%s  %s", r.Col(e.user, 10, th.Palette.Warning), th.Dim(truncate(e.cmd, 34))))
	}

	compact := fmt.Sprintf("%d recent commands", len(entries))
	return r.Section("Sudo activity", "sudo", compact, lines), nil
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

// NeedsRoot implements Privileged.
func (m *SudoModule) NeedsRoot() bool { return true }
