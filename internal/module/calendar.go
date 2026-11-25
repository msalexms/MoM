package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/msalexms/MoM/internal/module/render"
)

// CalendarModule displays the current month's calendar.
type CalendarModule struct{}

func (m *CalendarModule) Name() string  { return "calendar" }
func (m *CalendarModule) Title() string { return "Calendar" }
func (m *CalendarModule) Description() string {
	return "This month's calendar with today highlighted"
}
func (m *CalendarModule) Dependencies() []string { return []string{"cal"} }
func (m *CalendarModule) Available() bool        { return CheckDependency("cal") }

func (m *CalendarModule) Generate(ctx context.Context, opts render.Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "cal", "--color=always")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.CommandContext(ctx, "cal")
		output, err = cmd.Output()
		if err != nil {
			return "", nil
		}
	}

	lines := strings.Split(strings.TrimRight(string(output), "\n"), "\n")
	r := render.New(opts)

	var sb strings.Builder
	sb.WriteString(r.Header("Calendar", "calendar"))
	sb.WriteString("\n\n")

	for _, line := range lines {
		sb.WriteString(fmt.Sprintf("  %s\n", line))
	}

	return sb.String(), nil
}
