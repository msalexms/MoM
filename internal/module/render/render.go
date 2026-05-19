// Package render provides visual utilities for MOTD module output.
//
// The preferred API is the Renderer type (see renderer.go), which takes an
// Options{Theme, Variant, Width} and exposes Header/KeyValue/ProgressBar/
// AsciiBanner methods. The constants and free functions in this file are
// kept as a thin compatibility layer for legacy callers that still emit
// output using the default theme directly.
package render

import (
	"fmt"
)

// Reset and ANSI constants kept for backwards compatibility. New code should
// use the Renderer from renderer.go and let the active theme decide colors.
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"

	FgBlack   = "\033[30m"
	FgRed     = "\033[31m"
	FgGreen   = "\033[32m"
	FgYellow  = "\033[33m"
	FgBlue    = "\033[34m"
	FgMagenta = "\033[35m"
	FgCyan    = "\033[36m"
	FgWhite   = "\033[37m"

	FgBrightRed     = "\033[91m"
	FgBrightGreen   = "\033[92m"
	FgBrightYellow  = "\033[93m"
	FgBrightBlue    = "\033[94m"
	FgBrightMagenta = "\033[95m"
	FgBrightCyan    = "\033[96m"
	FgBrightWhite   = "\033[97m"

	BgRed    = "\033[41m"
	BgGreen  = "\033[42m"
	BgYellow = "\033[43m"
	BgBlue   = "\033[44m"
)

// Icons - simple ASCII that works in any terminal.
const (
	IconCPU      = "cpu"
	IconRAM      = "mem"
	IconDisk     = "disk"
	IconNet      = "net"
	IconWeather  = "wx"
	IconDocker   = "ctr"
	IconService  = "svc"
	IconUser     = "usr"
	IconClock    = "clk"
	IconCalendar = "cal"
	IconQuote    = "qt"
	IconLinux    = "sys"
	IconShell    = "sh"
	IconHost     = "host"
	IconKernel   = "kern"
	IconUp       = "up"
	IconDown     = "dn"
	IconCheck    = "+"
	IconCross    = "-"
	IconWarning  = "!"
	IconBullet   = "*"
	IconArrow    = ">"
	IconStar     = "*"
	IconTemp     = "tmp"
)

// defaultRenderer is a singleton renderer using the default theme. It powers
// the legacy free functions below.
var defaultRenderer = New(DefaultOptions())

// Header renders a section header with the default theme. Deprecated: prefer
// Renderer.Header / Renderer.HeaderColor.
func Header(title, color string) string {
	return defaultRenderer.HeaderColor(title, color)
}

// KeyValue renders an aligned key:value pair with the default theme.
// Deprecated: prefer Renderer.KeyValue.
func KeyValue(key, value, keyColor, valColor string) string {
	// Legacy signature took explicit colors; we emit them verbatim so existing
	// modules keep producing identical output until they are migrated.
	return fmt.Sprintf("    %s%-10s%s  %s%s", keyColor, key, Reset, valColor, value+Reset)
}

// StatusDot renders a status indicator with the default theme. Deprecated:
// prefer Renderer.StatusDot.
func StatusDot(status string) string {
	return defaultRenderer.StatusDot(status)
}

// Colorize wraps text in a color code.
func Colorize(text, color string) string { return color + text + Reset }

// BoldColor wraps text in bold + color.
func BoldColor(text, color string) string { return Bold + color + text + Reset }

// FormatBytes formats bytes into human readable string.
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%s", float64(bytes)/float64(div), []string{"KB", "MB", "GB", "TB"}[exp])
}

// ProgressBar renders a gradient progress bar with the default theme.
// Deprecated: prefer Renderer.ProgressBar.
func ProgressBar(percent float64, width int, label string) string {
	return defaultRenderer.ProgressBar(percent, width, label)
}

// AsciiBanner renders text as 5x5 ASCII block letters with the given color.
// Deprecated: prefer Renderer.AsciiBanner (color comes from the theme).
func AsciiBanner(text string, color string) string {
	if text == "" {
		return ""
	}
	// Use the legacy explicit-color path so existing modules look identical.
	r := New(Options{Theme: defaultRenderer.Theme(), Variant: VariantDefault})
	out := r.AsciiBanner(text)
	if color == "" {
		return out
	}
	// The renderer used the theme accent; if a caller passed a different
	// color we recolor by replacing the accent escapes. Cheap and good enough
	// for the legacy path.
	return replaceColor(out, defaultRenderer.Theme().Palette.Accent, color)
}

func replaceColor(s, oldEsc, newEsc string) string {
	if oldEsc == "" || oldEsc == newEsc {
		return s
	}
	out := make([]byte, 0, len(s))
	i := 0
	for i < len(s) {
		if i+len(oldEsc) <= len(s) && s[i:i+len(oldEsc)] == oldEsc {
			out = append(out, newEsc...)
			i += len(oldEsc)
			continue
		}
		out = append(out, s[i])
		i++
	}
	return string(out)
}
