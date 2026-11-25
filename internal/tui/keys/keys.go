// Package keys defines the key bindings of the TUI.
package keys

import "github.com/charmbracelet/bubbles/key"

// KeyMap holds every binding. Views match against these instead of raw
// strings so a key is defined in exactly one place.
type KeyMap struct {
	Up, Down, Enter, Space, Back, Help, Quit, Save key.Binding
	AllOn, AllOff, MoveUp, MoveDown, Add, Remove   key.Binding
}

// DefaultKeyMap returns the default bindings.
func DefaultKeyMap() KeyMap {
	b := func(keys ...string) key.Binding { return key.NewBinding(key.WithKeys(keys...)) }
	return KeyMap{
		Up:       b("up", "k"),
		Down:     b("down", "j"),
		Enter:    b("enter"),
		Space:    b(" "),
		Back:     b("esc"),
		Help:     b("?"),
		Quit:     b("q"),
		Save:     b("s", "ctrl+s"),
		AllOn:    b("a"),
		AllOff:   b("d"),
		MoveUp:   b("shift+up", "K"),
		MoveDown: b("shift+down", "J"),
		Add:      b("a"),
		Remove:   b("x", "delete"),
	}
}
