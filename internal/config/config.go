// Package config manages the TOML configuration for mom.
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the full configuration structure.
type Config struct {
	Motd    MotdConfig    `toml:"motd"`
	Modules ModulesConfig `toml:"modules"`
	Mode    ModeConfig    `toml:"mode"`
}

// MotdConfig holds global MOTD display options.
type MotdConfig struct {
	Header string `toml:"header"`
	Footer string `toml:"footer"`
}

// ModulesConfig holds the enabled/disabled state and per-module settings.
type ModulesConfig struct {
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

	// Per-module configuration
	WeatherConfig    WeatherModuleConfig    `toml:"weather_config"`
	ResourcesConfig  ResourcesModuleConfig  `toml:"resources_config"`
	CowsayConfig    CowsayModuleConfig     `toml:"cowsay_config"`
	UpdatesConfig   UpdatesModuleConfig    `toml:"updates_config"`
	ContainersConfig ContainersModuleConfig `toml:"containers_config"`
}

// WeatherModuleConfig holds weather module settings.
type WeatherModuleConfig struct {
	City  string `toml:"city"`
	Units string `toml:"units"` // metric | imperial
}

// ResourcesModuleConfig holds resources module settings.
type ResourcesModuleConfig struct {
	ShowTemp bool `toml:"show_temp"`
}

// CowsayModuleConfig holds cowsay module settings.
type CowsayModuleConfig struct {
	Mode    string `toml:"mode"`    // cowsay | figlet | lolcat | random
	Message string `toml:"message"`
}

// UpdatesModuleConfig holds updates module settings.
type UpdatesModuleConfig struct {
	IncludeAUR   bool `toml:"include_aur"`
	IncludeSnaps bool `toml:"include_snaps"`
}

// ContainersModuleConfig holds containers module settings.
type ContainersModuleConfig struct {
	Runtime string `toml:"runtime"` // auto | docker | podman
}

// ModeConfig holds operational mode settings.
type ModeConfig struct {
	Default      string `toml:"default"`       // manual | template | auto | full-auto
	LastTemplate string `toml:"last_template"`
}

// IsModuleEnabled returns whether a module is enabled by name.
func (c *Config) IsModuleEnabled(name string) bool {
	switch name {
	case "system":
		return c.Modules.System
	case "resources":
		return c.Modules.Resources
	case "weather":
		return c.Modules.Weather
	case "cowsay":
		return c.Modules.Cowsay
	case "network":
		return c.Modules.Network
	case "containers":
		return c.Modules.Containers
	case "updates":
		return c.Modules.Updates
	case "logins":
		return c.Modules.Logins
	case "quote":
		return c.Modules.Quote
	case "calendar":
		return c.Modules.Calendar
	case "services":
		return c.Modules.Services
	case "logo":
		return c.Modules.Logo
	default:
		return false
	}
}

// SetModuleEnabled sets a module's enabled state by name.
func (c *Config) SetModuleEnabled(name string, enabled bool) {
	switch name {
	case "system":
		c.Modules.System = enabled
	case "resources":
		c.Modules.Resources = enabled
	case "weather":
		c.Modules.Weather = enabled
	case "cowsay":
		c.Modules.Cowsay = enabled
	case "network":
		c.Modules.Network = enabled
	case "containers":
		c.Modules.Containers = enabled
	case "updates":
		c.Modules.Updates = enabled
	case "logins":
		c.Modules.Logins = enabled
	case "quote":
		c.Modules.Quote = enabled
	case "calendar":
		c.Modules.Calendar = enabled
	case "services":
		c.Modules.Services = enabled
	case "logo":
		c.Modules.Logo = enabled
	}
}

// EnabledModuleNames returns the list of enabled module names.
func (c *Config) EnabledModuleNames() []string {
	var names []string
	// Preserve insertion order matching MOTD output order
	modules := []struct {
		name    string
		enabled bool
	}{
		{"logo", c.Modules.Logo},
		{"system", c.Modules.System},
		{"resources", c.Modules.Resources},
		{"network", c.Modules.Network},
		{"weather", c.Modules.Weather},
		{"containers", c.Modules.Containers},
		{"services", c.Modules.Services},
		{"updates", c.Modules.Updates},
		{"logins", c.Modules.Logins},
		{"calendar", c.Modules.Calendar},
		{"quote", c.Modules.Quote},
		{"cowsay", c.Modules.Cowsay},
	}
	for _, m := range modules {
		if m.enabled {
			names = append(names, m.name)
		}
	}
	return names
}

// configPath returns the path to the config file.
func configPath() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "mom", "config.toml")
}

// Load reads the configuration from disk. Returns defaults if file does not exist.
func Load() (*Config, error) {
	return LoadFrom(configPath())
}

// LoadFrom reads configuration from a specific path.
func LoadFrom(path string) (*Config, error) {
	cfg := Defaults()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if _, err := toml.Decode(string(data), cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	cfg.Validate()
	return cfg, nil
}

// Save writes the configuration to disk.
func Save(cfg *Config) error {
	return SaveTo(cfg, configPath())
}

// SaveTo writes the configuration to a specific path.
func SaveTo(cfg *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

// Validate corrects invalid values in the config, forcing them to safe defaults.
func (c *Config) Validate() {
	// Weather units
	if c.Modules.WeatherConfig.Units != "metric" && c.Modules.WeatherConfig.Units != "imperial" {
		c.Modules.WeatherConfig.Units = "metric"
	}

	// Cowsay mode
	validCowsay := map[string]bool{"cowsay": true, "figlet": true, "lolcat": true, "random": true}
	if !validCowsay[c.Modules.CowsayConfig.Mode] {
		c.Modules.CowsayConfig.Mode = "cowsay"
	}

	// Containers runtime
	validRuntime := map[string]bool{"auto": true, "docker": true, "podman": true}
	if !validRuntime[c.Modules.ContainersConfig.Runtime] {
		c.Modules.ContainersConfig.Runtime = "auto"
	}

	// Mode default
	validMode := map[string]bool{"manual": true, "template": true, "auto": true, "full-auto": true}
	if !validMode[c.Mode.Default] {
		c.Mode.Default = "manual"
	}
}
