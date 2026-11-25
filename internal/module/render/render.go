// Package render provides visual utilities for MOTD module output.
//
// The Renderer type (see renderer.go) is the single API: it takes an
// Options{Theme, Variant, Width} and exposes Header / KeyValue /
// ProgressBar / AsciiBanner methods so modules never paste ANSI codes
// directly.
package render

import "fmt"

// FormatBytes formats bytes into a human readable string.
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
