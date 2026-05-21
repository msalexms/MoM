package template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/msalexms/MoM/internal/config"
)

func TestBuiltinTemplates(t *testing.T) {
	templates, err := BuiltinTemplates()
	if err != nil {
		t.Fatalf("BuiltinTemplates failed: %v", err)
	}
	if len(templates) != 10 {
		t.Errorf("expected 10 built-in templates, got %d", len(templates))
	}

	names := make(map[string]bool)
	for _, tmpl := range templates {
		names[tmpl.Name] = true
		if tmpl.Description == "" {
			t.Errorf("template %q has empty description", tmpl.Name)
		}
	}

	expected := []string{"Minimal", "Sysadmin", "Developer", "Hacker", "Full", "Gaming", "Server", "Homelab", "Security", "Aesthetic"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected built-in template %q not found", name)
		}
	}
}

func TestApply_DoesNotModifyPerModuleConfig(t *testing.T) {
	cfg := config.Defaults()
	cfg.Modules.WeatherConfig.City = "Madrid"
	cfg.Modules.WeatherConfig.Units = "imperial"
	cfg.Modules.ContainersConfig.Runtime = "podman"

	tmpl := &Template{
		Name: "test",
		Modules: moduleMap{
			System:  true,
			Weather: true,
			Logo:    true,
		},
	}

	Apply(tmpl, cfg)

	// Module enabled states should change
	if !cfg.Modules.Weather {
		t.Error("expected weather enabled after Apply")
	}

	// Per-module config should NOT change
	if cfg.Modules.WeatherConfig.City != "Madrid" {
		t.Errorf("weather city changed to %q", cfg.Modules.WeatherConfig.City)
	}
	if cfg.Modules.WeatherConfig.Units != "imperial" {
		t.Errorf("weather units changed to %q", cfg.Modules.WeatherConfig.Units)
	}
	if cfg.Modules.ContainersConfig.Runtime != "podman" {
		t.Errorf("containers runtime changed to %q", cfg.Modules.ContainersConfig.Runtime)
	}
}

func TestExportImport_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-template.toml")

	cfg := config.Defaults()
	cfg.Motd.Header = "Test Header"
	cfg.Modules.Weather = true
	cfg.Modules.Resources = true

	err := Export(cfg, path, "MyTemplate", "A test template")
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("exported file not found: %v", err)
	}

	// Import it back
	imported, err := Import(path)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if imported.Name != "MyTemplate" {
		t.Errorf("expected name 'MyTemplate', got %q", imported.Name)
	}
	if imported.Description != "A test template" {
		t.Errorf("expected description 'A test template', got %q", imported.Description)
	}
	if imported.Motd.Header != "Test Header" {
		t.Errorf("expected header 'Test Header', got %q", imported.Motd.Header)
	}
	if !imported.Modules.Weather {
		t.Error("expected weather enabled in imported template")
	}
	if !imported.Modules.Resources {
		t.Error("expected resources enabled in imported template")
	}
}

func TestImport_MissingName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad-template.toml")
	os.WriteFile(path, []byte(`description = "no name field"`), 0644)

	_, err := Import(path)
	if err == nil {
		t.Error("expected error for template without name")
	}
}
