package generator

import (
	"context"
	"strings"
	"testing"

	"github.com/msalexms/MoM/internal/config"
	"github.com/msalexms/MoM/internal/module"
	"github.com/msalexms/MoM/internal/module/render"
)

// testModule is a mock module for testing.
type testModule struct {
	name   string
	output string
	avail  bool
}

func (m *testModule) Name() string           { return m.name }
func (m *testModule) Title() string          { return m.name }
func (m *testModule) Description() string    { return "test" }
func (m *testModule) Dependencies() []string { return nil }
func (m *testModule) Available() bool        { return m.avail }
func (m *testModule) Generate(_ context.Context, _ render.Options) (string, error) {
	return m.output, nil
}

func TestGenerator_Generate_EnabledModules(t *testing.T) {
	reg := module.NewRegistry()
	reg.RegisterAll(
		&testModule{name: "system", output: "SYSTEM INFO", avail: true},
		&testModule{name: "logo", output: "LOGO ART", avail: true},
		&testModule{name: "weather", output: "WEATHER", avail: true},
	)

	cfg := config.Defaults() // system + logo enabled
	gen := NewGenerator(reg, cfg)

	result, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result == "" {
		t.Fatal("expected non-empty output")
	}
	if !strings.Contains(result, "SYSTEM INFO") {
		t.Error("expected system output in result")
	}
	if !strings.Contains(result, "LOGO ART") {
		t.Error("expected logo output in result")
	}
	if strings.Contains(result, "WEATHER") {
		t.Error("weather should not be in output (disabled)")
	}
}

func TestGenerator_Generate_HeaderFooter(t *testing.T) {
	reg := module.NewRegistry()
	reg.Register(&testModule{name: "system", output: "SYS", avail: true})

	cfg := config.Defaults()
	cfg.Motd.Header = "=== HEADER ==="
	cfg.Motd.Footer = "=== FOOTER ==="

	gen := NewGenerator(reg, cfg)

	result, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(result, "=== HEADER ===") {
		t.Error("expected header in output")
	}
	if !strings.Contains(result, "=== FOOTER ===") {
		t.Error("expected footer in output")
	}
}

func TestGenerator_Generate_EmptyWhenNoModules(t *testing.T) {
	reg := module.NewRegistry()
	reg.Register(&testModule{name: "weather", output: "W", avail: true})

	cfg := config.Defaults() // system and logo enabled but not registered
	cfg.Modules.System = false
	cfg.Modules.Logo = false

	gen := NewGenerator(reg, cfg)

	result, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty output, got %q", result)
	}
}

func TestGenerate_PlainThemeHasNoEscapes(t *testing.T) {
	reg := module.NewRegistry()
	reg.Register(&testModule{name: "system", output: "\033[1mSYS\033[0m", avail: true})
	cfg := config.Defaults()
	cfg.Modules.Logo = false
	cfg.Mode.Theme = "ascii"

	out, err := NewGenerator(reg, cfg).Generate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "\033") || out != "SYS" {
		t.Errorf("ascii theme output = %q", out)
	}
}
