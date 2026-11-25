package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestPadKeepsColumnsAligned(t *testing.T) {
	for _, s := range []string{"Sistema", "Logo de la distro", Accent.Render("Clima"), "Ñandú"} {
		if w := lipgloss.Width(Pad(s, 12)); w < 12 {
			t.Errorf("Pad(%q) width %d, want >= 12", s, w)
		}
	}
	if w := lipgloss.Width(Pad("un texto demasiado largo", 8)); w != 8 {
		t.Errorf("plain text should be truncated to 8 cells, got %d", w)
	}
}

func TestCursorWidthIsStable(t *testing.T) {
	if a, b := lipgloss.Width(Cursor(true)), lipgloss.Width(Cursor(false)); a != b {
		t.Errorf("cursor widths differ: %d vs %d", a, b)
	}
}

func TestWindowKeepsCursorVisible(t *testing.T) {
	for _, tc := range []struct{ total, cursor, height int }{
		{5, 2, 10}, {32, 0, 10}, {32, 15, 10}, {32, 31, 10}, {32, 31, 1},
	} {
		start, end := Window(tc.total, tc.cursor, tc.height)
		if tc.cursor < start || tc.cursor >= end {
			t.Errorf("%+v: cursor outside [%d,%d)", tc, start, end)
		}
		if end-start > max(tc.height, 0) && tc.total > tc.height {
			t.Errorf("%+v: window larger than height", tc)
		}
	}
}

func TestThorLogoShape(t *testing.T) {
	rows := strings.Split(ThorLogo(), "\n")
	if len(rows) != 6 {
		t.Fatalf("want 6 rows, got %d", len(rows))
	}
	for i, r := range rows {
		if w := lipgloss.Width(r); w != 12 {
			t.Errorf("row %d width %d, want 12", i, w)
		}
	}
}
