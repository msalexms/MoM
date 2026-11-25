package template

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/msalexms/MoM/internal/config"
	"github.com/msalexms/MoM/internal/version"
)

// Export saves the current configuration as a shareable template file.
// Every known module is written, so an export/import round-trip preserves
// the full module selection.
func Export(cfg *config.Config, path string, name string, description string) error {
	modules := make(map[string]bool, len(config.KnownModuleNames()))
	for _, moduleName := range config.KnownModuleNames() {
		modules[moduleName] = cfg.IsModuleEnabled(moduleName)
	}

	t := &Template{
		Name:        name,
		Description: description,
		Author:      fmt.Sprintf("mom %s · AMSTech (exported %s)", version.Version, time.Now().Format("2006-01-02")),
		Motd:        cfg.Motd,
		Modules:     modules,
	}

	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(t); err != nil {
		return fmt.Errorf("encoding template: %w", err)
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}

// Import reads a template from a TOML file.
func Import(path string) (*Template, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading template file: %w", err)
	}

	var t Template
	if _, err := toml.Decode(string(data), &t); err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	if t.Name == "" {
		return nil, fmt.Errorf("template missing name field")
	}

	return &t, nil
}
