# AGENTS.md — mom

> Compact operational guide for OpenCode sessions working on `mom`.

## Project at a glance

- **What**: Interactive TUI (Bubble Tea) for managing the Linux MOTD.
- **Module**: `github.com/ams/mom`
- **Go version**: 1.25.0
- **Build target**: Linux only (amd64, arm64, armv7).
- **Entrypoint**: `cmd/mom/main.go`
- **Config path**: `~/.config/mom/config.toml`

## Exact developer commands

```bash
make build        # linux/amd64 binary → bin/mom
make build-all    # amd64 + arm64 + armv7
make test         # go test ./... -v -race -count=1
make lint         # go vet ./... + staticcheck ./... (optional)
make fmt          # go fmt ./...
make run          # go run ./cmd/mom
make clean        # rm -rf bin/
make release      # goreleaser release --clean
```

Run `scripts/smoke-test.sh` for a quick vet → build → test → binary-version check.
Run `scripts/build.sh` for multi-arch local builds with checksums.

## Testing

- Uses standard `testing` only — no testify/assert.
- Race detector is enabled (`-race`).
- Tests rely on testable constructors like `DetectFrom(path)` and `LoadFrom(path)` so they can point at temp files. Prefer these when writing new tests instead of mocking globals.
- Run a single package: `go test -v ./internal/distro/`

## Architecture & boundaries

```
cmd/mom/main.go           # CLI flag parsing, TUI init, headless modes
internal/distro/          # /etc/os-release parsing, MOTD path resolution per family
internal/config/          # TOML load/save/defaults; uses BurntSushi/toml
internal/module/          # 12 modules implementing the Module interface
internal/module/render/   # Renderer (theme-aware), Options, Variant, legacy compat layer
internal/generator/       # Assembles MOTD from enabled modules; writes to system paths
internal/template/        # 5 built-in templates (embed/templates/*.toml) + export/import
internal/backup/          # Snapshots, immutable original, rollback
internal/permission/      # Punctual sudo elevation (never run TUI as root)
internal/theme/           # 6 built-in themes (default, dracula, nord, solarized-dark, monochrome, ascii)
internal/tui/             # Bubble Tea app (router) + view files
internal/tui/components/  # Shared lipgloss styles
internal/tui/keys/        # KeyMap definitions
embed/                    # Logos and template TOML files embedded with //go:embed
```

### Module system

- **Module interface** (`internal/module/module.go`) is the base contract: `Name/Title/Description/Dependencies/Available/Generate/DefaultEnabled`.
- **Configurable** (optional): `Variants() []Variant`, `DefaultVariant()`, `Settings() []SettingDef` — exposes per-module settings to the TUI.
- **Themeable** (optional): `GenerateThemed(ctx, render.Options)` — accepts theme + variant at generation time. Generator uses this via type assertion.
- **Registry** preserves insertion order via `Ordered()`; `All()` returns alphabetically sorted.
- **Generator** gives each module a 3-second timeout per `Generate` call. Uses `GenerateThemed` when available.

### Theme system

- `internal/theme/theme.go` — `Theme` struct with semantic `Palette` (Accent, Success, Warning, Danger, per-section colors, gradient stops) and `Attrs` (Bold, Dim, Italic).
- `internal/theme/themes.go` — 6 built-in themes registered via `init()`.
- Themes are selected in config (`mode.theme`) and applied via `render.Options` passed to modules.
- `render.Renderer` wraps a theme and provides `Header/KeyValue/ProgressBar/StatusDot/AsciiBanner`.

### Render variants

Modules declare supported variants (`default`, `compact`, `detailed`, `minimal`, `ascii`, `boxed`). The global variant is set in `mode.variant` in config. Per-module override is possible via the TUI.

### Services

- `ListSystemServices(ctx)` enumerates real systemd services via `systemctl list-unit-files`.
- User selects services in the TUI Services Picker (with incremental filter).
- Selection persisted in `modules.services_config.services []string`.

### TUI structure

- `app.go` — Model, Init, Update (router), View (dispatcher), shared styles, actions.
- `view_dashboard.go` — Main menu with 11 items.
- `view_modules.go` — Module toggle list.
- `view_templates.go` — Template selector.
- `view_preview.go` — MOTD preview.
- `view_rollback.go` — Backup restore.
- `view_help.go` — Keyboard shortcuts.
- `view_asciiart_services.go` — ASCII art input + Services picker with filter.
- `view_theme.go` — Theme selector.

## Security constraints (hard rules)

1. **Never run the TUI as root.** Sudo elevation is punctual — only the `cp`/`mkdir`/`chmod` operation runs via `sudo`.
2. **Always backup before writing.** `Writer.Write` calls `BackupManager.Backup` first.
3. **Original MOTD backup is immutable.** Once `Init` saves it, it must never be overwritten.
4. **No `os.Exit` outside `main()`** and no `panic` in business logic.

## Style & conventions

- `go fmt`, `go vet`, Effective Go. No additional linters are enforced in CI.
- Errors: return `error` as last value; wrap with `fmt.Errorf("...: %w", err)`.
- Logging: `log/slog` (stdlib). Levels: Debug (TUI), Info (operations), Error (failures).
- Comments: GoDoc for exported symbols; explain *why*, not *what*.
- Context: I/O operations accept `context.Context` as first param.
- Timeouts: network 5s (wttr.in), file ops 2s, per-module generate 3s.

## Build & release

- `Makefile` injects version/commit/date via `-ldflags` into `main.version`, `main.commit`, `main.date`.
- `.goreleaser.yaml` builds linux/amd64, arm64, arm(v7), outputs `tar.gz` + `deb` + `rpm`.
- `bin/` and `dist/` are gitignored.

## Notes

- `PLAN.md` is a detailed design spec / roadmap. It describes intended behavior and the 12-module set. Treat it as authoritative for architecture decisions, but verify against the actual code before assuming a feature is already implemented.
- There are no CI workflows, pre-commit hooks, or automated checks beyond the Makefile and smoke-test script.
