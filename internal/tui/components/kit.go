// Package components is the visual kit of the TUI: the AMSTech palette,
// the brand mark and the small width-aware helpers every view is built
// from, so all screens share one accent, one cursor and one rhythm.
//
// Colors follow the "noche" mode of the AMSTech design system (manual de
// marca v1.5, amstech.dev): indigo brand #7F89E8, Thor #9EA6F2, warm light
// text #E9E6E0 and the yellow/green/red signal colors, which are the ones
// that read well on the dark background most terminals use.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// AMSTech palette, noche mode.
const (
	ColorText   = "#E9E6E0" // texto
	ColorText2  = "#C4C0CB" // texto-2
	ColorMuted  = "#9E9AA8" // gris
	ColorLine   = "#4A4A57" // borde, lifted a step for pure-black terminals
	ColorBrand  = "#7F89E8" // marca
	ColorThor   = "#9EA6F2" // thor (logo)
	ColorInk    = "#18181D" // fondo noche, used inside the logo
	ColorYellow = "#D9C26A" // amarillo
	ColorOK     = "#8FD3A4" // ok
	ColorBad    = "#F0A094" // mal
)

var (
	Text   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorText))
	Text2  = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorText2))
	Muted  = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMuted))
	Line   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorLine))
	Brand  = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBrand))
	Strong = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorText)).Bold(true)
	Accent = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBrand)).Bold(true)
	Yellow = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorYellow))
	OK     = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOK))
	Bad    = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBad))

	// Key renders a keyboard shortcut in hints and help tables.
	Key = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorThor)).Bold(true)
	// Label is a small uppercase group heading ("CONFIGURAR").
	Label = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMuted)).Bold(true)
)

// Cursor renders the 2-cell list cursor.
func Cursor(active bool) string {
	if active {
		return Accent.Render("▸") + " "
	}
	return "  "
}

// Check renders a checkbox.
func Check(on bool) string {
	if on {
		return OK.Render("[✓]")
	}
	return Line.Render("[ ]")
}

// Dot renders an on/off indicator.
func Dot(on bool) string {
	if on {
		return OK.Render("●")
	}
	return Line.Render("○")
}

// Item renders a list label, highlighted when the cursor is on it and
// dimmed when it is disabled.
func Item(text string, active, enabled bool) string {
	switch {
	case active:
		return Accent.Render(text)
	case !enabled:
		return Muted.Render(text)
	default:
		return Text.Render(text)
	}
}

// Pad pads s (which may contain ANSI escapes) to width visible cells,
// truncating plain text that is too long.
func Pad(s string, width int) string {
	w := lipgloss.Width(s)
	if w > width && !strings.Contains(s, "\x1b") {
		s = runewidth.Truncate(s, width, "…")
		w = lipgloss.Width(s)
	}
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

// Truncate shortens plain text to at most width cells.
func Truncate(s string, width int) string {
	return runewidth.Truncate(s, width, "…")
}

// Rule renders a horizontal separator of the given width.
func Rule(width int) string {
	return Line.Render(strings.Repeat("─", max(0, width)))
}

// Hints renders "key action · key action" pairs for the footer.
func Hints(pairs ...string) string {
	var parts []string
	for i := 0; i+1 < len(pairs); i += 2 {
		parts = append(parts, Key.Render(pairs[i])+" "+Muted.Render(pairs[i+1]))
	}
	return strings.Join(parts, Line.Render("  ·  "))
}

// Window returns the [start, end) slice of a list of total items that
// keeps cursor visible within height rows.
func Window(total, cursor, height int) (start, end int) {
	if height <= 0 || total <= height {
		return 0, total
	}
	start = cursor - height/2
	start = max(0, min(start, total-height))
	return start, start + height
}

// ScrollInfo renders "12–30 de 32" under a windowed list, or "" when the
// whole list is visible.
func ScrollInfo(start, end, total int) string {
	if start == 0 && end == total {
		return ""
	}
	return Muted.Render(fmt.Sprintf("%d–%d de %d", start+1, end, total))
}

// BrandMark renders the "[AMS]" signature.
func BrandMark() string {
	return Muted.Render("[") + lipgloss.NewStyle().Foreground(lipgloss.Color(ColorThor)).Bold(true).Render("AMS") + Muted.Render("]")
}
