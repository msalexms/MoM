package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_NonExistent_ReturnsDefaults(t *testing.T) {
	cfg, err := LoadFrom("/tmp/mom-test-nonexistent-config.toml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.Modules.System {
		t.Error("expected system module enabled by default")
	}
	if !cfg.Modules.Logo {
		t.Error("expected logo module enabled by default")
	}
	if cfg.Modules.Weather {
		t.Error("expected weather module disabled by default")
	}
	if cfg.Mode.Default != "manual" {
		t.Errorf("expected mode 'manual', got %q", cfg.Mode.Default)
	}
}

func TestLoad_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `[motd]
header = "Hello!"
footer = "Goodbye."

[modules]
system = true
resources = true
weather = true
cowsay = false
network = false
containers = false
updates = false
logins = false
quote = false
calendar = false
services = false
logo = true

[modules.weather_config]
city = "Madrid"
units = "metric"

[modules.cowsay_config]
mode = "figlet"
message = "Hi there"

[modules.containers_config]
runtime = "docker"

[mode]
default = "template"
last_template = "sysadmin"
`
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Motd.Header != "Hello!" {
		t.Errorf("expected header 'Hello!', got %q", cfg.Motd.Header)
	}
	if cfg.Motd.Footer != "Goodbye." {
		t.Errorf("expected footer 'Goodbye.', got %q", cfg.Motd.Footer)
	}
	if !cfg.Modules.Resources {
		t.Error("expected resources enabled")
	}
	if !cfg.Modules.Weather {
		t.Error("expected weather enabled")
	}
	if cfg.Modules.WeatherConfig.City != "Madrid" {
		t.Errorf("expected city 'Madrid', got %q", cfg.Modules.WeatherConfig.City)
	}
	if cfg.Modules.CowsayConfig.Mode != "figlet" {
		t.Errorf("expected cowsay mode 'figlet', got %q", cfg.Modules.CowsayConfig.Mode)
	}
	if cfg.Modules.ContainersConfig.Runtime != "docker" {
		t.Errorf("expected runtime 'docker', got %q", cfg.Modules.ContainersConfig.Runtime)
	}
	if cfg.Mode.Default != "template" {
		t.Errorf("expected mode 'template', got %q", cfg.Mode.Default)
	}
	if cfg.Mode.LastTemplate != "sysadmin" {
		t.Errorf("expected last_template 'sysadmin', got %q", cfg.Mode.LastTemplate)
	}
}

func TestSaveAndReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	cfg := Defaults()
	cfg.Motd.Header = "Test Header"
	cfg.Modules.Weather = true
	cfg.Modules.WeatherConfig.City = "Barcelona"
	cfg.Mode.Default = "auto"

	err := SaveTo(cfg, path)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("Load after save failed: %v", err)
	}

	if loaded.Motd.Header != "Test Header" {
		t.Errorf("header mismatch: %q", loaded.Motd.Header)
	}
	if !loaded.Modules.Weather {
		t.Error("expected weather enabled after reload")
	}
	if loaded.Modules.WeatherConfig.City != "Barcelona" {
		t.Errorf("city mismatch: %q", loaded.Modules.WeatherConfig.City)
	}
	if loaded.Mode.Default != "auto" {
		t.Errorf("mode mismatch: %q", loaded.Mode.Default)
	}
}

func TestValidate_BadValues(t *testing.T) {
	cfg := &Config{
		Modules: ModulesConfig{
			WeatherConfig:    WeatherModuleConfig{Units: "kelvin"},
			CowsayConfig:     CowsayModuleConfig{Mode: "invalid"},
			ContainersConfig: ContainersModuleConfig{Runtime: "containerd"},
		},
		Mode: ModeConfig{Default: "turbo"},
	}

	cfg.Validate()

	if cfg.Modules.WeatherConfig.Units != "metric" {
		t.Errorf("expected units forced to 'metric', got %q", cfg.Modules.WeatherConfig.Units)
	}
	if cfg.Modules.CowsayConfig.Mode != "cowsay" {
		t.Errorf("expected cowsay mode forced to 'cowsay', got %q", cfg.Modules.CowsayConfig.Mode)
	}
	if cfg.Modules.ContainersConfig.Runtime != "auto" {
		t.Errorf("expected runtime forced to 'auto', got %q", cfg.Modules.ContainersConfig.Runtime)
	}
	if cfg.Mode.Default != "manual" {
		t.Errorf("expected mode forced to 'manual', got %q", cfg.Mode.Default)
	}
}

func TestIsModuleEnabled(t *testing.T) {
	cfg := Defaults()

	if !cfg.IsModuleEnabled("system") {
		t.Error("system should be enabled by default")
	}
	if cfg.IsModuleEnabled("weather") {
		t.Error("weather should be disabled by default")
	}
	if cfg.IsModuleEnabled("nonexistent") {
		t.Error("nonexistent module should return false")
	}
}

func TestSetModuleEnabled(t *testing.T) {
	cfg := Defaults()

	cfg.SetModuleEnabled("weather", true)
	if !cfg.Modules.Weather {
		t.Error("weather should be enabled after SetModuleEnabled")
	}

	cfg.SetModuleEnabled("system", false)
	if cfg.Modules.System {
		t.Error("system should be disabled after SetModuleEnabled")
	}
}

func TestEnabledModuleNames(t *testing.T) {
	cfg := Defaults() // system and logo enabled
	names := cfg.EnabledModuleNames()

	if len(names) != 2 {
		t.Fatalf("expected 2 enabled modules, got %d: %v", len(names), names)
	}

	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	if !found["system"] || !found["logo"] {
		t.Errorf("expected system and logo, got %v", names)
	}
}
