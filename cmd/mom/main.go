// Command mom configures the Linux MOTD from a terminal UI or headless flags.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/msalexms/MoM/internal/i18n"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/msalexms/MoM/internal/backup"
	"github.com/msalexms/MoM/internal/config"
	"github.com/msalexms/MoM/internal/distro"
	"github.com/msalexms/MoM/internal/generator"
	"github.com/msalexms/MoM/internal/module"
	"github.com/msalexms/MoM/internal/permission"
	"github.com/msalexms/MoM/internal/template"
	"github.com/msalexms/MoM/internal/theme"
	"github.com/msalexms/MoM/internal/tui"
	"github.com/msalexms/MoM/internal/version"
)

const usage = `mom · MOTD manager — [AMS] tech · amstech.dev

Usage:
  mom                          open the interface
  mom --print                  print the MOTD with the current config, without installing it
  mom --apply                  generate and install the MOTD with the current config
  mom --full-auto              enable every available module and install the MOTD
  mom --apply-template NAME    apply a built-in template and install the MOTD
  mom --import-template PATH   apply a template from a file and install the MOTD
  mom --export-template PATH   save the current config as a template
  mom --uninstall              restore the original MOTD
  mom --version                show the version

The language (English or Spanish) is chosen in the interface, under Appearance.

Options:
`

func main() {
	// The config is read first so flag help and errors use its language.
	cfg, cfgErr := config.Load()
	if cfgErr == nil {
		i18n.Set(i18n.Lang(cfg.Mode.Language))
	}

	var (
		flagVersion  = flag.Bool("version", false, i18n.T("show the version and exit"))
		flagPrint    = flag.Bool("print", false, i18n.T("print the MOTD without installing it"))
		flagMotd     = flag.Bool("motd", false, i18n.T("print the installed MOTD (used by the login hook)"))
		flagApply    = flag.Bool("apply", false, i18n.T("generate and install the MOTD with the current config"))
		flagFullAuto = flag.Bool("full-auto", false, i18n.T("enable every available module and install the MOTD"))
		flagApplyTpl = flag.String("apply-template", "", i18n.T("apply a built-in template by name"))
		flagExport   = flag.String("export-template", "", i18n.T("export the current config as a template"))
		flagImport   = flag.String("import-template", "", i18n.T("import and apply a template from a file"))
		flagUninst   = flag.Bool("uninstall", false, i18n.T("remove mom from login and restore the original MOTD"))
		flagRollback = flag.Bool("rollback", false, i18n.T("alias of --uninstall"))
	)
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, i18n.T(usage))
		flag.PrintDefaults()
	}
	flag.Parse()

	if *flagVersion {
		fmt.Printf(i18n.T("mom %s · [AMS] tech\ncommit %s · built %s\n"), version.Version, version.Commit, version.Date)
		return
	}

	if *flagMotd {
		runMotd()
		return
	}

	if cfgErr != nil {
		fatal(i18n.T("loading the config"), cfgErr)
	}

	if *flagExport != "" {
		if err := template.Export(cfg, *flagExport, "exported", "Exported from mom"); err != nil {
			fatal(i18n.T("exporting the template"), err)
		}
		fmt.Printf(i18n.T("Config exported to %s\n"), *flagExport)
		return
	}

	di, err := distro.Detect()
	if err != nil {
		fatal(i18n.T("detecting the distribution"), err)
	}
	if err := theme.LoadCustomThemes(); err != nil {
		fmt.Fprintf(os.Stderr, i18n.T("warning: could not load custom themes: %v\n"), err)
	}

	reg := buildRegistry(di)
	reg.SyncConfig(cfg)
	gen := generator.NewGenerator(reg, cfg)
	ctx := context.Background()

	if *flagPrint {
		out, err := gen.Generate(ctx)
		if err != nil {
			fatal(i18n.T("generating the MOTD"), err)
		}
		fmt.Println(out)
		return
	}

	bm := backup.NewManager(backup.DefaultDir())
	if err := bm.Init(generator.MotdPath); err != nil {
		fatal(i18n.T("preparing the backups"), err)
	}
	w := generator.NewWriter(bm, di.Family)

	switch {
	case *flagApply:
		apply(ctx, cfg, gen, w, i18n.T("MOTD installed."))
	case *flagFullAuto:
		for _, mod := range reg.Available() {
			cfg.SetModuleEnabled(mod.Name(), true)
		}
		apply(ctx, cfg, gen, w, fmt.Sprintf(i18n.T("MOTD installed with %d modules."), len(cfg.EnabledModuleNames())))
	case *flagApplyTpl != "":
		tmpl, err := template.GetBuiltin(*flagApplyTpl)
		if err != nil {
			names, _ := template.BuiltinNames()
			fatal(i18n.T("looking up the template"), fmt.Errorf(i18n.T("%w (available: %s)"), err, strings.Join(names, ", ")))
		}
		applyTemplate(ctx, tmpl, reg, cfg, gen, w)
	case *flagImport != "":
		tmpl, err := template.Import(*flagImport)
		if err != nil {
			fatal(i18n.T("importing the template"), err)
		}
		applyTemplate(ctx, tmpl, reg, cfg, gen, w)
	case *flagUninst, *flagRollback:
		authenticate(ctx)
		if err := w.Uninstall(ctx); err != nil {
			fatal(i18n.T("restoring the original MOTD"), err)
		}
		fmt.Println(i18n.T("Original MOTD restored."))
	default:
		runTUI(reg, cfg, gen, w, bm, di)
	}
}

// motdTimeout bounds the whole render at login, so a slow module (network,
// package manager) delays the login by at most this much and is skipped.
const motdTimeout = 3 * time.Second

// runMotd renders the installed MOTD. It runs at every login from the hook,
// so it never fails loudly: without an installed config it prints nothing.
func runMotd() {
	if _, err := os.Stat(config.SystemPath); err != nil {
		return
	}
	cfg, err := config.LoadFrom(config.SystemPath)
	if err != nil {
		return
	}
	i18n.Set(i18n.Lang(cfg.Mode.Language))
	di, _ := distro.Detect()
	reg := buildRegistry(di)
	reg.SyncConfig(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), motdTimeout)
	defer cancel()
	if out, _ := generator.NewGenerator(reg, cfg).Generate(ctx); out != "" {
		fmt.Println(out)
	}
}

// runTUI launches the interactive interface. It refuses to run as root:
// only the final file write needs privileges, and sudo is asked for then.
func runTUI(reg *module.Registry, cfg *config.Config, gen *generator.Generator, w *generator.Writer, bm *backup.Manager, di distro.Info) {
	if permission.IsRoot() {
		fmt.Fprintln(os.Stderr, i18n.T("mom: refusing to open the interface as root. Run it as your user;"))
		fmt.Fprintln(os.Stderr, i18n.T("     sudo is asked for only when installing the MOTD."))
		os.Exit(1)
	}
	p := tea.NewProgram(tui.NewModel(reg, cfg, gen, w, bm, di, func() *module.Registry { return buildRegistry(di) }), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fatal(i18n.T("running the interface"), err)
	}
}

func applyTemplate(ctx context.Context, tmpl *template.Template, reg *module.Registry, cfg *config.Config, gen *generator.Generator, w *generator.Writer) {
	template.Apply(tmpl, cfg)
	reg.SyncConfig(cfg)
	apply(ctx, cfg, gen, w, fmt.Sprintf(i18n.T("Template “%s” applied."), tmpl.Name))
}

func apply(ctx context.Context, cfg *config.Config, gen *generator.Generator, w *generator.Writer, done string) {
	authenticate(ctx)
	if err := generator.Apply(ctx, cfg, gen, w); err != nil {
		fatal(i18n.T("applying the MOTD"), err)
	}
	fmt.Println(done)
}

func authenticate(ctx context.Context) {
	if err := permission.Authenticate(ctx); err != nil {
		fatal(i18n.T("asking for permissions"), err)
	}
}

// buildRegistry registers every module, in the order the TUI lists them.
func buildRegistry(di distro.Info) *module.Registry {
	reg := module.NewRegistry()
	reg.RegisterAll(
		&module.LogoModule{Distro: di},
		&module.SystemModule{},
		&module.ResourcesModule{},
		&module.NetworkModule{},
		&module.WeatherModule{},
		&module.ContainersModule{},
		&module.ServicesModule{},
		&module.UpdatesModule{Distro: di.Family},
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
		&module.GitStatusModule{},
		&module.TmuxModule{},
		&module.RebootModule{},
		&module.CalendarModule{},
		&module.QuoteModule{},
		&module.CowsayModule{},
	)
	return reg
}

func fatal(what string, err error) {
	fmt.Fprintf(os.Stderr, i18n.T("mom: error %s: %v\n"), what, err)
	os.Exit(1)
}
