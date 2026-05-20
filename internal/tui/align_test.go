package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestAlignment_Modules(t *testing.T) {
	rows := []struct {
		active  bool
		enabled bool
		name    string
		desc    string
	}{
		{true, true, "System", "Hostname, kernel, uptime"},
		{false, true, "Resources", "CPU load, RAM, disk usage"},
		{false, false, "Weather", "Current weather via wttr.in"},
		{false, true, "Distro Logo", "ASCII art logo"},
		{false, false, "Listening Ports", "TCP ports in LISTEN state"},
	}

	for _, r := range rows {
		cursor := listCursor(r.active, colCyan)
		checkbox := fixedCol("[ ]", 3, colGray)
		if r.enabled {
			checkbox = fixedCol("[✓]", 3, colGreen+colBold)
		}
		nameColor := colWhite
		if r.active {
			nameColor = colCyan
		}
		name := fixedCol(r.name, 20, nameColor)
		desc := dimText(r.desc)
		line := cursor + checkbox + " " + name + " " + desc

		w := lipgloss.Width(cursor + checkbox + " " + name + " ")
		fmt.Printf("w=%2d  %s\n", w, line)
		if w != 29 {
			t.Errorf("row %q: desc starts at col %d, want 29", r.name, w)
		}
	}
}

func TestAlignment_Dashboard(t *testing.T) {
	items := []struct{ icon, label string }{
		{" + ", "Select Modules"},
		{" ↕ ", "Reorder Modules"},
		{" # ", "Apply Template"},
		{" T ", "Theme & Style"},
		{" S ", "Services Picker"},
		{" A ", "ASCII Art Text"},
		{" > ", "Preview MOTD"},
	}

	for _, item := range items {
		cursor := listCursor(false, colCyan)
		icon := fixedCol("["+item.icon+"]", 5, colGray)
		label := col(item.label, colWhite)
		line := cursor + icon + " " + label

		w := lipgloss.Width(cursor + icon + " ")
		fmt.Printf("w=%2d  %s\n", w, line)
		if w != 10 {
			t.Errorf("item %q: label starts at col %d, want 10", item.label, w)
		}
	}
}

func TestAlignment_Order(t *testing.T) {
	rows := []struct {
		num     int
		enabled bool
		title   string
	}{
		{1, true, "System"},
		{2, true, "Resources"},
		{3, false, "Weather"},
		{10, true, "Distro Logo"},
		{15, false, "Listening Ports"},
	}

	for _, r := range rows {
		cursor := listCursor(false, colCyan)
		num := fixedCol(fmt.Sprintf("%2d.", r.num), 3, colGray)

		indicatorColor := colGray
		if r.enabled {
			indicatorColor = colGreen
		}
		indicator := fixedCol("●", 1, indicatorColor)
		if !r.enabled {
			indicator = fixedCol("○", 1, colGray)
		}

		name := fixedCol(r.title, 20, colWhite)
		line := cursor + num + " " + indicator + " " + name

		w := lipgloss.Width(cursor + num + " " + indicator + " ")
		fmt.Printf("w=%2d  %s\n", w, line)
		if w != 10 {
			t.Errorf("row %d (%q): name starts at col %d, want 10", r.num, r.title, w)
		}
	}
}

func TestAlignment_Profiles(t *testing.T) {
	items := []struct {
		save bool
		name string
	}{
		{true, "Save current config as new profile"},
		{false, "production"},
		{false, "staging"},
		{false, "dev-box"},
	}

	for _, item := range items {
		cursor := listCursor(false, colYellow)
		prefix := fixedCol("  ", 2, colGray)
		if item.save {
			prefix = fixedCol("+ ", 2, colYellow)
		}
		name := col(item.name, colWhite)
		line := cursor + prefix + name

		w := lipgloss.Width(cursor + prefix)
		fmt.Printf("w=%2d  %s\n", w, line)
		if w != 6 {
			t.Errorf("item %q: name starts at col %d, want 6", item.name, w)
		}
	}
}

func TestFixedCol_UnicodeWidth(t *testing.T) {
	tests := []struct {
		text  string
		width int
	}{
		{"hello", 10},
		{"✓", 3},
		{"[✓]", 5},
		{"↕", 5},
		{"[ ↕ ]", 7},
		{"日本語", 10},
		{"café", 8},
	}

	for _, tc := range tests {
		result := fixedCol(tc.text, tc.width, "")
		got := lipgloss.Width(result)
		if got != tc.width {
			t.Errorf("fixedCol(%q, %d) = width %d, want %d", tc.text, tc.width, got, tc.width)
		}
	}
}

func TestNormalizeLineWidths(t *testing.T) {
	input := "short\na much longer line here\ntiny"
	result := normalizeLineWidths(input)
	lines := strings.Split(result, "\n")

	// All lines should have the same visible width
	maxW := lipgloss.Width(lines[0])
	for i, line := range lines {
		w := lipgloss.Width(line)
		if w != maxW {
			t.Errorf("line %d has width %d, want %d", i, w, maxW)
		}
	}
}

func TestNormalizeLineWidths_Empty(t *testing.T) {
	result := normalizeLineWidths("")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestTruncateToWidth(t *testing.T) {
	tests := []struct {
		text     string
		maxWidth int
		want     string
	}{
		{"hello world", 5, "hello"},
		{"日本語テスト", 6, "日本語"},
		{"café", 4, "café"},
		{"[✓] enabled", 3, "[✓]"},
		{"", 5, ""},
		{"a", 0, ""},
	}

	for _, tc := range tests {
		got := truncateToWidth(tc.text, tc.maxWidth)
		if got != tc.want {
			t.Errorf("truncateToWidth(%q, %d) = %q, want %q", tc.text, tc.maxWidth, got, tc.want)
		}
	}
}
