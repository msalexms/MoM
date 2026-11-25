# Changelog

All notable changes to **mom** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and the project follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.2.0] - 2026-11-25

### Added

- **Bilingual interface** — the TUI, its help text, and the generated MOTD are
  available in English and Spanish, switchable live from **Appearance → Language**.
- **AMSTech theme** (`amstech`) — the brand palette built from the AMSTech design
  tokens (brand indigo, Thor, signal, and series colors). It ships as the seventh
  built-in theme and the default for new configurations.
- **AMSTech brand components** — Thor mark, wordmark, and a grouped home screen in
  the terminal UI, plus banners, seals, and the Thor Terminal sticker under
  `assets/brand/`.
- **`mom --version`** — prints the release version, commit, and build date injected
  at build time through `internal/version`.
- **Spanish documentation** — `README.es.md` and the Spanish banner, with the same
  structure as the English README.
- **Screen recordings** — a GIF tour for modules, module settings, appearance,
  templates and profiles, preview, install, and backups.

### Changed

- **Home screen** — reorganised into *Configure*, *Start from something*, and *MOTD*
  groups, with one-key shortcuts for every screen.
- **TUI structure** — screens consolidated into focused files (home, modules,
  settings, backups, misc) and the brand kit extracted into shared components.
- **Backups** — snapshot creation, listing, pruning, and restore consolidated into
  the backup manager and a single `generator.Writer` with its own tests.
- **Backup format** — snapshots now store the TOML configuration instead of the
  rendered text, so a saved setup can be previewed and installed again. Snapshots
  created by 0.1.x are still listed and previewable but are flagged as legacy and
  cannot be reinstalled; the original `/etc/motd` is never touched.
- **README** — rewritten and restyled to the AMSTech brand: banner, badges, feature
  tour, full module/theme/template tables, and a bilingual language switch.

### Removed

- Legacy dashboard, rollback, help, module-help, and theme views, replaced by the
  new screens.
- `scripts/build.sh` and `scripts/smoke-test.sh`, superseded by the `Makefile`.

### Fixed

- The release badge and repository links now point at `msalexms/MoM`.

## [0.1.1] - 2026-05-21

### Changed

- Renamed the Go module to `github.com/msalexms/MoM` and updated every import.
- Updated the nfpm metadata with the correct author and repository information.

## [0.1.0] - 2026-05-21

### Added

- Initial public release: interactive TUI, 32 auto-detecting modules, 6 themes,
  6 section styles, 10 templates, saved profiles, live preview, backups, and
  distro-aware login hooks for Debian/Ubuntu (`pam_motd`) and other families
  (`profile.d`).
- Headless flags: `--print`, `--apply`, `--full-auto`, `--apply-template`,
  `--export-template`, `--import-template`, `--uninstall`, and `--motd`.
- CI and GoReleaser release workflows producing amd64/arm64/armv7 archives and
  `.deb`/`.rpm` packages.

[Unreleased]: https://github.com/msalexms/MoM/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/msalexms/MoM/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/msalexms/MoM/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/msalexms/MoM/releases/tag/v0.1.0
