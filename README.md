<div align="center">

<img src="assets/banner.svg" alt="mom — Message Of the Day Manager" width="900" />

[![GitHub release](https://img.shields.io/github/v/release/ams/mom?style=flat-square&label=latest)](https://github.com/msalexms/MoM/releases)
[![Go Version](https://img.shields.io/badge/go-1.25.0-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![Platform](https://img.shields.io/badge/platform-Linux-FCC624?style=flat-square&logo=linux&logoColor=black)](https://github.com/msalexms/MoM)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](LICENSE)
[![Build](https://img.shields.io/badge/build-amd64%20%7C%20arm64%20%7C%20armv7-success?style=flat-square)](https://github.com/msalexms/MoM/releases)

**Your terminal deserves a better welcome.**

`mom` is an interactive TUI that turns your boring static MOTD into a live, beautiful, data-rich dashboard — without editing a single text file.

[Installation](#installation) · [Quick Start](#quick-start) · [Themes](#themes) · [Templates](#templates) · [Configuration](#configuration)

</div>

---

<!-- GIF: Full TUI workflow (~900px wide, ~15-20s loop)
     Record: launch mom → navigate dashboard → toggle modules → preview → apply -->
<div align="center">
  <img src="assets/demo.gif" alt="mom TUI demo" width="900" />
  <br/>
  <sub><i>Configure, preview, and apply your MOTD without leaving the terminal.</i></sub>
</div>

---

## Why mom?

- 🚫 No more hand-editing `/etc/motd` or writing fragile shell scripts.
- 🔌 **32 modules** that auto-detect their dependencies — if a tool isn't installed, the module silently skips itself.
- 🎨 Pick a theme, choose your modules, preview the result, apply — all without leaving the terminal.
- 🔒 Never runs as root. Sudo is used only for the final file write.
- ↩️ Messed up? One command to rollback to the original.

---

## Quick Start

```bash
# Interactive TUI
mom

# Zero-config: detect everything and write the MOTD now
mom --full-auto

# Apply a curated preset
mom --apply-template sysadmin
```

On first run, `mom` detects your distro, resolves MOTD paths, and stores an immutable backup of the original file.

---

## Installation

**One-liner (amd64):**

```bash
curl -sL https://github.com/msalexms/MoM/releases/latest/download/mom_linux_amd64.tar.gz | tar xz && sudo install -Dm755 mom /usr/local/bin/mom
```

**Packages:**

```bash
sudo dpkg -i mom_*_amd64.deb   # Debian / Ubuntu
sudo rpm -i mom_*_amd64.rpm    # Fedora / RHEL
```

**From source** (Go 1.25+):

```bash
git clone https://github.com/msalexms/MoM.git && cd mom && make build
```

Releases ship `.tar.gz`, `.deb`, and `.rpm` for amd64, arm64, and armv7.

---

## Features

| | |
|---|---|
| **Interactive TUI** | Keyboard-driven, vim-style. 12 dashboard actions, no mouse needed. |
| **32 Modules** | System, resources, GPU, ZFS, containers, fail2ban, weather, git, and more. |
| **6 Themes** | Default, Dracula, Nord, Solarized Dark, Monochrome, ASCII. Custom themes via TOML. |
| **10 Templates** | Minimal, Sysadmin, Developer, Hacker, Full, Aesthetic, Security, Homelab, Gaming, Server. |
| **Module Settings** | Unified settings hub — configure any module from one place. |
| **Profiles** | Save, load, and delete config profiles for quick switching between setups. |
| **Smart Backups** | Every write is snapshotted. Original MOTD is immutable. Rollback in one command. |
| **Export / Import** | Share your setup across machines as a TOML file. |
| **Cross-arch** | Native static binaries for amd64, arm64, armv7. |

---

## TUI Navigation

| Key | Action |
|-----|--------|
| `↑`/`↓` or `k`/`j` | Move cursor |
| `Enter` | Select / Confirm |
| `Space` | Toggle |
| `Tab` | Next field |
| `s` / `Ctrl+s` | Save & apply MOTD |
| `a` / `d` | Enable / Disable all modules |
| `Esc` | Back / Cancel |
| `q` / `Ctrl+c` | Quit |
| `?` | Help overlay |

---

## Themes

Six built-in themes. Drop a custom `.toml` into `~/.config/mom/` and it loads automatically.

| Theme | Style |
|-------|-------|
| `default` | Bright 16-color ANSI — works everywhere |
| `dracula` | Vivid pink/purple truecolor |
| `nord` | Arctic blues, muted greens |
| `solarized-dark` | Schoonover's classic dark |
| `monochrome` | Bold/dim/italic only — zero color |
| `ascii` | Plain text, no Unicode — pipe-friendly |

Switch via TUI Theme Picker or set `theme = "dracula"` in config.

---

## Templates

Curated presets that define which modules are enabled and their order. Template names are case-insensitive.

| Template | Focus |
|----------|-------|
| `minimal` | Logo + system info only |
| `sysadmin` | Resources, services, updates, logins |
| `developer` | Git, tmux, containers, procs |
| `hacker` | Figlet, weather, quote, network |
| `full` | Every available module |
| `aesthetic` | Logo, system, calendar, quote |
| `security` | Firewall, fail2ban, SSH keys, certs |
| `homelab` | Services, ZFS, containers, timers |
| `gaming` | GPU, resources, reboot, cowsay |
| `server` | Network, traffic, disk I/O, ports, journal |

```bash
mom --export-template ./my-setup.toml   # save your config
mom --import-template ./my-setup.toml   # apply on another machine
```

---

## Modules

<details>
<summary><strong>32 modules</strong> — click to expand full list</summary>

| Module | Output | Requires |
|--------|--------|----------|
| `logo` | Distro ASCII logo | — |
| `system` | Hostname, OS, kernel, uptime, arch | — |
| `resources` | CPU, memory, disk, swap | — |
| `gpu` | GPU model and utilization | — |
| `network` | IP addresses and gateway | `ip` |
| `traffic` | Network RX/TX counters | — |
| `weather` | Current conditions | `curl` |
| `containers` | Running containers | `docker` / `podman` |
| `services` | Systemd service status | `systemctl` |
| `ports` | Listening TCP/UDP ports | `ss` |
| `procs` | Top processes by CPU/mem | `ps` |
| `diskio` | Disk read/write throughput | — |
| `timers` | Systemd timer next-run | `systemctl` |
| `journal` | Recent error/warning logs | `journalctl` |
| `zfs` | Pool health and usage | `zpool` |
| `certs` | TLS certificate expiry | `openssl` |
| `firewall` | Firewall rules/status | `ufw` / `firewall-cmd` |
| `fail2ban` | Banned IPs and jails | `fail2ban-client` |
| `failed-logins` | Failed auth attempts | `lastb` |
| `sudo` | Recent sudo usage | — |
| `battery` | Charge and power state | — |
| `users` | Logged-in users | — |
| `sshkeys` | SSH authorized key count | — |
| `boot` | Last boot timestamp | — |
| `git` | Repos with uncommitted changes | `git` |
| `tmux` | Active tmux sessions | `tmux` |
| `reboot` | Pending reboot indicator | — |
| `updates` | Pending package updates | `apt`/`dnf`/`pacman`/… |
| `logins` | Last login summary | `last` |
| `calendar` | Date and upcoming events | — |
| `quote` | Random motivational quote | — |
| `cowsay` | ASCII art message | `cowsay`/`figlet`/`lolcat` |

</details>

---

## Configuration

Config lives at `~/.config/mom/config.toml` (auto-created on first run).

```toml
[mode]
theme   = "dracula"
variant = "default"
module_order = ["logo", "system", "resources", "network", "weather"]

[modules]
logo      = true
system    = true
resources = true
network   = true
weather   = true

[modules.weather_config]
city  = "London"
units = "metric"

[modules.cowsay_config]
mode    = "figlet"
message = "Stay sharp."
```

All module settings are also editable from the TUI via **Module Settings** (⚙).

### Render Variants

| Variant | Style |
|---------|-------|
| `default` | Standard key/value layout |
| `compact` | Reduced spacing |
| `minimal` | Bare essentials |
| `ascii` | ASCII borders, no Unicode |
| `boxed` | Unicode box frames |

---

## Headless Flags

```bash
mom --version                         # Version, commit, build date
mom --full-auto                       # Enable all modules and write MOTD
mom --apply-template <name>           # Apply a built-in template
mom --export-template <file.toml>     # Export config as template
mom --import-template <file.toml>     # Import and apply template
mom --rollback                        # Restore original MOTD
mom --uninstall                       # Remove all mom changes
```

---

## Security

- **No TUI as root.** Refuses to start interactively under `sudo`.
- **Immutable original.** First backup is locked — never overwritten.
- **Snapshot every write.** Timestamped backups, any state restorable.
- **Punctual sudo.** Only the final `cp`/`chmod` runs elevated.

---

## Development

```bash
make build      # linux/amd64 → bin/mom
make build-all  # amd64 + arm64 + armv7
make test       # go test ./... -v -race -count=1
make lint       # go vet + staticcheck
make fmt        # go fmt ./...
make run        # go run ./cmd/mom
make clean      # rm -rf bin/
```

```bash
./scripts/smoke-test.sh   # vet → build → test → version check
```

---

## License

MIT — see [LICENSE](LICENSE).

---

<div align="center">
  <sub>Built with <a href="https://github.com/charmbracelet/bubbletea">Bubble Tea</a>, <a href="https://github.com/charmbracelet/lipgloss">Lipgloss</a>, and <a href="https://github.com/charmbracelet/bubbles">Bubbles</a> · Linux only · Go 1.25</sub>
</div>
