// Package theme provides ANSI color palettes and semantic colors used to render
// MOTD module output. Themes decouple modules from concrete color codes: a
// module asks for theme.Accent or theme.Success, and the theme decides which
// ANSI sequence to emit. This makes it trivial to add new palettes (Dracula,
// Nord, Solarized, monochrome, ascii-only) without changing module code.
//
// Themes are intentionally simple: a struct of named ANSI strings. Modules
// must always close colored regions with theme.Reset.
package theme

import (
	"fmt"
	"sort"
	"strings"
)

// Reset is the universal ANSI reset sequence. Always available regardless
// of the active theme so callers don't need to import a specific theme to
// reset attributes.
const Reset = "\033[0m"

// Style attributes that are independent of the palette. A theme may mark
// any of these empty (e.g. ascii-only) to suppress the attribute.
type Attrs struct {
	Bold      string
	Dim       string
	Italic    string
	Underline string
}

// Palette holds the semantic colors of a theme. Each field is a raw ANSI
// escape sequence (foreground only). Background colors are handled via
// dedicated fields when needed.
type Palette struct {
	// Foundational
	Foreground string // primary text color
	Muted      string // secondary / dim text
	Subtle     string // even dimmer (separators, hints)

	// Semantic
	Accent    string // primary accent (titles, focused items)
	Secondary string // secondary accent (alternative emphasis)
	Success   string // OK, active, available
	Warning   string // pending, partial
	Danger    string // failed, critical, missing dependency
	Info      string // neutral informational

	// Module-section colors (used by Header). Modules may pick any.
	SectionSystem    string
	SectionResources string
	SectionNetwork   string
	SectionWeather   string
	SectionContainer string
	SectionService   string
	SectionUpdate    string
	SectionLogin     string
	SectionCalendar  string
	SectionQuote     string
	SectionArt       string
	SectionLogo      string

	// Progress bar gradient stops (low → mid → high)
	GradientLow  string
	GradientMid  string
	GradientHigh string
}

// Theme bundles a palette with its attributes and presentation toggles.
type Theme struct {
	ID          string
	Name        string
	Description string
	Palette     Palette
	Attrs       Attrs
	// UseUnicode controls whether glyph helpers may emit non-ASCII characters.
	// ascii-only themes set this to false; bars become "#"/"-" and dots "*".
	UseUnicode bool
}

// Color returns a colored copy of s, automatically appending Reset.
// If color is empty, s is returned unchanged. Safe for nested usage.
func (t *Theme) Color(s, color string) string {
	if color == "" || s == "" {
		return s
	}
	return color + s + Reset
}

// Bold wraps s in the theme's bold attribute (or returns s unchanged when
// the theme suppresses bold).
func (t *Theme) Bold(s string) string {
	if t.Attrs.Bold == "" {
		return s
	}
	return t.Attrs.Bold + s + Reset
}

// Dim wraps s in the theme's dim attribute.
func (t *Theme) Dim(s string) string {
	if t.Attrs.Dim == "" {
		return s
	}
	return t.Attrs.Dim + s + Reset
}

// Italic wraps s in the theme's italic attribute.
func (t *Theme) Italic(s string) string {
	if t.Attrs.Italic == "" {
		return s
	}
	return t.Attrs.Italic + s + Reset
}

// SectionColor returns the appropriate accent color for a known module name,
// falling back to the theme accent if the module isn't recognized.
func (t *Theme) SectionColor(moduleName string) string {
	switch strings.ToLower(moduleName) {
	case "system":
		return t.Palette.SectionSystem
	case "resources":
		return t.Palette.SectionResources
	case "network":
		return t.Palette.SectionNetwork
	case "weather":
		return t.Palette.SectionWeather
	case "containers":
		return t.Palette.SectionContainer
	case "services":
		return t.Palette.SectionService
	case "updates":
		return t.Palette.SectionUpdate
	case "logins":
		return t.Palette.SectionLogin
	case "calendar":
		return t.Palette.SectionCalendar
	case "quote":
		return t.Palette.SectionQuote
	case "cowsay":
		return t.Palette.SectionArt
	case "logo":
		return t.Palette.SectionLogo
	default:
		return t.Palette.Accent
	}
}

// PercentColor returns a color appropriate for a numeric percentage:
// success (low), warning (mid), danger (high). Thresholds at 60 and 85.
func (t *Theme) PercentColor(percent float64) string {
	switch {
	case percent >= 85:
		return t.Palette.Danger
	case percent >= 60:
		return t.Palette.Warning
	default:
		return t.Palette.Success
	}
}

// GradientAt returns the gradient color for position (0..1) at a given
// percentage. Used by ProgressBar to interpolate green→yellow→red.
func (t *Theme) GradientAt(pos float64, percent float64) string {
	switch {
	case percent < 50:
		return t.Palette.GradientLow
	case percent < 70:
		if pos < 0.6 {
			return t.Palette.GradientLow
		}
		return t.Palette.GradientMid
	case percent < 85:
		if pos < 0.3 {
			return t.Palette.GradientLow
		}
		if pos < 0.7 {
			return t.Palette.GradientMid
		}
		return t.Palette.GradientHigh
	default:
		if pos < 0.2 {
			return t.Palette.GradientMid
		}
		return t.Palette.GradientHigh
	}
}

// Status returns a (color, label) pair for a service/status string.
func (t *Theme) Status(status string) (color string, label string) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active", "running", "ok", "up":
		return t.Palette.Success, "active"
	case "inactive", "stopped", "dead", "down":
		return t.Palette.Danger, "inactive"
	case "failed", "error":
		return t.Palette.Warning, "failed"
	default:
		return t.Palette.Subtle, status
	}
}

// --- Registry ---

var registry = map[string]*Theme{}

// Register adds a theme to the global registry. Safe to call from init().
func Register(t *Theme) {
	if t == nil || t.ID == "" {
		return
	}
	registry[t.ID] = t
}

// Get returns the theme with the given ID, or the default theme if not found.
// The boolean indicates whether the requested theme existed.
func Get(id string) (*Theme, bool) {
	if t, ok := registry[id]; ok {
		return t, true
	}
	return Default(), false
}

// MustGet returns the theme with the given ID, or the default theme silently.
// Use Get when you need to know whether the lookup succeeded.
func MustGet(id string) *Theme {
	t, _ := Get(id)
	return t
}

// All returns all registered themes sorted by ID.
func All() []*Theme {
	out := make([]*Theme, 0, len(registry))
	for _, t := range registry {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// IDs returns the sorted list of registered theme IDs.
func IDs() []string {
	ids := make([]string, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// Default returns the default theme. Always non-nil.
func Default() *Theme {
	if t, ok := registry["default"]; ok {
		return t
	}
	// Fallback: synthesize a minimal palette so unit tests work even if
	// the package's init() hasn't run yet.
	return &Theme{
		ID:         "default",
		Name:       "Default",
		Palette:    Palette{Foreground: "\033[97m", Accent: "\033[96m", Success: "\033[92m", Warning: "\033[93m", Danger: "\033[91m", Info: "\033[94m"},
		Attrs:      Attrs{Bold: "\033[1m", Dim: "\033[2m", Italic: "\033[3m", Underline: "\033[4m"},
		UseUnicode: true,
	}
}

// Describe returns a one-line human-readable summary of a theme, useful for
// the theme picker view.
func Describe(t *Theme) string {
	return fmt.Sprintf("%-14s %s", t.Name, t.Description)
}
