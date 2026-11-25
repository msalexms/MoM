// Package generator assembles the MOTD from the enabled modules and installs it.
package generator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/msalexms/MoM/internal/i18n"

	"github.com/msalexms/MoM/internal/config"
	"github.com/msalexms/MoM/internal/module"
	"github.com/msalexms/MoM/internal/module/render"
	"github.com/msalexms/MoM/internal/theme"
)

// moduleTimeout bounds each module so one slow dependency (a network call,
// a hung package manager) can never stall the whole MOTD.
const moduleTimeout = 5 * time.Second

// Generator assembles MOTD content from the modules enabled in the config.
// It keeps no render state: theme and variant are read from the config on
// every call, so profile loads and template applies take effect at once.
type Generator struct {
	registry *module.Registry
	config   *config.Config
}

// NewGenerator creates a MOTD generator bound to a registry and a config.
func NewGenerator(registry *module.Registry, cfg *config.Config) *Generator {
	return &Generator{registry: registry, config: cfg}
}

// Generate renders every enabled module and joins the sections in the
// configured order. Modules run concurrently, so the total time is that of
// the slowest module rather than the sum of all of them. A module that
// fails or runs out of time is skipped; it never aborts the MOTD.
func (g *Generator) Generate(ctx context.Context) (string, error) {
	opts := render.Options{
		Theme:   theme.MustGet(g.config.ThemeID()),
		Variant: render.Variant(g.config.GlobalVariant()),
	}

	mods := g.registry.Enabled(g.config)
	sections := make([]string, len(mods))

	var wg sync.WaitGroup
	for i, mod := range mods {
		wg.Add(1)
		go func() {
			defer wg.Done()
			modCtx, cancel := context.WithTimeout(ctx, moduleTimeout)
			defer cancel()

			out, err := mod.Generate(modCtx, opts)
			if err != nil {
				slog.Debug("module generation failed", "module", mod.Name(), "error", err)
				return
			}
			if opts.Theme.Plain() {
				out = render.StripANSI(out)
			}
			sections[i] = out
		}()
	}
	wg.Wait()

	var parts []string
	for _, s := range sections {
		if s != "" {
			parts = append(parts, s)
		}
	}
	if len(parts) == 0 {
		return "", nil
	}

	if h := g.config.Motd.Header; h != "" {
		parts = append([]string{h}, parts...)
	}
	if f := g.config.Motd.Footer; f != "" {
		parts = append(parts, f)
	}
	return strings.Join(parts, "\n\n"), nil
}

// Apply is the single check → save → install pipeline shared by the TUI
// and every headless flag. The MOTD is rendered once to make sure it is not
// empty; what gets installed is the config, which the login hook renders
// again at every login with fresh data.
func Apply(ctx context.Context, cfg *config.Config, gen *Generator, w *Writer) error {
	content, err := gen.Generate(ctx)
	if err != nil {
		return fmt.Errorf(i18n.T("generating the MOTD: %w"), err)
	}
	if content == "" {
		return errors.New(i18n.T("no enabled module produced any output; nothing to install"))
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf(i18n.T("saving the config: %w"), err)
	}

	if err := w.Install(ctx, cfg); err != nil {
		return fmt.Errorf(i18n.T("installing the MOTD: %w"), err)
	}
	return nil
}
