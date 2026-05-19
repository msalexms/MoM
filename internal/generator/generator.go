// Package generator assembles the MOTD from enabled modules and writes it to the system.
package generator

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/ams/mom/internal/config"
	"github.com/ams/mom/internal/distro"
	"github.com/ams/mom/internal/module"
	"github.com/ams/mom/internal/module/render"
)

// Generator assembles MOTD content from enabled modules.
type Generator struct {
	Registry   *module.Registry
	Config     *config.Config
	Distro     distro.Info
	RenderOpts render.Options
}

// NewGenerator creates a new MOTD generator.
func NewGenerator(registry *module.Registry, cfg *config.Config, di distro.Info) *Generator {
	return &Generator{
		Registry:   registry,
		Config:     cfg,
		Distro:     di,
		RenderOpts: render.DefaultOptions(),
	}
}

// Generate produces the full MOTD output from all enabled modules.
func (g *Generator) Generate(ctx context.Context) (string, error) {
	enabledModules := g.Registry.Enabled(g.Config)
	return g.generateFromModules(ctx, enabledModules)
}

// GenerateLive produces MOTD output from only the specified module names.
// Useful for live preview in the TUI.
func (g *Generator) GenerateLive(ctx context.Context, moduleNames []string) (string, error) {
	var modules []module.Module
	for _, name := range moduleNames {
		if m, ok := g.Registry.Get(name); ok {
			modules = append(modules, m)
		}
	}
	return g.generateFromModules(ctx, modules)
}

func (g *Generator) generateFromModules(ctx context.Context, modules []module.Module) (string, error) {
	var parts []string

	// Header
	if g.Config.Motd.Header != "" {
		parts = append(parts, g.Config.Motd.Header)
	}

	// Generate each module with individual timeout
	for _, mod := range modules {
		modCtx, cancel := context.WithTimeout(ctx, 3*time.Second)

		var output string
		var err error

		// Prefer GenerateThemed if the module supports it
		if tm, ok := mod.(module.Themeable); ok {
			output, err = tm.GenerateThemed(modCtx, g.RenderOpts)
		} else {
			output, err = mod.Generate(modCtx)
		}
		cancel()

		if err != nil {
			slog.Error("module generation failed",
				"module", mod.Name(),
				"error", err)
			continue
		}
		if output == "" {
			continue
		}
		parts = append(parts, output)
	}

	if len(parts) == 0 && g.Config.Motd.Footer == "" {
		return "", nil
	}

	// Footer
	if g.Config.Motd.Footer != "" {
		parts = append(parts, g.Config.Motd.Footer)
	}

	return strings.Join(parts, "\n\n"), nil
}
