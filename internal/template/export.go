package template

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/ams/mom/internal/config"
)

// Export saves the current configuration as a shareable template file.
func Export(cfg *config.Config, path string, name string, description string) error {
	t := &Template{
		Name:        name,
		Description: description,
		Author:      fmt.Sprintf("mom v0.1.0 (exported %s)", time.Now().Format("2006-01-02")),
		Motd:        cfg.Motd,
		Modules: moduleMap{
			System:     cfg.Modules.System,
			Resources:  cfg.Modules.Resources,
			Weather:    cfg.Modules.Weather,
			Cowsay:     cfg.Modules.Cowsay,
			Network:    cfg.Modules.Network,
			Containers: cfg.Modules.Containers,
			Updates:    cfg.Modules.Updates,
			Logins:     cfg.Modules.Logins,
			Quote:      cfg.Modules.Quote,
			Calendar:   cfg.Modules.Calendar,
			Services:   cfg.Modules.Services,
			Logo:       cfg.Modules.Logo,
		},
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
