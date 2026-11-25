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

// UsersModule displays currently logged-in users and their sessions.
type UsersModule struct{}

func (m *UsersModule) Name() string           { return "users" }
func (m *UsersModule) Title() string          { return "Active users" }
func (m *UsersModule) Description() string    { return "Logged-in users with their TTY" }
func (m *UsersModule) Dependencies() []string { return nil }
func (m *UsersModule) Available() bool        { return true }

type userSession struct {
	user string
	tty  string
}

func (m *UsersModule) Generate(ctx context.Context, opts render.Options) (string, error) {
	sessions := getUtmpSessions()
	if len(sessions) == 0 {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()

	var lines []string
	for _, s := range sessions {
		lines = append(lines, fmt.Sprintf("%-10s %s %s", s.user, r.Col(s.tty, 8, th.Palette.Warning), th.Dim("local")))
	}

	var names []string
	seen := make(map[string]bool)
	for _, s := range sessions {
		if !seen[s.user] {
			names = append(names, s.user)
			seen[s.user] = true
		}
	}
	compact := fmt.Sprintf("%d active: %s", len(sessions), strings.Join(names, ", "))

	return r.Section("Logged-in users", "users", compact, lines), nil
}

// getUtmpSessions scans /proc for shells with a controlling terminal —
// a pragmatic alternative to parsing utmp directly.
func getUtmpSessions() []userSession {
	procEntries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}

	var sessions []userSession
	seen := make(map[string]bool)

	for _, e := range procEntries {
		if !e.IsDir() || len(e.Name()) == 0 || e.Name()[0] < '0' || e.Name()[0] > '9' {
			continue
		}
		stat, err := os.ReadFile("/proc/" + e.Name() + "/stat")
		if err != nil {
			continue
		}
		fields := strings.Fields(string(stat))
		if len(fields) < 7 {
			continue
		}
		// field[1] = (comm), field[6] = tty_nr
		comm := strings.Trim(fields[1], "()")
		if comm != "bash" && comm != "zsh" && comm != "fish" && comm != "sh" && comm != "login" && comm != "sshd" {
			continue
		}

		status, err := os.ReadFile("/proc/" + e.Name() + "/status")
		if err != nil {
			continue
		}
		var uid int
		for _, line := range strings.Split(string(status), "\n") {
			if strings.HasPrefix(line, "Uid:") {
				fmt.Sscanf(strings.TrimPrefix(line, "Uid:"), "%d", &uid)
				break
			}
		}

		// Only real users (uid >= 1000) or root
		if uid != 0 && uid < 1000 {
			continue
		}

		user := lookupUser(uid)
		key := user + comm
		if seen[key] {
			continue
		}
		seen[key] = true

		tty := "pts/?"
		// Try to get the fd link
		fdPath := "/proc/" + e.Name() + "/fd/0"
		if link, err := os.Readlink(fdPath); err == nil {
			if strings.Contains(link, "/pts/") {
				tty = "pts/" + link[strings.LastIndex(link, "/")+1:]
			}
		}

		sessions = append(sessions, userSession{user: user, tty: tty})
		if len(sessions) >= 8 {
			break
		}
	}

	if len(sessions) == 0 {
		// Minimal fallback
		if user := os.Getenv("USER"); user != "" {
			sessions = append(sessions, userSession{user: user, tty: i18n.T("current")})
		}
	}
	return sessions
}

// lookupUser maps a uid to its account name via /etc/passwd.
func lookupUser(uid int) string {
	data, err := os.ReadFile("/etc/passwd")
	if err != nil {
		return strconv.Itoa(uid)
	}
	want := strconv.Itoa(uid)
	for _, line := range strings.Split(string(data), "\n") {
		// name:passwd:uid:gid:gecos:home:shell
		parts := strings.Split(line, ":")
		if len(parts) >= 3 && parts[2] == want {
			return parts[0]
		}
	}
	return strconv.Itoa(uid)
}
