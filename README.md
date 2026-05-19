# mom

> Interactive TUI (Terminal User Interface) for managing the Linux MOTD (Message of the Day). Customize your login message with modules, templates, and automatic detection — all without writing a single script.

## Requisitos

- **Sistema operativo**: Linux (amd64, arm64, armv7)
- **Go**: 1.25+ (solo para desarrollo / build desde fuente)
- **Dependencias opcionales del sistema**:
  - `cowsay`, `figlet`, `lolcat` — módulo cowsay
  - `docker` / `podman` — módulo containers
  - `apt`, `dnf`, `pacman`, `zypper` — módulo updates
  - `neofetch` / `fastfetch` — logo mejorado
  - `systemctl` — módulo services
  - `cal` / `ncal` — módulo calendar

## Instalación

### Desde release (recomendado)

Descarga el binario para tu arquitectura desde la página de [Releases](../../releases) y colócalo en tu `PATH`:

```bash
curl -L -o mom https://github.com/ams/mom/releases/latest/download/mom_linux_amd64
chmod +x mom
sudo mv mom /usr/local/bin/
```

### Desde fuente

```bash
git clone https://github.com/ams/mom.git
cd mom
make build
sudo cp bin/mom /usr/local/bin/
```

## Uso

### Modo interactivo (TUI)

```bash
mom
```

Navega con las flechas, activa/desactiva módulos con `Espacio`, y aplica los cambios con `Enter`.

### Modo headless (CLI)

```bash
mom --version                          # Muestra la versión
mom --full-auto                        # Configura automáticamente el MOTD
mom --apply-template Sysadmin          # Aplica una plantilla built-in
mom --rollback                         # Restaura el MOTD original
mom --uninstall                        # Desinstala mom y restaura el MOTD
```

## Desarrollo

```bash
# Compilar
make build

# Ejecutar tests
make test

# Linting
make lint

# Ejecutar en modo desarrollo
make run

# Limpiar binarios
make clean

# Release local (requiere GoReleaser)
make release
```

## Licencia

MIT
