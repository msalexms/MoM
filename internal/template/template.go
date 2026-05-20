// Package template manages built-in and user-created MOTD templates.
package template

import (
	"fmt"

	"github.com/BurntSushi/toml"

	embeds "github.com/ams/mom/embed"
	"github.com/ams/mom/internal/config"
)

// Template represents a MOTD configuration profile.
type Template struct {
	Name        string            `toml:"name"`
	Description string            `toml:"description"`
	Author      string            `toml:"author,omitempty"`
	Motd        config.MotdConfig `toml:"motd"`
	Modules     moduleMap         `toml:"modules"`
}

// moduleMap maps module names to enabled/disabled state.
type moduleMap struct {
	System     bool `toml:"system"`
	Resources  bool `toml:"resources"`
	Weather    bool `toml:"weather"`
	Cowsay     bool `toml:"cowsay"`
	Network    bool `toml:"network"`
	Containers bool `toml:"containers"`
	Updates    bool `toml:"updates"`
	Logins     bool `toml:"logins"`
	Quote      bool `toml:"quote"`
	Calendar   bool `toml:"calendar"`
	Services   bool `toml:"services"`
	Logo       bool `toml:"logo"`
}

// Apply applies a template to the given configuration.
// It overwrites module enabled/disabled states and motd header/footer,
// but does NOT modify per-module settings (city, units, runtime, etc.).
func Apply(t *Template, cfg *config.Config) {
	cfg.Motd.Header = t.Motd.Header
	cfg.Motd.Footer = t.Motd.Footer

	cfg.Modules.System = t.Modules.System
	cfg.Modules.Resources = t.Modules.Resources
	cfg.Modules.Weather = t.Modules.Weather
	cfg.Modules.Cowsay = t.Modules.Cowsay
	cfg.Modules.Network = t.Modules.Network
	cfg.Modules.Containers = t.Modules.Containers
	cfg.Modules.Updates = t.Modules.Updates
	cfg.Modules.Logins = t.Modules.Logins
	cfg.Modules.Quote = t.Modules.Quote
	cfg.Modules.Calendar = t.Modules.Calendar
	cfg.Modules.Services = t.Modules.Services
	cfg.Modules.Logo = t.Modules.Logo
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

// GetBuiltin returns a built-in template by name.
func GetBuiltin(name string) (*Template, error) {
	templates, err := BuiltinTemplates()
	if err != nil {
		return nil, err
	}
	for _, t := range templates {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, fmt.Errorf("template %q not found", name)
}
