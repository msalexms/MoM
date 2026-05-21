package module

import (
	"context"
	"testing"

	"github.com/msalexms/MoM/internal/config"
)

// mockModule is a test implementation of the Module interface.
type mockModule struct {
	name           string
	title          string
	description    string
	deps           []string
	available      bool
	output         string
	defaultEnabled bool
}

func (m *mockModule) Name() string           { return m.name }
func (m *mockModule) Title() string          { return m.title }
func (m *mockModule) Description() string    { return m.description }
func (m *mockModule) Dependencies() []string { return m.deps }
func (m *mockModule) Available() bool        { return m.available }
func (m *mockModule) DefaultEnabled() bool   { return m.defaultEnabled }
func (m *mockModule) Generate(ctx context.Context) (string, error) {
	return m.output, nil
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockModule{name: "test", title: "Test"})

	if r.Count() != 1 {
		t.Errorf("expected 1 module, got %d", r.Count())
	}
}

func TestRegistry_RegisterAll(t *testing.T) {
	r := NewRegistry()
	r.RegisterAll(
		&mockModule{name: "a", title: "A"},
		&mockModule{name: "b", title: "B"},
		&mockModule{name: "c", title: "C"},
	)

	if r.Count() != 3 {
		t.Errorf("expected 3 modules, got %d", r.Count())
	}
}

func TestRegistry_Get_Existing(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockModule{name: "weather", title: "Weather"})

	m, ok := r.Get("weather")
	if !ok {
		t.Fatal("expected module found")
	}
	if m.Name() != "weather" {
		t.Errorf("expected name 'weather', got %q", m.Name())
	}
}

func TestRegistry_Get_NonExistent(t *testing.T) {
	r := NewRegistry()

	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("expected module not found")
	}
}

func TestRegistry_All_Sorted(t *testing.T) {
	r := NewRegistry()
	r.RegisterAll(
		&mockModule{name: "zebra"},
		&mockModule{name: "alpha"},
		&mockModule{name: "middle"},
	)

	all := r.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 modules, got %d", len(all))
	}
	if all[0].Name() != "alpha" || all[1].Name() != "middle" || all[2].Name() != "zebra" {
		t.Errorf("expected sorted order, got: %s, %s, %s",
			all[0].Name(), all[1].Name(), all[2].Name())
	}
}

func TestRegistry_Available_Filter(t *testing.T) {
	r := NewRegistry()
	r.RegisterAll(
		&mockModule{name: "a", available: true},
		&mockModule{name: "b", available: false},
		&mockModule{name: "c", available: true},
	)

	avail := r.Available()
	if len(avail) != 2 {
		t.Errorf("expected 2 available modules, got %d", len(avail))
	}
}

func TestRegistry_Enabled_Filter(t *testing.T) {
	r := NewRegistry()
	r.RegisterAll(
		&mockModule{name: "system"},
		&mockModule{name: "logo"},
		&mockModule{name: "weather"},
	)

	cfg := config.Defaults() // system and logo enabled
	enabled := r.Enabled(cfg)
	if len(enabled) != 2 {
		t.Errorf("expected 2 enabled modules, got %d", len(enabled))
	}

	names := make(map[string]bool)
	for _, m := range enabled {
		names[m.Name()] = true
	}
	if !names["system"] || !names["logo"] {
		t.Errorf("expected system and logo enabled, got %v", names)
	}
}

func TestRegistry_Ordered(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockModule{name: "c"})
	r.Register(&mockModule{name: "a"})
	r.Register(&mockModule{name: "b"})

	ordered := r.Ordered()
	if ordered[0].Name() != "c" || ordered[1].Name() != "a" || ordered[2].Name() != "b" {
		t.Errorf("expected registration order c,a,b, got %s,%s,%s",
			ordered[0].Name(), ordered[1].Name(), ordered[2].Name())
	}
}

func TestCheckDependency(t *testing.T) {
	// "ls" should always be available on Linux
	if !CheckDependency("ls") {
		t.Error("expected 'ls' to be available")
	}

	// non-existent binary
	if CheckDependency("mom_nonexistent_binary_xyzzy") {
		t.Error("expected nonexistent binary to not be available")
	}
}
