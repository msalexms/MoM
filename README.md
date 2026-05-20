<div align="center">

# `mom` — Message Of the Day Manager

[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/ams/mom?style=flat-square)](https://github.com/ams/mom/releases)
[![Go Version](https://img.shields.io/badge/go-1.25.0-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![Platform](https://img.shields.io/badge/platform-Linux-333333?style=flat-square&logo=linux)](https://github.com/ams/mom)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE)

An interactive, themeable TUI for crafting beautiful Linux MOTDs. No static text files — just live system data, smart defaults, and a gorgeous terminal interface.

[Installation](#installation) · [Usage](#usage) · [Modules](#modules) · [Themes](#themes) · [Templates](#templates)

</div>

---

## Preview

> ![Dashboard Placeholder](https://via.placeholder.com/800x450/0d1117/58a6ff?text=Dashboard+View)
> *The main dashboard — navigate between modules, themes, templates and previews.*

> ![Preview Placeholder](https://via.placeholder.com/800x450/0d1117/3fb950?text=MOTD+Preview)
> *Live MOTD preview rendered with your selected theme and module order.*

> ![Theme Placeholder](https://via.placeholder.com/800x450/0d1117/d2a8ff?text=Theme+Selector)
> *Browse and switch between 6 built-in color themes in real time.*

---

## Features

- **32 Built-in Modules** — From system info and weather to ZFS, containers, fail2ban, and GPU stats. Each module auto-detects availability based on installed binaries.
- **Interactive Bubble Tea TUI** — Keyboard-driven interface with vim-style navigation. No mouse required.
- **6 Color Themes** — Default, Dracula, Nord, Solarized Dark, Monochrome, and pure ASCII.
- **10 Built-in Templates** — Minimal, Sysadmin, Developer, Hacker, Full, Aesthetic, Security, Homelab, Gaming, Server.
- **Smart Backups** — Every write is backed up. The original MOTD is snapshotted once and kept immutable forever. Rollback anytime.
- **Punctual Sudo** — The TUI never runs as root. Elevation is requested only for the final `cp`/`mkdir`/`chmod` operation.
- **Fully Configurable** — Per-module settings, custom header/footer, module reordering, service picker, and more.
- **Cross-Architecture** — Native binaries for amd64, arm64, and armv7.

---

## Installation

### Prebuilt Binaries

Download the latest release for your architecture:

```bash
curl -L -o mom.tar.gz https://github.com/ams/mom/releases/latest/download/mom_$(curl -s https://api.github.com/repos/ams/mom/releases/latest | grep tag_name | cut -d '"' -f 4)_linux_amd64.tar.gz
tar -xzf mom.tar.gz
sudo install -Dm755 mom /usr/local/bin/mom
```

Available as `.tar.gz`, `.deb`, and `.rpm` in every release.

### Build from Source

Requires **Go 1.25.0+** and a Linux host (or cross-compilation setup).

```bash
git clone https://github.com/ams/mom.git
cd mom
make build        # linux/amd64 binary → bin/mom
make build-all    # amd64 + arm64 + armv7
```

### Run without Installing

```bash
make run
```

---

## Quick Start

```bash
# Launch the interactive TUI
mom

# One-shot setup: enable all available modules and write the MOTD
mom --full-auto

# Apply a built-in template from the command line
mom --apply-template hacker
```

On first run, `mom` detects your distribution, resolves the correct MOTD paths, and creates an immutable backup of your original file at `~/.config/mom/backups/`.

---

## Usage

### TUI Navigation

| Key | Action |
|-----|--------|
| `↑` / `↓` or `k` / `j` | Navigate lists |
| `Enter` | Select / Confirm |
| `Esc` or `q` | Go back / Quit |
| `?` | Show keyboard help |
| `Ctrl+s` | Save and apply MOTD |

### Headless Commands

```bash
mom --version                           # Print version and build info
mom --full-auto                         # Enable all modules and write MOTD
mom --apply-template <name>             # Apply a built-in template
mom --export-template <file.toml>       # Export current config as a reusable template
mom --import-template <file.toml>       # Import and apply a template from file
mom --rollback                          # Restore the original MOTD backup
mom --uninstall                         # Remove mom's changes and restore original
```

### Built-in Templates

| Template | Description |
|----------|-------------|
| `minimal` | Just the essentials: distro logo + system info |
| `sysadmin` | Server-focused: resources, services, updates, logins |
| `developer` | Dev machine: git, tmux, containers, procs |
| `hacker` | Cyberpunk aesthetic: figlet, quotes, weather |
| `full` | Everything available enabled at once |
| `aesthetic` | Clean, curated look with selective modules |
| `security` | Firewall, fail2ban, SSH keys, failed logins, certs |
| `homelab` | Services, ZFS, containers, timers, updates |
| `gaming` | GPU, resources, reboot hint, fun art |
| `server` | Network, traffic, disk I/O, ports, journal |

---

## Modules

`mom` ships with **32 modules**. Each module declares its own dependencies and only activates if the required tools are present on your system.

| Module | What it Shows | Depends On |
|--------|---------------|------------|
| `logo` | Distro ASCII logo | — |
| `system` | Hostname, OS, kernel, uptime, arch | — |
| `resources` | CPU, memory, disk, swap (optional temp) | — |
| `gpu` | GPU model and usage | — |
| `network` | IP addresses, default gateway | `ip` |
| `traffic` | Network RX/TX stats | — |
| `weather` | Current weather for a configured city | `curl` |
| `containers` | Docker/Podman container counts | `docker` or `podman` |
| `services` | Status of user-selected systemd services | `systemctl` |
| `ports` | Listening TCP/UDP ports | `ss` |
| `procs` | Top processes by CPU/memory | `ps` |
| `diskio` | Disk read/write throughput | — |
| `timers` | systemd timer statuses | `systemctl` |
| `journal` | Recent error/warning log entries | `journalctl` |
| `zfs` | ZFS pool health and usage | `zpool` |
| `certs` | TLS certificate expiry dates | `openssl` |
| `firewall` | Firewall status (ufw/firewalld) | `ufw` or `firewalld-cmd` |
| `fail2ban` | Banned IPs and jail status | `fail2ban-client` |
| `failed-logins` | Recent authentication failures | `lastb` |
| `sudo` | Sudo attempts and policy info | — |
| `battery` | Battery charge and status | — |
| `users` | Currently logged-in users | — |
| `sshkeys` | SSH authorized keys count | — |
| `boot` | Last boot time | — |
| `git` | Git repository status in `$HOME` | `git` |
| `tmux` | Active tmux sessions | `tmux` |
| `reboot` | Reboot required indicator | — |
| `updates` | Pending system updates | package manager (apt/dnf/pacman/etc.) |
| `logins` | Last login summary | `last` |
| `calendar` | Current date with events | — |
| `quote` | Random quote or fortune | — |
| `cowsay` | cowsay / figlet / lolcat message | `cowsay`, `figlet`, or `lolcat` |

---

## Themes

Themes define the color palette and text attributes used when rendering the MOTD.

| Theme | Description |
|-------|-------------|
| `default` | Bright 16-color ANSI — works everywhere |
| `dracula` | Vivid pink/purple truecolor scheme |
| `nord` | Cold arctic blues and frosted tones |
| `solarized-dark` | Ethan Schoonover's classic dark palette |
| `monochrome` | Bold/dim/italic only — color-blind friendly |
| `ascii` | Plain text, no colors, no Unicode |

You can also drop custom theme files into the config directory — `mom` picks them up automatically on startup.

---

## Templates

Templates are reusable TOML presets that define which modules are enabled and in what order. The binary embeds 10 templates, and you can export/import your own.

```bash
# Export your current setup as a template
mom --export-template ./my-server.toml

# Share it across machines
mom --import-template ./my-server.toml
```

---

## Configuration

Config lives at `~/.config/mom/config.toml`.

```toml
[motd]
header = "Welcome to $(hostname)"
footer = "Have a productive day!"

[mode]
theme = "dracula"
variant = "default"
module_order = ["logo", "system", "resources", "weather", "cowsay"]

[modules]
system = true
resources = true
weather = true

[modules.weather_config]
city = "London"
units = "metric"

[modules.cowsay_config]
mode = "figlet"
message = "Stay sharp"
```

### Variants

Modules can render in different styles:

| Variant | Style |
|---------|-------|
| `default` | Standard balanced output |
| `compact` | Condensed, less padding |
| `detailed` | Extended info and labels |
| `minimal` | Bare essentials |
| `ascii` | ASCII-only, no Unicode borders |
| `boxed` | Content wrapped in Unicode boxes |

---

## Security

- **Never run the TUI as root.** `mom` detects if it is running as root and refuses to start the interactive interface.
- **Immutable original backup.** The very first time `mom` writes the MOTD, it stores an immutable snapshot of the original. This snapshot can never be overwritten by later backups.
- **Backup before every write.** Each generation creates a timestamped backup so you can restore any previous state.
- **Punctual privilege elevation.** Only the final file-copy operation that writes to `/etc/update-motd.d/` or `/etc/motd` runs via `sudo`. Everything else — generation, preview, config editing — runs as your normal user.

---

## Development

```bash
make build      # Build linux/amd64 binary
make build-all  # Build for amd64, arm64, and armv7
make test       # Run tests with race detector
make lint       # go vet + staticcheck
make fmt        # go fmt
make run        # Run directly without building
make clean      # Remove build artifacts
make release    # goreleaser release --clean
```

### Smoke Test

```bash
./scripts/smoke-test.sh   # vet → build → test → binary-version check
```

---

## License

MIT License — see [LICENSE](LICENSE) for details.

---

<div align="center">

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss) in Go.

</div>
