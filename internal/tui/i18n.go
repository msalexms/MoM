package tui

import "github.com/msalexms/MoM/internal/i18n"

// T and Tf translate interface text into the active language. Tables that
// are built once at start-up keep English strings and are passed through T
// when rendered, so a language change takes effect immediately.
func T(s string) string { return i18n.T(s) }

// Tf translates format and formats it like fmt.Sprintf.
func Tf(format string, args ...any) string { return i18n.Tf(format, args...) }
