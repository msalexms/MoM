package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ams/mom/internal/backup"
	"github.com/ams/mom/internal/config"
	"github.com/ams/mom/internal/distro"
	"github.com/ams/mom/internal/generator"
	"github.com/ams/mom/internal/module"
	"github.com/ams/mom/internal/template"
	"github.com/ams/mom/internal/theme"
	"github.com/ams/mom/internal/tui"
)

var (
	version = "0.1.0"
	commit  = "dev"
	date    = "unknown"
)

func main() {
	// CLI flags
	flagVersion := flag.Bool("version", false, "Print version and exit")
	flagFullAuto := flag.Bool("full-auto", false, "Run full-auto setup without TUI")
	flagApplyTemplate := flag.String("apply-template", "", "Apply a built-in template by name")
	flagExportTemplate := flag.String("export-template", "", "Export current config as template to file")
	flagImportTemplate := flag.String("import-template", "", "Import and apply a template from file")
	flagRollback := flag.Bool("rollback", false, "Restore the original MOTD backup")
	flagUninstall := flag.Bool("uninstall", false, "Remove mom's changes and restore original MOTD")
	flag.Parse()

	if *flagVersion {
		fmt.Printf("mom v%s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Initialize core components
	ctx := context.Background()

	// Detect distro
	di, err := distro.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting distro: %v\n", err)
		os.Exit(1)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize backup manager
	bm := backup.NewManager(backup.DefaultBackupDir())
	paths := distro.GetPaths(di.Family)
	if err := bm.Init(ctx, paths.MotdFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing backups: %v\n", err)
		os.Exit(1)
	}

	// Build module registry
	reg := buildRegistry(di, cfg)

	// Load custom themes
	theme.LoadCustomThemes()

	// Generator and writer
	gen := generator.NewGenerator(reg, cfg, di)
	w := generator.NewWriter(bm, paths, di)

	// Handle CLI flags (headless mode)
	if *flagFullAuto {
		runFullAuto(ctx, reg, cfg, gen, w)
		return
	}
	if *flagApplyTemplate != "" {
		runApplyTemplate(ctx, *flagApplyTemplate, cfg, gen, w)
		return
	}
	if *flagExportTemplate != "" {
		runExport(cfg, *flagExportTemplate)
		return
	}
	if *flagImportTemplate != "" {
		runImportTemplate(ctx, *flagImportTemplate, cfg, gen, w)
		return
	}
	if *flagRollback {
		runRollback(ctx, bm, paths)
		return
	}
	if *flagUninstall {
		runUninstall(ctx, w)
		return
	}

	// Start TUI
	model := tui.NewModel(reg, cfg, gen, w, bm, di)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

func buildRegistry(di distro.Info, cfg *config.Config) *module.Registry {
	reg := module.NewRegistry()
	reg.RegisterAll(
		&module.LogoModule{Distro: di},
		&module.SystemModule{},
		&module.ResourcesModule{ShowTemp: cfg.Modules.ResourcesConfig.ShowTemp},
		&module.NetworkModule{},
		&module.WeatherModule{
			City:  cfg.Modules.WeatherConfig.City,
			Units: cfg.Modules.WeatherConfig.Units,
		},
		&module.ContainersModule{Runtime: cfg.Modules.ContainersConfig.Runtime},
		&module.ServicesModule{Services: cfg.Modules.ServicesConfig.Services},
		&module.UpdatesModule{Distro: di.Family, IncludeAUR: cfg.Modules.UpdatesConfig.IncludeAUR},
		&module.LoginsModule{},
		&module.PortsModule{},
		&module.TopProcsModule{},
		&module.Fail2banModule{},
		&module.BatteryModule{},
		&module.UsersModule{},
		&module.SSHKeysModule{},
		&module.LastBootModule{},
		&module.GPUModule{},
		&module.NetTrafficModule{},
		&module.DiskIOModule{},
		&module.TimersModule{},
		&module.JournalModule{},
		&module.ZFSModule{},
		&module.CertsModule{},
		&module.FirewallModule{},
		&module.FailedLoginsModule{},
		&module.SudoModule{},
		&module.GitStatusModule{
			Paths:    cfg.Modules.GitConfig.Paths,
			MaxRepos: cfg.Modules.GitConfig.MaxRepos,
		},
		&module.TmuxModule{},
		&module.RebootModule{},
		&module.CalendarModule{},
		&module.QuoteModule{},
		&module.CowsayModule{
			Mode:    cfg.Modules.CowsayConfig.Mode,
			Message: cfg.Modules.CowsayConfig.Message,
		},
	)
	return reg
}

func runFullAuto(ctx context.Context, reg *module.Registry, cfg *config.Config, gen *generator.Generator, w *generator.Writer) {
	// Enable all available modules
	for _, mod := range reg.Available() {
		cfg.SetModuleEnabled(mod.Name(), true)
	}

	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	content, err := gen.Generate(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating MOTD: %v\n", err)
		os.Exit(1)
	}

	modules := cfg.EnabledModuleNames()
	if err := w.Write(ctx, content, modules); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing MOTD: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Full-auto setup complete! MOTD has been updated.")
}

func runApplyTemplate(ctx context.Context, name string, cfg *config.Config, gen *generator.Generator, w *generator.Writer) {
	tmpl, err := template.GetBuiltin(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Available templates: Minimal, Sysadmin, Developer, Hacker, Full\n")
		os.Exit(1)
	}

	template.Apply(tmpl, cfg)
	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	content, err := gen.Generate(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating MOTD: %v\n", err)
		os.Exit(1)
	}

	modules := cfg.EnabledModuleNames()
	if err := w.Write(ctx, content, modules); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing MOTD: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Template '%s' applied successfully!\n", tmpl.Name)
}

func runExport(cfg *config.Config, path string) {
	if err := template.Export(cfg, path, "exported", "Exported from mom"); err != nil {
		fmt.Fprintf(os.Stderr, "Error exporting template: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Config exported to %s\n", path)
}

func runImportTemplate(ctx context.Context, path string, cfg *config.Config, gen *generator.Generator, w *generator.Writer) {
	tmpl, err := template.Import(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error importing template: %v\n", err)
		os.Exit(1)
	}

	template.Apply(tmpl, cfg)
	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	content, err := gen.Generate(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating MOTD: %v\n", err)
		os.Exit(1)
	}

	modules := cfg.EnabledModuleNames()
	if err := w.Write(ctx, content, modules); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing MOTD: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Template '%s' imported and applied!\n", tmpl.Name)
}

func runRollback(ctx context.Context, bm *backup.Manager, paths distro.MotdPaths) {
	if err := bm.RollbackToOriginal(ctx, paths.MotdFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error rolling back: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("MOTD rolled back to original!")
}

func runUninstall(ctx context.Context, w *generator.Writer) {
	if err := w.Remove(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error uninstalling: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("mom uninstalled. Original MOTD restored.")
}
