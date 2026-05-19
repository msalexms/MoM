package tui

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestAlignment_Modules(t *testing.T) {
	// Simulate what viewModules produces for each row
	rows := []struct{ active bool; enabled bool; name string; desc string }{
		{true, true, "System", "Hostname, kernel, uptime"},
		{false, true, "Resources", "CPU load, RAM, disk usage"},
		{false, false, "Weather", "Current weather via wttr.in"},
		{false, true, "Distro Logo", "ASCII art logo"},
		{false, false, "Listening Ports", "TCP ports in LISTEN state"},
	}

	for _, r := range rows {
		cursor := listCursor(r.active, colCyan)
		checkbox := colGray + "[ ]" + colReset
		if r.enabled {
			checkbox = colGreen + colBold + "[✓]" + colReset
		}
		nameColor := colWhite
		if r.active { nameColor = colCyan }
		name := fixedCol(r.name, 20, nameColor)
		desc := dimText(r.desc)
		line := cursor + checkbox + " " + name + " " + desc

		// Measure: the description should start at the same column
		// Strip ANSI and check that char at position 30 is consistent
		w := lipgloss.Width(cursor + checkbox + " " + name + " ")
		fmt.Printf("w=%2d  %s\n", w, line)
		if w != 29 {
			t.Errorf("row %q: desc starts at col %d, want 29", r.name, w)
		}
	}
}
