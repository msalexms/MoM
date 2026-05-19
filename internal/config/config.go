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
	Ports      bool `toml:"ports"`
	Procs      bool `toml:"procs"`
	Fail2ban   bool `toml:"fail2ban"`
	Battery    bool `toml:"battery"`
	Users      bool `toml:"users"`
	SSHKeys    bool `toml:"sshkeys"`
	Boot       bool `toml:"boot"`
	GPU        bool `toml:"gpu"`
	Traffic    bool `toml:"traffic"`
	DiskIO     bool `toml:"diskio"`

	// Per-module configuration
	WeatherConfig    WeatherModuleConfig    `toml:"weather_config"`
	ResourcesConfig  ResourcesModuleConfig  `toml:"resources_config"`
	CowsayConfig     CowsayModuleConfig    `toml:"cowsay_config"`
	UpdatesConfig    UpdatesModuleConfig    `toml:"updates_config"`
	ContainersConfig ContainersModuleConfig `toml:"containers_config"`
	ServicesConfig   ServicesModuleConfig   `toml:"services_config"`
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

// ServicesModuleConfig holds services module settings.
type ServicesModuleConfig struct {
	Services []string `toml:"services"` // user-selected services to monitor
}

// ModeConfig holds operational mode settings.
type ModeConfig struct {
	Default      string   `toml:"default"`       // manual | template | auto | full-auto
	LastTemplate string   `toml:"last_template"`
	Theme        string   `toml:"theme"`         // theme ID (default, dracula, nord, etc.)
	Variant      string   `toml:"variant"`       // global default variant
	ModuleOrder  []string `toml:"module_order"`  // custom module output order
}

// --- Module order (used for MOTD output ordering) ---

// moduleOrder defines the canonical output order of modules in the MOTD.
var moduleOrder = []string{
	"logo", "system", "resources", "gpu", "network", "traffic", "weather",
	"containers", "services", "ports", "procs", "diskio", "fail2ban",
	"battery", "users", "sshkeys", "boot",
	"updates", "logins", "calendar", "quote", "cowsay",
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
	case "ports":
		return c.Modules.Ports
	case "procs":
		return c.Modules.Procs
	case "fail2ban":
		return c.Modules.Fail2ban
	case "battery":
		return c.Modules.Battery
	case "users":
		return c.Modules.Users
	case "sshkeys":
		return c.Modules.SSHKeys
	case "boot":
		return c.Modules.Boot
	case "gpu":
		return c.Modules.GPU
	case "traffic":
		return c.Modules.Traffic
	case "diskio":
		return c.Modules.DiskIO
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
	case "ports":
		c.Modules.Ports = enabled
	case "procs":
		c.Modules.Procs = enabled
	case "fail2ban":
		c.Modules.Fail2ban = enabled
	case "battery":
		c.Modules.Battery = enabled
	case "users":
		c.Modules.Users = enabled
	case "sshkeys":
		c.Modules.SSHKeys = enabled
	case "boot":
		c.Modules.Boot = enabled
	case "gpu":
		c.Modules.GPU = enabled
	case "traffic":
		c.Modules.Traffic = enabled
	case "diskio":
		c.Modules.DiskIO = enabled
	}
}

// EnabledModuleNames returns the list of enabled module names in MOTD order.
// If a custom ModuleOrder is set, it uses that; otherwise falls back to the
// canonical moduleOrder.
func (c *Config) EnabledModuleNames() []string {
	order := moduleOrder
	if len(c.Mode.ModuleOrder) > 0 {
		order = c.Mode.ModuleOrder
	}
	var names []string
	for _, name := range order {
		if c.IsModuleEnabled(name) {
			names = append(names, name)
		}
	}
	// Append any enabled modules not in the order list (safety net)
	for _, name := range moduleOrder {
		if c.IsModuleEnabled(name) && !contains(names, name) {
			names = append(names, name)
		}
	}
	return names
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// ThemeID returns the configured theme ID, defaulting to "default".
func (c *Config) ThemeID() string {
	if c.Mode.Theme == "" {
		return "default"
	}
	return c.Mode.Theme
}

// GlobalVariant returns the configured global variant, defaulting to "default".
func (c *Config) GlobalVariant() string {
	if c.Mode.Variant == "" {
		return "default"
	}
	return c.Mode.Variant
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
