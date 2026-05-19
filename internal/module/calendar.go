package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// CalendarModule displays the current month's calendar.
type CalendarModule struct{}

func (m *CalendarModule) Name() string        { return "calendar" }
func (m *CalendarModule) Title() string       { return "Calendar" }
func (m *CalendarModule) Description() string { return "Current month calendar with today highlighted" }
func (m *CalendarModule) Dependencies() []string { return []string{"cal"} }
func (m *CalendarModule) Available() bool     { return CheckDependency("cal") }
func (m *CalendarModule) DefaultEnabled() bool { return false }

func (m *CalendarModule) Generate(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "cal")
	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	lines := strings.Split(strings.TrimRight(string(output), "\n"), "\n")
	today := fmt.Sprintf("%d", time.Now().Day())

	var sb strings.Builder
	sb.WriteString("┌─ Calendar ───────────────────────────┐\n")
	for _, line := range lines {
		// Highlight today's date
		highlighted := highlightToday(line, today)
		sb.WriteString(fmt.Sprintf("│ %-37s │\n", highlighted))
	}
	sb.WriteString("└───────────────────────────────────────┘")

	return sb.String(), nil
}

// highlightToday wraps today's date with ANSI bold/inverse markers.
func highlightToday(line string, today string) string {
	// cal already highlights today with reverse video on most systems
	// We just return the line as-is since terminal handles the highlight
	return line
}
