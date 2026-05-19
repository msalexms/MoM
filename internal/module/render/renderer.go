package render

import (
	"fmt"
	"strings"

	"github.com/ams/mom/internal/theme"
)

// Variant identifies a presentation style of a module's output. Modules
// declare which variants they support; if asked for an unknown variant they
// must fall back to VariantDefault.
type Variant string

const (
	VariantDefault  Variant = "default"
	VariantCompact  Variant = "compact"
	VariantDetailed Variant = "detailed"
	VariantMinimal  Variant = "minimal"
	VariantASCII    Variant = "ascii"
	VariantBoxed    Variant = "boxed"
)

// Options bundles every rendering choice that can vary at MOTD-generation
// time: which palette to use, which variant of the module to emit, and how
// wide the terminal is (0 = unknown, fall back to module defaults).
//
// Modules should always tolerate a zero-value Options by calling Resolve().
type Options struct {
	Theme   *theme.Theme
	Variant Variant
	Width   int
}

// DefaultOptions returns an Options with the default theme and variant,
// suitable for legacy callers that don't yet pass an Options around.
func DefaultOptions() Options {
	return Options{Theme: theme.Default(), Variant: VariantDefault, Width: 0}
}

// Resolve fills in missing fields with sensible defaults so callers can
// safely use the returned Options without nil checks.
func (o Options) Resolve() Options {
	if o.Theme == nil {
		o.Theme = theme.Default()
	}
	if o.Variant == "" {
		o.Variant = VariantDefault
	}
	return o
}

// Renderer ties an Options to a string-building API. It is the preferred
// entry point for new module code: instead of pasting ANSI codes directly,
// modules call r.Header / r.KeyValue / r.ProgressBar etc.
//
// Renderer methods never write to stdout; they return strings. This keeps
// modules pure and easy to test.
type Renderer struct {
	Opts Options
}

// New returns a Renderer bound to the given options. Pass DefaultOptions()
// for the default theme.
func New(opts Options) *Renderer {
	return &Renderer{Opts: opts.Resolve()}
}

// Theme returns the renderer's active theme.
func (r *Renderer) Theme() *theme.Theme { return r.Opts.Theme }

// Variant returns the active variant.
func (r *Renderer) Variant() Variant { return r.Opts.Variant }

// --- Section header ---

// Header renders a module section header. The color is taken from the
// theme's section palette using the module name as key.
func (r *Renderer) Header(title, moduleName string) string {
	th := r.Opts.Theme
	color := th.SectionColor(moduleName)
	return r.headerWithColor(title, color)
}

// HeaderColor renders a header with an explicit color, bypassing the
// section palette. Useful for one-off headings (e.g. "Containers (docker)").
func (r *Renderer) HeaderColor(title, color string) string {
	return r.headerWithColor(title, color)
}

func (r *Renderer) headerWithColor(title, color string) string {
	th := r.Opts.Theme
	if r.Opts.Variant == VariantMinimal {
		return "  " + th.Bold(th.Color(title, color))
	}
	line := strings.Repeat("-", 36)
	return fmt.Sprintf("  %s%s%s%s %s%s%s",
		color, th.Attrs.Bold, title, theme.Reset,
		th.Attrs.Dim+color, line, theme.Reset)
}

// --- Key/value pair ---

// KeyValue renders an aligned "key: value" row with the default key width.
func (r *Renderer) KeyValue(key, value string) string {
	return r.KeyValueWidth(key, value, 10)
}

// KeyValueWidth renders an aligned "key: value" row with an explicit key
// column width. Pad is in characters of the displayable key.
func (r *Renderer) KeyValueWidth(key, value string, keyWidth int) string {
	th := r.Opts.Theme
	keyCol := th.Palette.Warning
	if keyCol == "" {
		keyCol = th.Palette.Accent
	}
	return fmt.Sprintf("    %s%-*s%s  %s",
		keyCol, keyWidth, key, theme.Reset,
		th.Color(value, th.Palette.Foreground))
}

// KeyValueColored is like KeyValueWidth but lets the caller pick the value
// color (e.g. green for ok, red for error).
func (r *Renderer) KeyValueColored(key, value, valueColor string, keyWidth int) string {
	th := r.Opts.Theme
	keyCol := th.Palette.Warning
	if keyCol == "" {
		keyCol = th.Palette.Accent
	}
	return fmt.Sprintf("    %s%-*s%s  %s",
		keyCol, keyWidth, key, theme.Reset,
		th.Color(value, valueColor))
}

// --- Status dot ---

// StatusDot renders a colored bullet for a service/process status string.
func (r *Renderer) StatusDot(status string) string {
	th := r.Opts.Theme
	color, _ := th.Status(status)
	dot := IconBullet
	if !th.UseUnicode {
		dot = "*"
	}
	return th.Color(dot, color)
}

// --- Progress bar ---

// ProgressBar renders a gradient progress bar with a percentage label.
// width is the bar width in cells. label is appended after the percentage.
// When the theme disables Unicode, blocks are substituted by '#' and '-'.
func (r *Renderer) ProgressBar(percent float64, width int, label string) string {
	th := r.Opts.Theme
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	if width <= 0 {
		width = 20
	}

	full := "█"
	empty := "░"
	if !th.UseUnicode {
		full = "#"
		empty = "-"
	}

	filled := int(percent / 100.0 * float64(width))
	emptyN := width - filled

	var b strings.Builder
	for i := 0; i < filled; i++ {
		col := th.GradientAt(float64(i)/float64(width), percent)
		if col != "" {
			b.WriteString(col)
		}
		b.WriteString(full)
	}
	if filled > 0 {
		b.WriteString(theme.Reset)
	}
	if th.Attrs.Dim != "" {
		b.WriteString(th.Attrs.Dim)
	}
	for i := 0; i < emptyN; i++ {
		b.WriteString(empty)
	}
	b.WriteString(theme.Reset)

	pct := fmt.Sprintf("%5.1f%%", percent)
	pct = th.Color(pct, th.PercentColor(percent))

	if label == "" {
		return b.String() + " " + pct
	}
	return b.String() + " " + pct + "  " + th.Dim(label)
}

// --- Banner ---

// AsciiBanner renders text as 5x5 ASCII block letters using the theme's
// accent color. When the theme disables Unicode, '█' is replaced by '#'.
func (r *Renderer) AsciiBanner(text string) string {
	th := r.Opts.Theme
	color := th.Palette.Accent
	if text == "" {
		return ""
	}
	text = strings.ToUpper(text)

	full := "█"
	if !th.UseUnicode {
		full = "#"
	}

	lines := make([]string, 5)
	for _, ch := range text {
		pattern := charPattern(ch)
		for i := 0; i < 5; i++ {
			row := pattern[i]
			if !th.UseUnicode {
				row = strings.ReplaceAll(row, "█", full)
			}
			if color != "" {
				lines[i] += color + row + theme.Reset
			} else {
				lines[i] += row
			}
			lines[i] += " "
		}
	}
	return strings.Join(lines, "\n")
}

// FormatBytes is re-exported as a method to avoid having to import both
// the package and the renderer in modules.
func (r *Renderer) FormatBytes(b uint64) string { return FormatBytes(b) }
