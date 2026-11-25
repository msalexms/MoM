package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProfileDir returns the directory where profiles are stored.
func ProfileDir() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "mom", "profiles")
}

// SaveProfile saves the current config as a named profile.
func SaveProfile(cfg *Config, name string) error {
	dir := ProfileDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating profiles dir: %w", err)
	}
	path := filepath.Join(dir, name+".toml")
	return SaveTo(cfg, path)
}

// LoadProfile loads a named profile and returns it as a Config.
func LoadProfile(name string) (*Config, error) {
	path := filepath.Join(ProfileDir(), name+".toml")
	return LoadFrom(path)
}

// ListProfiles returns the names of all saved profiles.
func ListProfiles() []string {
	dir := ProfileDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".toml") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".toml")
		names = append(names, name)
	}
	return names
}

// DeleteProfile removes a named profile.
func DeleteProfile(name string) error {
	path := filepath.Join(ProfileDir(), name+".toml")
	return os.Remove(path)
}

// SanitizeProfileName turns free text into a safe file name: spaces become
// dashes and anything other than letters, digits, - and _ is dropped. It
// returns "" when nothing usable is left.
func SanitizeProfileName(name string) string {
	var sb strings.Builder
	for _, r := range strings.TrimSpace(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			sb.WriteRune(r)
		case r == ' ':
			sb.WriteRune('-')
		}
	}
	return sb.String()
}
