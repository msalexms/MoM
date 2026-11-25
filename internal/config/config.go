// Package config manages the TOML configuration for mom.
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/msalexms/MoM/internal/i18n"

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
	Timers     bool `toml:"timers"`
	Journal    bool `toml:"journal"`
	ZFS        bool `toml:"zfs"`
	Certs      bool `toml:"certs"`
	Firewall   bool `toml:"firewall"`
	FailedLog  bool `toml:"failed_logins"`
	Sudo       bool `toml:"sudo"`
	Git        bool `toml:"git"`
	Tmux       bool `toml:"tmux"`
	Reboot     bool `toml:"reboot"`

	// Per-module configuration
	WeatherConfig    WeatherModuleConfig    `toml:"weather_config"`
	ResourcesConfig  ResourcesModuleConfig  `toml:"resources_config"`
	CowsayConfig     CowsayModuleConfig     `toml:"cowsay_config"`
	UpdatesConfig    UpdatesModuleConfig    `toml:"updates_config"`
	ContainersConfig ContainersModuleConfig `toml:"containers_config"`
	ServicesConfig   ServicesModuleConfig   `toml:"services_config"`
	GitConfig        GitModuleConfig        `toml:"git_config"`
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
	Mode    string `toml:"mode"` // cowsay | figlet | ascii-art | random
	Message string `toml:"message"`
}

// UpdatesModuleConfig holds updates module settings.
type UpdatesModuleConfig struct {
	IncludeAUR bool `toml:"include_aur"`
}

// ContainersModuleConfig holds containers module settings.
type ContainersModuleConfig struct {
	Runtime string `toml:"runtime"` // auto | docker | podman
}

// ServicesModuleConfig holds services module settings.
type ServicesModuleConfig struct {
	Services []string `toml:"services"` // user-selected services to monitor
}

// GitModuleConfig holds git status module settings.
type GitModuleConfig struct {
	// Paths is the list of directories to scan for git repos.
	// Each entry may contain a leading ~/ which is expanded to $HOME.
	// When empty, defaults to ~/projects, ~/repos, ~/src.
	Paths    []string `toml:"paths"`
	MaxRepos int      `toml:"max_repos"` // maximum dirty repos to display (default 5)
}

// ModeConfig holds operational mode settings.
type ModeConfig struct {
	Language    string   `toml:"language"`     // interface and MOTD language: en | es
	Theme       string   `toml:"theme"`        // theme ID (amstech, default, dracula, nord, etc.)
	Variant     string   `toml:"variant"`      // global default variant
	ModuleOrder []string `toml:"module_order"` // custom module output order
}

// --- Module order (used for MOTD output ordering) ---

// moduleOrder defines the canonical output order of modules in the MOTD.
var moduleOrder = []string{
	"logo", "system", "resources", "gpu", "network", "traffic", "weather",
	"containers", "services", "ports", "procs", "diskio",
	"timers", "journal", "zfs", "certs",
	"firewall", "fail2ban", "failed-logins", "sudo",
	"battery", "users", "sshkeys", "boot",
	"git", "tmux", "reboot",
	"updates", "logins", "calendar", "quote", "cowsay",
}

// DefaultModuleOrder returns a copy of the canonical module order.
func DefaultModuleOrder() []string {
	out := make([]string, len(moduleOrder))
	copy(out, moduleOrder)
	return out
}

// moduleFlags maps every module name to the enabled flag inside ModulesConfig.
// IsModuleEnabled, SetModuleEnabled and the template system all go through
// this single mapping, so a new module only needs one entry here (plus the
// TOML field) instead of three synchronized switches.
var moduleFlags = map[string]func(*ModulesConfig) *bool{
	"system":        func(m *ModulesConfig) *bool { return &m.System },
	"resources":     func(m *ModulesConfig) *bool { return &m.Resources },
	"weather":       func(m *ModulesConfig) *bool { return &m.Weather },
	"cowsay":        func(m *ModulesConfig) *bool { return &m.Cowsay },
	"network":       func(m *ModulesConfig) *bool { return &m.Network },
	"containers":    func(m *ModulesConfig) *bool { return &m.Containers },
	"updates":       func(m *ModulesConfig) *bool { return &m.Updates },
	"logins":        func(m *ModulesConfig) *bool { return &m.Logins },
	"quote":         func(m *ModulesConfig) *bool { return &m.Quote },
	"calendar":      func(m *ModulesConfig) *bool { return &m.Calendar },
	"services":      func(m *ModulesConfig) *bool { return &m.Services },
	"logo":          func(m *ModulesConfig) *bool { return &m.Logo },
	"ports":         func(m *ModulesConfig) *bool { return &m.Ports },
	"procs":         func(m *ModulesConfig) *bool { return &m.Procs },
	"fail2ban":      func(m *ModulesConfig) *bool { return &m.Fail2ban },
	"battery":       func(m *ModulesConfig) *bool { return &m.Battery },
	"users":         func(m *ModulesConfig) *bool { return &m.Users },
	"sshkeys":       func(m *ModulesConfig) *bool { return &m.SSHKeys },
	"boot":          func(m *ModulesConfig) *bool { return &m.Boot },
	"gpu":           func(m *ModulesConfig) *bool { return &m.GPU },
	"traffic":       func(m *ModulesConfig) *bool { return &m.Traffic },
	"diskio":        func(m *ModulesConfig) *bool { return &m.DiskIO },
	"timers":        func(m *ModulesConfig) *bool { return &m.Timers },
	"journal":       func(m *ModulesConfig) *bool { return &m.Journal },
	"zfs":           func(m *ModulesConfig) *bool { return &m.ZFS },
	"certs":         func(m *ModulesConfig) *bool { return &m.Certs },
	"firewall":      func(m *ModulesConfig) *bool { return &m.Firewall },
	"failed-logins": func(m *ModulesConfig) *bool { return &m.FailedLog },
	"sudo":          func(m *ModulesConfig) *bool { return &m.Sudo },
	"git":           func(m *ModulesConfig) *bool { return &m.Git },
	"tmux":          func(m *ModulesConfig) *bool { return &m.Tmux },
	"reboot":        func(m *ModulesConfig) *bool { return &m.Reboot },
}

// IsModuleEnabled returns whether a module is enabled by name.
func (c *Config) IsModuleEnabled(name string) bool {
	if flag, ok := moduleFlags[name]; ok {
		return *flag(&c.Modules)
	}
	return false
}

// SetModuleEnabled sets a module's enabled state by name.
// Unknown names are ignored.
func (c *Config) SetModuleEnabled(name string, enabled bool) {
	if flag, ok := moduleFlags[name]; ok {
		*flag(&c.Modules) = enabled
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
		if c.IsModuleEnabled(name) && !slices.Contains(names, name) {
			names = append(names, name)
		}
	}
	return names
}

// KnownModuleNames returns the names of every module the config knows about.
// The template system uses it to apply and export the complete module set.
func KnownModuleNames() []string {
	return DefaultModuleOrder()
}

// DefaultTheme is the theme used when none is configured.
const DefaultTheme = "amstech"

// ThemeID returns the configured theme ID, defaulting to DefaultTheme.
func (c *Config) ThemeID() string {
	if c.Mode.Theme == "" {
		return DefaultTheme
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

// SystemPath is the copy of the config installed for the login hook. It
// is what "mom --motd" renders, so every user sees the installed MOTD
// regardless of their own ~/.config.
const SystemPath = "/etc/mom/config.toml"

// Load reads the user configuration. It returns defaults if there is none.
func Load() (*Config, error) {
	return LoadFrom(configPath())
}

// LoadFrom reads configuration from a specific path, returning defaults
// when the file does not exist.
func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Defaults(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	return Decode(data)
}

// Decode parses a TOML config on top of the defaults and validates it.
func Decode(data []byte) (*Config, error) {
	cfg := Defaults()
	if _, err := toml.Decode(string(data), cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	cfg.Validate()
	return cfg, nil
}

// Encode serializes the config as TOML.
func Encode(cfg *Config) ([]byte, error) {
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		return nil, fmt.Errorf("encoding config: %w", err)
	}
	return buf.Bytes(), nil
}

// Save writes the user configuration.
func Save(cfg *Config) error {
	return SaveTo(cfg, configPath())
}

// SaveTo writes the configuration to a specific path.
func SaveTo(cfg *Config, path string) error {
	data, err := Encode(cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// Validate corrects invalid values in the config, forcing them to safe defaults.
func (c *Config) Validate() {
	c.Mode.Language = string(i18n.Parse(c.Mode.Language))

	// Weather units
	if c.Modules.WeatherConfig.Units != "metric" && c.Modules.WeatherConfig.Units != "imperial" {
		c.Modules.WeatherConfig.Units = "metric"
	}

	// Cowsay mode — must stay in sync with CowsayModule.Settings().
	if !slices.Contains(validCowsayModes, c.Modules.CowsayConfig.Mode) {
		c.Modules.CowsayConfig.Mode = "ascii-art"
	}

	// Containers runtime
	if !slices.Contains(validRuntimes, c.Modules.ContainersConfig.Runtime) {
		c.Modules.ContainersConfig.Runtime = "auto"
	}

	// Git max repos
	if c.Modules.GitConfig.MaxRepos <= 0 {
		c.Modules.GitConfig.MaxRepos = 5
	}
}

// validCowsayModes matches the enum exposed by the cowsay module settings.
var validCowsayModes = []string{"cowsay", "figlet", "ascii-art", "random"}

// validRuntimes matches the enum exposed by the containers module settings.
var validRuntimes = []string{"auto", "docker", "podman"}

// SaveLanguage stores only the language in the config file on disk,
// leaving every other saved value as it is.
func SaveLanguage(lang string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.Mode.Language = lang
	return Save(cfg)
}
