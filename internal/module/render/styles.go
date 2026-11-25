package render

import (
	"strings"

	"github.com/mattn/go-runewidth"

	"github.com/msalexms/MoM/internal/theme"
)

// --- Unicode Icons ---
// These use standard Unicode symbols that render correctly in any modern
// terminal without requiring Nerd Fonts. They degrade to ASCII labels when
// the theme has UseUnicode=false.

// UnicodeIcons maps semantic names to standard Unicode glyphs.
var UnicodeIcons = map[string]string{
	"cpu":       "◈",
	"ram":       "◇",
	"disk":      "◉",
	"temp":      "♨",
	"net":       "⇄",
	"globe":     "◎",
	"weather":   "☁",
	"docker":    "▣",
	"service":   "◆",
	"update":    "↑",
	"user":      "●",
	"login":     "→",
	"clock":     "◷",
	"calendar":  "▦",
	"quote":     "❝",
	"terminal":  "▶",
	"host":      "⌂",
	"kernel":    "⚙",
	"shell":     "$",
	"uptime":    "△",
	"check":     "✓",
	"cross":     "✗",
	"warning":   "⚠",
	"arrow":     "▸",
	"dot":       "●",
	"star":      "★",
	"separator": "│",
	"linux":     "⚙",
	"logo":      "◈",
}

// Icon returns the Unicode icon for the given key, or the ASCII fallback
// if the theme doesn't support Unicode.
func (r *Renderer) Icon(name string) string {
	if r.Opts.Theme.UseUnicode {
		if icon, ok := UnicodeIcons[name]; ok {
			return icon
		}
	}
	// ASCII fallback
	switch name {
	case "cpu":
		return "cpu"
	case "ram":
		return "mem"
	case "disk":
		return "dsk"
	case "net", "globe":
		return "net"
	case "check":
		return "+"
	case "cross":
		return "x"
	case "warning":
		return "!"
	case "dot":
		return "*"
	case "separator":
		return "|"
	default:
		return ""
	}
}

// --- Box drawing (for VariantBoxed) ---

// Box wraps content in a Unicode box with rounded corners.
// Width is auto-detected from the longest line if 0.
func (r *Renderer) Box(content string, title string) string {
	th := r.Opts.Theme
	if !th.UseUnicode {
		return r.boxASCII(content, title)
	}

	lines := strings.Split(content, "\n")
	// Use a fixed inner width for consistent alignment
	innerW := 44
	for _, l := range lines {
		w := visibleLen(l)
		if w > innerW {
			innerW = w
		}
	}

	color := th.Palette.Subtle
	titleColor := th.Palette.Accent

	var sb strings.Builder
	// Top border
	if title != "" {
		titleRendered := th.Color(" "+title+" ", titleColor)
		// Total visible: ╭─ + " title " + ─*remaining + ╮ = 2 + (visibleLen(title)+2) + remaining + 1
		// Must equal innerW + 4 (same as content lines: │ + space + innerW + space + │)
		remaining := innerW + 4 - 2 - visibleLen(title) - 2 - 1
		if remaining < 0 {
			remaining = 0
		}
		sb.WriteString(th.Color("╭─", color) + titleRendered + th.Color(strings.Repeat("─", remaining)+"╮", color) + "\n")
	} else {
		sb.WriteString(th.Color("╭"+strings.Repeat("─", innerW+2)+"╮", color) + "\n")
	}

	// Content lines — pad each to innerW
	for _, l := range lines {
		pad := innerW - visibleLen(l)
		if pad < 0 {
			pad = 0
		}
		sb.WriteString(th.Color("│", color) + " " + l + strings.Repeat(" ", pad) + " " + th.Color("│", color) + "\n")
	}

	// Bottom border
	sb.WriteString(th.Color("╰"+strings.Repeat("─", innerW+2)+"╯", color))

	return sb.String()
}

func (r *Renderer) boxASCII(content string, title string) string {
	lines := strings.Split(content, "\n")
	maxW := 0
	for _, l := range lines {
		w := visibleLen(l)
		if w > maxW {
			maxW = w
		}
	}
	if maxW < visibleLen(title)+4 {
		maxW = visibleLen(title) + 4
	}

	var sb strings.Builder
	if title != "" {
		sb.WriteString("+-" + title + strings.Repeat("-", maxW-visibleLen(title)+1) + "+\n")
	} else {
		sb.WriteString("+" + strings.Repeat("-", maxW+2) + "+\n")
	}
	for _, l := range lines {
		pad := maxW - visibleLen(l)
		sb.WriteString("| " + l + strings.Repeat(" ", pad) + " |\n")
	}
	sb.WriteString("+" + strings.Repeat("-", maxW+2) + "+")
	return sb.String()
}

// --- Sparkline (mini inline chart) ---

// Sparkline renders a tiny inline bar chart from values (0-100 each).
// Useful for showing load history or trends in compact mode.
func (r *Renderer) Sparkline(values []float64) string {
	if !r.Opts.Theme.UseUnicode {
		return r.sparklineASCII(values)
	}
	blocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	var sb strings.Builder
	th := r.Opts.Theme
	for _, v := range values {
		if v < 0 {
			v = 0
		}
		if v > 100 {
			v = 100
		}
		idx := int(v / 100.0 * 7)
		if idx > 7 {
			idx = 7
		}
		color := th.GradientAt(v/100.0, v)
		if color != "" {
			sb.WriteString(color)
		}
		sb.WriteRune(blocks[idx])
	}
	sb.WriteString(theme.Reset)
	return sb.String()
}

func (r *Renderer) sparklineASCII(values []float64) string {
	chars := []byte{' ', '.', '-', '=', '#'}
	var sb strings.Builder
	for _, v := range values {
		idx := int(v / 100.0 * 4)
		if idx > 4 {
			idx = 4
		}
		sb.WriteByte(chars[idx])
	}
	return sb.String()
}

// --- Utility ---

// Indent prepends prefix to every line of s.
func Indent(s string, prefix string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}

// --- Cards helpers ---

// Card wraps content in a double-bordered card with optional shadow effect.
// Uses ╔═╗║╚═╝ characters for a heavier, more prominent look.
func (r *Renderer) Card(content string, title string) string {
	th := r.Opts.Theme
	if !th.UseUnicode {
		return r.boxASCII(content, title)
	}

	lines := strings.Split(content, "\n")
	innerW := 42
	for _, l := range lines {
		w := visibleLen(l)
		if w > innerW {
			innerW = w
		}
	}

	color := th.Palette.Subtle
	titleColor := th.Palette.Accent
	shadow := th.Palette.Muted

	var sb strings.Builder

	// Top border with centered title
	if title != "" {
		remaining := innerW + 4 - 2 - visibleLen(title) - 2 - 1
		if remaining < 0 {
			remaining = 0
		}
		sb.WriteString(th.Color("╔═", color) + th.Color(" "+title+" ", titleColor) + th.Color(strings.Repeat("═", remaining)+"╗", color) + "\n")
	} else {
		sb.WriteString(th.Color("╔"+strings.Repeat("═", innerW+2)+"╗", color) + "\n")
	}

	// Content lines with side borders
	for _, l := range lines {
		pad := innerW - visibleLen(l)
		if pad < 0 {
			pad = 0
		}
		sb.WriteString(th.Color("║", color) + " " + l + strings.Repeat(" ", pad) + " " + th.Color("║", color) + th.Color("░", shadow) + "\n")
	}

	// Bottom border with shadow
	sb.WriteString(th.Color("╚"+strings.Repeat("═", innerW+2)+"╝", color) + th.Color("░", shadow) + "\n")
	sb.WriteString(" " + th.Color(strings.Repeat("░", innerW+4), shadow))

	return sb.String()
}

// visibleLen returns the visible width of a string, stripping ANSI escapes
// and accounting for wide Unicode characters.
func visibleLen(s string) int {
	// Strip ANSI escape sequences first
	var clean strings.Builder
	inEsc := false
	for _, r := range s {
		if r == '\033' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		clean.WriteRune(r)
	}
	return runewidth.StringWidth(clean.String())
}

// PadRight pads a string (which may contain ANSI escapes) to a fixed visible
// width using spaces. If the string is already wider, it is returned unchanged.
func PadRight(s string, width int) string {
	visible := visibleLen(s)
	if visible >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visible)
}
