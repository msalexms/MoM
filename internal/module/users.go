package module

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ams/mom/internal/module/render"
)

// UsersModule displays currently logged-in users and their sessions.
type UsersModule struct{}

func (m *UsersModule) Name() string           { return "users" }
func (m *UsersModule) Title() string          { return "Active Users" }
func (m *UsersModule) Description() string    { return "Currently logged-in users with TTY and source" }
func (m *UsersModule) Dependencies() []string { return nil }
func (m *UsersModule) Available() bool        { return true }
func (m *UsersModule) DefaultEnabled() bool   { return false }

func (m *UsersModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *UsersModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *UsersModule) Settings() []SettingDef         { return nil }

func (m *UsersModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

type userSession struct {
	user string
	tty  string
	from string
	idle string
}

func (m *UsersModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	sessions := getUtmpSessions()
	if len(sessions) == 0 {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()
	var sb strings.Builder

	switch r.Variant() {
	case render.VariantCompact:
		sb.WriteString(r.Header("Users", "users"))
		sb.WriteString(fmt.Sprintf("\n    %d active: ", len(sessions)))
		var names []string
		seen := make(map[string]bool)
		for _, s := range sessions {
			if !seen[s.user] {
				names = append(names, s.user)
				seen[s.user] = true
			}
		}
		sb.WriteString(strings.Join(names, ", "))

	case render.VariantBoxed:
		var content strings.Builder
		for _, s := range sessions {
			from := s.from
			if from == "" {
				from = "local"
			}
			content.WriteString(fmt.Sprintf("%-10s %-8s %s\n", s.user, s.tty, th.Dim(from)))
		}
		sb.WriteString(render.Indent(r.Box(strings.TrimRight(content.String(), "\n"), "Active Users"), "  "))

	case render.VariantPowerline:
		sb.WriteString(r.Header("Active Users", "users"))
		sb.WriteString("\n\n")
		for _, s := range sessions {
			from := s.from
			if from == "" {
				from = "local"
			}
			sb.WriteString(fmt.Sprintf("    %s %-10s %s %s\n",
				th.Color("▌", th.Palette.Accent), s.user,
				th.Color(s.tty, th.Palette.Warning),
				th.Dim(from)))
		}

	case render.VariantCards:
		var content strings.Builder
		for _, s := range sessions {
			from := s.from
			if from == "" {
				from = "local"
			}
			content.WriteString(fmt.Sprintf("  %-10s %-8s %s\n", s.user, s.tty, th.Dim(from)))
		}
		sb.WriteString(render.Indent(r.Card(strings.TrimRight(content.String(), "\n"), "Active Users"), "  "))

	default:
		sb.WriteString(r.Header("Active Users", "users"))
		sb.WriteString("\n\n")
		for _, s := range sessions {
			from := s.from
			if from == "" {
				from = "local"
			}
			sb.WriteString(fmt.Sprintf("    %-10s %-8s %s\n", s.user, th.Color(s.tty, th.Palette.Warning), th.Dim(from)))
		}
	}

	return sb.String(), nil
}

func getUtmpSessions() []userSession {
	// Parse /var/run/utmp is complex; use /proc approach instead
	// Read /proc/*/loginuid and match with /proc/*/status
	// Simpler: parse /run/utmp via reading who-style from /var/run/utmp
	// Fallback: read /proc for pts sessions
	entries, err := os.ReadDir("/dev/pts")
	if err != nil {
		return nil
	}

	var sessions []userSession
	seen := make(map[string]bool)

	// Scan /proc for processes with a controlling terminal
	procEntries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}

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
		if uid > 65000 {
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
				tty = link[strings.LastIndex(link, "/")+1:]
				tty = "pts/" + tty
			}
		}

		sessions = append(sessions, userSession{user: user, tty: tty})
		if len(sessions) >= 8 {
			break
		}
	}

	_ = entries
	if len(sessions) == 0 {
		// Minimal fallback
		if user := os.Getenv("USER"); user != "" {
			sessions = append(sessions, userSession{user: user, tty: "current"})
		}
	}
	return sessions
}

func lookupUser(uid int) string {
	data, err := os.ReadFile("/etc/passwd")
	if err != nil {
		return fmt.Sprintf("%d", uid)
	}
	prefix := fmt.Sprintf(":%d:", uid)
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, prefix) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 0 {
				return parts[0]
			}
		}
	}
	return fmt.Sprintf("%d", uid)
}
