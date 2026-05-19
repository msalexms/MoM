package generator

import (
	"context"
	"testing"

	"github.com/ams/mom/internal/config"
	"github.com/ams/mom/internal/distro"
	"github.com/ams/mom/internal/module"
)

// testModule is a mock module for testing.
type testModule struct {
	name    string
	output  string
	avail   bool
	defOn   bool
}

func (m *testModule) Name() string                          { return m.name }
func (m *testModule) Title() string                         { return m.name }
func (m *testModule) Description() string                   { return "test" }
func (m *testModule) Dependencies() []string                { return nil }
func (m *testModule) Available() bool                       { return m.avail }
func (m *testModule) DefaultEnabled() bool                  { return m.defOn }
func (m *testModule) Generate(ctx context.Context) (string, error) { return m.output, nil }

func TestGenerator_Generate_EnabledModules(t *testing.T) {
	reg := module.NewRegistry()
	reg.RegisterAll(
		&testModule{name: "system", output: "SYSTEM INFO", avail: true, defOn: true},
		&testModule{name: "logo", output: "LOGO ART", avail: true, defOn: true},
		&testModule{name: "weather", output: "WEATHER", avail: true, defOn: false},
	)

	cfg := config.Defaults() // system + logo enabled
	gen := NewGenerator(reg, cfg, distro.Info{Family: distro.FamilyDebian})

	result, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result == "" {
		t.Fatal("expected non-empty output")
	}
	if !contains(result, "SYSTEM INFO") {
		t.Error("expected system output in result")
	}
	if !contains(result, "LOGO ART") {
		t.Error("expected logo output in result")
	}
	if contains(result, "WEATHER") {
		t.Error("weather should not be in output (disabled)")
	}
}

func TestGenerator_Generate_HeaderFooter(t *testing.T) {
	reg := module.NewRegistry()
	reg.Register(&testModule{name: "system", output: "SYS", avail: true})

	cfg := config.Defaults()
	cfg.Motd.Header = "=== HEADER ==="
	cfg.Motd.Footer = "=== FOOTER ==="

	gen := NewGenerator(reg, cfg, distro.Info{})

	result, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !contains(result, "=== HEADER ===") {
		t.Error("expected header in output")
	}
	if !contains(result, "=== FOOTER ===") {
		t.Error("expected footer in output")
	}
}

func TestGenerator_Generate_EmptyWhenNoModules(t *testing.T) {
	reg := module.NewRegistry()
	reg.Register(&testModule{name: "weather", output: "W", avail: true})

	cfg := config.Defaults() // system and logo enabled but not registered
	cfg.Modules.System = false
	cfg.Modules.Logo = false

	gen := NewGenerator(reg, cfg, distro.Info{})

	result, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty output, got %q", result)
	}
}

func TestGenerator_GenerateLive(t *testing.T) {
	reg := module.NewRegistry()
	reg.RegisterAll(
		&testModule{name: "system", output: "SYS", avail: true},
		&testModule{name: "logo", output: "LOGO", avail: true},
		&testModule{name: "weather", output: "WEATHER", avail: true},
	)

	cfg := config.Defaults()
	gen := NewGenerator(reg, cfg, distro.Info{})

	// Only generate weather and logo
	result, err := gen.GenerateLive(context.Background(), []string{"weather", "logo"})
	if err != nil {
		t.Fatalf("GenerateLive failed: %v", err)
	}

	if !contains(result, "WEATHER") {
		t.Error("expected weather in result")
	}
	if !contains(result, "LOGO") {
		t.Error("expected logo in result")
	}
	if contains(result, "SYS") {
		t.Error("system should not be in live result")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
