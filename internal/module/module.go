// Package module defines the Module interface and provides a Registry
// for discovering and managing MOTD modules.
package module

import (
	"context"
	"os/exec"
	"sort"

	"github.com/ams/mom/internal/config"
)

// Module defines the interface that all MOTD modules must implement.
type Module interface {
	// Name returns the unique module identifier (e.g. "system", "weather").
	Name() string

	// Title returns the human-readable display name (e.g. "System Information").
	Title() string

	// Description returns a one-line description for the TUI.
	Description() string

	// Dependencies returns external binaries required (e.g. ["cowsay"]).
	// Empty slice means no external dependencies.
	Dependencies() []string

	// Available returns true if all dependencies are found in PATH.
	Available() bool

	// Generate produces the MOTD output for this module.
	// Returns empty string if the module is unavailable or encounters an error.
	Generate(ctx context.Context) (string, error)

	// DefaultEnabled returns whether this module should be on by default
	// in auto-detection and full-auto modes.
	DefaultEnabled() bool
}

// Registry holds registered modules and provides lookup/filter operations.
type Registry struct {
	modules map[string]Module
	order   []string // preserve registration order
}

// NewRegistry creates a new empty module registry.
func NewRegistry() *Registry {
	return &Registry{
		modules: make(map[string]Module),
	}
}

// Register adds a module to the registry.
func (r *Registry) Register(m Module) {
	name := m.Name()
	if _, exists := r.modules[name]; !exists {
		r.order = append(r.order, name)
	}
	r.modules[name] = m
}

// RegisterAll adds multiple modules to the registry.
func (r *Registry) RegisterAll(modules ...Module) {
	for _, m := range modules {
		r.Register(m)
	}
}

// Get returns a module by name, or (nil, false) if not found.
func (r *Registry) Get(name string) (Module, bool) {
	m, ok := r.modules[name]
	return m, ok
}

// All returns all registered modules sorted alphabetically by name.
func (r *Registry) All() []Module {
	result := make([]Module, 0, len(r.modules))
	for _, m := range r.modules {
		result = append(result, m)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name() < result[j].Name()
	})
	return result
}

// Ordered returns all registered modules in their registration order.
func (r *Registry) Ordered() []Module {
	result := make([]Module, 0, len(r.order))
	for _, name := range r.order {
		if m, ok := r.modules[name]; ok {
			result = append(result, m)
		}
	}
	return result
}

// Available returns modules whose dependencies are all satisfied.
func (r *Registry) Available() []Module {
	var result []Module
	for _, m := range r.All() {
		if m.Available() {
			result = append(result, m)
		}
	}
	return result
}

// Enabled returns modules that are enabled in the given config.
func (r *Registry) Enabled(cfg *config.Config) []Module {
	var result []Module
	for _, name := range cfg.EnabledModuleNames() {
		if m, ok := r.modules[name]; ok {
			result = append(result, m)
		}
	}
	return result
}

// Count returns the total number of registered modules.
func (r *Registry) Count() int {
	return len(r.modules)
}

// CheckDependency returns true if the given binary is found in PATH.
func CheckDependency(binary string) bool {
	_, err := exec.LookPath(binary)
	return err == nil
}
