// Package template manages built-in and user-created MOTD templates.
package template

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"

	embeds "github.com/msalexms/MoM/embed"
	"github.com/msalexms/MoM/internal/config"
)

// Template represents a MOTD configuration profile. The [modules] table
// holds one boolean per module name; every known module is present, so
// applying a template always produces the exact module set the template
// was designed with.
type Template struct {
	Name        string            `toml:"name"`
	Description string            `toml:"description"`
	Author      string            `toml:"author,omitempty"`
	Motd        config.MotdConfig `toml:"motd"`
	Modules     map[string]bool   `toml:"modules"`
}

// Apply applies a template to the given configuration.
// It overwrites module enabled/disabled states and motd header/footer,
// but does NOT modify per-module settings (city, units, runtime, etc.).
func Apply(t *Template, cfg *config.Config) {
	cfg.Motd.Header = t.Motd.Header
	cfg.Motd.Footer = t.Motd.Footer

	for _, name := range config.KnownModuleNames() {
		enabled, present := t.Modules[name]
		cfg.SetModuleEnabled(name, present && enabled)
	}
}

// BuiltinTemplates returns all built-in templates embedded in the binary.
func BuiltinTemplates() ([]*Template, error) {
	entries, err := embeds.TemplatesFS.ReadDir("templates")
	if err != nil {
		return nil, fmt.Errorf("reading embedded templates: %w", err)
	}

	var templates []*Template
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := embeds.TemplatesFS.ReadFile("templates/" + entry.Name())
		if err != nil {
			continue
		}
		var t Template
		if _, err := toml.Decode(string(data), &t); err != nil {
			continue
		}
		templates = append(templates, &t)
	}

	return templates, nil
}

// GetBuiltin returns a built-in template by name (case-insensitive).
func GetBuiltin(name string) (*Template, error) {
	templates, err := BuiltinTemplates()
	if err != nil {
		return nil, err
	}
	for _, t := range templates {
		if strings.EqualFold(t.Name, name) {
			return t, nil
		}
	}
	return nil, fmt.Errorf("template %q not found", name)
}

// BuiltinNames returns the list of built-in template names.
func BuiltinNames() ([]string, error) {
	templates, err := BuiltinTemplates()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(templates))
	for i, t := range templates {
		names[i] = t.Name
	}
	return names, nil
}
