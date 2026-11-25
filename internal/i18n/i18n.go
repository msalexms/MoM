// Package i18n translates user-facing text. Text is written in English in
// the code and looked up in a per-language catalog; a string without a
// translation is shown as is, so English never needs a catalog.
package i18n

import (
	"fmt"
	"sync/atomic"
)

// Lang is a supported interface language.
type Lang string

const (
	EN Lang = "en"
	ES Lang = "es"
)

// Default is the language used when none is configured.
const Default = EN

// Languages lists the supported languages with their own names, in the
// order the language picker shows them.
var Languages = []struct {
	ID   Lang
	Name string
}{
	{EN, "English"},
	{ES, "Español"},
}

var catalogs = map[Lang]map[string]string{ES: es}

var current atomic.Value // Lang

func init() { current.Store(Default) }

// Parse returns the language for id, or Default if it is not supported.
func Parse(id string) Lang {
	for _, l := range Languages {
		if string(l.ID) == id {
			return l.ID
		}
	}
	return Default
}

// Set changes the active language. Safe to call while modules render.
func Set(l Lang) { current.Store(Parse(string(l))) }

// Current returns the active language.
func Current() Lang { return current.Load().(Lang) }

// T translates s into the active language.
func T(s string) string {
	if v, ok := catalogs[Current()][s]; ok {
		return v
	}
	return s
}

// Tf translates format and then formats it like fmt.Sprintf.
func Tf(format string, args ...any) string {
	return fmt.Sprintf(T(format), args...)
}
