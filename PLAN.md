# PLAN.md — `mom` (Message Of the day Manager)

> Documento de especificación y hoja de ruta para agentes de IA (AI Coding Agents).
> Leer completamente antes de escribir una sola línea de código.

---

## 1. Descripción General

`mom` es una herramienta TUI (Terminal User Interface) para Linux que permite a los usuarios personalizar el MOTD (Message of the Day) de forma interactiva, sencilla y segura. Soporta 4 modos de uso: manual, plantillas predefinidas, detección automática de módulos disponibles, y modo full-auto.

**Binario**: `mom`
**Distribución**: Binary releases estáticos (Go) vía GitHub Releases.

---

## 2. Stack Tecnológico

| Componente | Tecnología | Versión mínima |
|------------|-----------|----------------|
| Lenguaje | Go | 1.22+ |
| Framework TUI | [Bubble Tea](https://github.com/charmbracelet/bubbletea) | latest |
| Estilos TUI | [Lipgloss](https://github.com/charmbracelet/lipgloss) | latest |
| Componentes TUI | [Bubbles](https://github.com/charmbracelet/bubbles) | latest |
| Configuración | [BurntSushi/toml](https://github.com/BurntSushi/toml) | latest |
| HTTP (clima) | `net/http` (stdlib) | — |
| Releases | [GoReleaser](https://goreleaser.com/) | latest |
| Build | `Makefile` + Go toolchain | — |

**Dependencias del sistema operativo (no empaquetadas, solo referenciadas en tiempo de ejecución):**

- `cowsay`, `figlet`, `lolcat` — módulo cowsay
- `docker`, `podman` — módulo containers
- `apt`, `dnf`, `pacman`, `zypper` — módulo updates
- `neofetch` / `fastfetch` — opcional para logo mejorado
- `systemctl` — módulo services (systemd)
- `curl` o `wget` — comunicación con wttr.in (se usa `net/http` de Go, no dependencia externa)

---

## 3. Arquitectura y Estructura de Archivos

```
mom/
├── cmd/
│   └── mom/
│       └── main.go                  # Punto de entrada
├── internal/
│   ├── distro/
│   │   ├── detect.go                # Detección de familia de distro
│   │   ├── detect_test.go
│   │   └── paths.go                 # Resolución de rutas MOTD según distro
│   ├── backup/
│   │   ├── backup.go                # Backup automático, lista, gestión
│   │   ├── backup_test.go
│   │   ├── rollback.go              # Comando/lógica de restauración
│   │   └── rollback_test.go
│   ├── config/
│   │   ├── config.go                # Estructura TOML, load/save
│   │   ├── config_test.go
│   │   └── defaults.go              # Valores por defecto
│   ├── module/
│   │   ├── module.go                # Interfaz Module + Registry
│   │   ├── module_test.go
│   │   ├── system.go                # Hostname, kernel, uptime, shell
│   │   ├── resources.go             # CPU, RAM, disco, temperatura
│   │   ├── weather.go               # wttr.in (sin API key)
│   │   ├── cowsay.go                # cowsay / figlet / lolcat wrapper
│   │   ├── network.go               # IP pública y privada
│   │   ├── containers.go            # Docker / Podman status
│   │   ├── updates.go               # Paquetes pendientes (apt/dnf/pacman/zypper)
│   │   ├── logins.go                # Últimos logins, sesiones SSH activas
│   │   ├── quote.go                 # Frase aleatoria del día
│   │   ├── calendar.go              # Calendario del día actual
│   │   ├── services.go              # Estado de servicios systemd
│   │   └── logo.go                  # Logo ASCII de la distro
│   ├── generator/
│   │   ├── generator.go             # Ensamblador de MOTD a partir de módulos
│   │   ├── generator_test.go
│   │   ├── writer.go                # Escritura en la ruta correcta según distro
│   │   └── writer_test.go
│   ├── template/
│   │   ├── template.go              # Definición de Template, gestión
│   │   ├── template_test.go
│   │   ├── builtin.go               # 5 plantillas built-in (embed)
│   │   └── export.go                # Export/import como archivos TOML
│   ├── tui/
│   │   ├── app.go                   # Modelo principal Bubble Tea
│   │   ├── app_test.go
│   │   ├── views/
│   │   │   ├── dashboard.go         # Vista principal / menú
│   │   │   ├── modules.go           # Selector de módulos (checkbox list)
│   │   │   ├── templates.go         # Selector de plantillas
│   │   │   ├── preview.go           # Vista previa del MOTD generado
│   │   │   ├── help.go              # Vista de ayuda / atajos
│   │   │   └── install.go           # Diálogo de instalación asistida de dependencias
│   │   ├── components/
│   │   │   ├── stylist.go           # Helpers de Lipgloss reutilizables
│   │   │   └── statusbar.go         # Barra de estado inferior
│   │   └── keys/
│   │       └── keys.go              # Keybindings centralizados
│   └── permission/
│       ├── sudo.go                  # Elevación puntual con sudo
│       └── sudo_test.go
├── embed/
│   ├── templates/
│   │   ├── minimal.toml
│   │   ├── sysadmin.toml
│   │   ├── developer.toml
│   │   ├── hacker.toml
│   │   └── full.toml
│   └── logos/
│       ├── ubuntu.txt
│       ├── debian.txt
│       ├── arch.txt
│       ├── fedora.txt
│       ├── opensuse.txt
│       └── default.txt
├── scripts/
│   ├── build.sh                     # Build multiplataforma
│   └── smoke-test.sh                # Smoke test post-build
├── go.mod
├── go.sum
├── Makefile                         # Targets: build, test, lint, run, clean, release
├── .goreleaser.yaml                 # Configuración de GoReleaser
└── README.md
```

---

## 4. Instrucciones Generales para el Agente

### 4.1 Reglas de estilo

- **Go idiomático**: Seguir `go fmt`, `go vet`, y las convenciones de Effective Go.
- **Nombrado**: CamelCase exportado, camelCase no exportado.
- **Errores**: Siempre retornar `error` como último valor. Usar `fmt.Errorf` con `%w` para wrapping. Nunca usar `panic` en lógica de negocio. `panic` solo en `main()` para errores irrecuperables de inicialización.
- **Logging**: Usar `log/slog` (stdlib Go 1.21+) para logging estructurado. Niveles: `Debug` para TUI, `Info` para operaciones, `Error` para fallos.
- **Comentarios**: Documentar todas las funciones y tipos exportados con comentarios estilo GoDoc (`// FuncName does X.`).
- **Sin comentarios superfluos**: No comentar lo obvio. Los comentarios deben explicar el "por qué", no el "qué".
- **Manejo de contexto**: Todas las operaciones de I/O deben aceptar `context.Context` como primer parámetro.
- **Timeouts**: Operaciones de red (wttr.in) deben tener timeout de 5 segundos. Operaciones de archivo, 2 segundos.
- **Interfaces pequeñas**: Seguir el principio de "accept interfaces, return structs".
- **Tests**: Usar `testing` estándar + `testify/assert` si es necesario. Nombres de test: `TestPackage_FuncName_Scenario`.

### 4.2 Manejo de errores

- **Errores de usuario**: Mostrar en la TUI con mensaje amigable y sugerencia de solución.
- **Errores de sistema**: Loggear con `slog.Error` y mostrar diálogo de error en TUI.
- **Errores de permisos**: Detectar `os.ErrPermission` y sugerir ejecutar con sudo.
- **Errores de red**: Reintentar 1 vez con backoff de 2s para wttr.in. Si falla, mostrar "sin conexión" en el módulo de clima.
- **NUNCA hacer `os.Exit()` fuera de `main()`**. La TUI debe manejar todos los errores grácilmente.

### 4.3 Validación de permisos (CRÍTICO - SEGURIDAD)

Antes de CUALQUIER operación de escritura en rutas del sistema (`/etc/motd`, `/etc/update-motd.d/`, `/etc/profile.d/`), el agente DEBE:

1. Verificar `os.Geteuid() == 0`. Si no es root, usar elevación puntual (ver `internal/permission/sudo.go`).
2. NUNCA ejecutar toda la aplicación como root. La TUI corre como usuario normal.
3. La elevación con sudo debe ser exclusivamente para la operación de escritura, usando `os/exec` con `sudo`.
4. Si el usuario cancela la elevación (Ctrl+C en prompt de sudo, o contraseña incorrecta), mostrar error claro y NO modificar nada.

### 4.4 Backups (CRÍTICO - SEGURIDAD)

Antes de modificar cualquier archivo del sistema, el agente DEBE:

1. **Primera ejecución de `mom`**: Guardar el MOTD original en `~/.config/mom/backups/original_TIMESTAMP`. Este archivo es **inmutable** — nunca se sobrescribe.
2. **Cada escritura**: Guardar el MOTD actual en `~/.config/mom/backups/backup_TIMESTAMP` antes de sobrescribir.
3. **Listado de backups**: Orden cronológico inverso (más reciente primero).
4. **Rollback**: Restaurar desde cualquier backup. También restaurar el original inmutable con flag `--original`.
5. Los backups se guardan con metadatos: timestamp, distro detectada, lista de módulos activos en ese momento.

### 4.5 Estructura del archivo de configuración

Ruta: `~/.config/mom/config.toml`

```toml
[motd]
# header = texto opcional al inicio del MOTD
header = ""
# footer = texto opcional al final del MOTD
footer = ""

# Módulos activos (orden determina orden en MOTD)
[modules]
system = true
resources = false
weather = false
cowsay = false
network = false
containers = false
updates = false
logins = false
quote = false
calendar = false
services = false
logo = true

# Configuración específica de módulos
[modules.weather]
city = ""          # vacío = auto-detección por IP
units = "metric"   # metric | imperial

[modules.resources]
show_temp = false  # requiere lm-sensors

[modules.cowsay]
mode = "cowsay"    # cowsay | figlet | lolcat | random
message = "Welcome back!"

[modules.updates]
include_aur = false # solo Arch
include_snaps = false

[modules.containers]
runtime = "auto"   # auto | docker | podman

# Modo de operación
[mode]
default = "manual" # manual | template | auto | full-auto
last_template = "" # nombre del último template aplicado
```

### 4.6 Formato de plantillas exportables

Las plantillas exportables usan el mismo esquema TOML que la config, pero solo incluyen las secciones `[motd]` y `[modules]`. Se guardan como archivos `.toml` que el usuario puede compartir.

---

## 5. Fases de Desarrollo (Paso a Paso)

Cada fase es atómica y secuencial. El agente debe completar una fase (incluyendo tests) antes de pasar a la siguiente.

---

### Fase 0: Scaffold del Proyecto

**Objetivo**: Crear la estructura de directorios, inicializar el módulo Go, y configurar las herramientas de build.

**Tareas**:

1. Crear todos los directorios listados en la sección 3.
2. Inicializar módulo Go: `go mod init github.com/<user>/mom` (usar placeholder, se actualizará al repo real).
3. Crear `Makefile` con los siguientes targets:
   - `build`: Compila para linux/amd64 y linux/arm64 con flags `-ldflags="-s -w"`.
   - `test`: Ejecuta `go test ./... -v -race -count=1`.
   - `lint`: Ejecuta `go vet ./...` y `staticcheck ./...` (si está instalado).
   - `run`: Compila y ejecuta `go run ./cmd/mom`.
   - `clean`: Elimina binarios generados.
   - `release`: Invoca GoReleaser con `--clean`.
4. Crear `.goreleaser.yaml` con builds para linux/amd64, linux/arm64, linux/arm (armv7). Formato: tar.gz + deb + rpm.
5. Crear `cmd/mom/main.go` con un `func main()` mínimo que imprima `"mom v0.1.0"` y salga.
6. Crear `scripts/build.sh` que compile para las 3 arquitecturas y genere checksums.
7. Crear `scripts/smoke-test.sh` que ejecute `go vet ./...`, `go build ./...`, y `go test ./...`.

**Verificación**: `make build && make test && make lint` debe ejecutarse sin errores. El binario `./bin/mom` debe imprimir la versión.

---

### Fase 1: Detección de Distro y Resolución de Rutas MOTD

**Objetivo**: Detectar la familia de distro Linux y determinar la ruta correcta donde escribir el MOTD.

**Tareas en `internal/distro/detect.go`**:

1. Definir tipo `Family` como `string` con constantes: `FamilyDebian`, `FamilyRHEL`, `FamilyArch`, `FamilySUSE`, `FamilyUnknown`.
2. Definir struct `Info` con campos: `Family`, `Name` (ej: "Ubuntu 24.04"), `Version`.
3. Implementar función `Detect() (Info, error)`:
   - Leer `/etc/os-release` (formato clave=valor, líneas con `=`).
   - Parsear `ID` y `ID_LIKE` para determinar la familia.
   - Mapeo:
     - `debian`, `ubuntu`, `linuxmint`, `pop`, `elementary` → `FamilyDebian`
     - `rhel`, `fedora`, `centos`, `rocky`, `almalinux` → `FamilyRHEL`
     - `arch`, `manjaro`, `endeavouros` → `FamilyArch`
     - `opensuse`, `suse` → `FamilySUSE`
   - Si `ID_LIKE` contiene alguna de las claves anteriores, usar esa.
4. Manejar el caso de archivo no encontrado (devolver `FamilyUnknown`).

**Tareas en `internal/distro/paths.go`**:

1. Definir struct `MotdPaths` con campos: `MotdFile` (ruta al archivo MOTD), `MotdDir` (directorio de scripts dinámicos, vacío si no aplica), `ProfileScript` (ruta en `/etc/profile.d/`, vacío si no aplica).
2. Implementar función `GetPaths(family Family) MotdPaths`:

   | Family | MotdFile | MotdDir | ProfileScript |
   |--------|----------|---------|---------------|
   | Debian/Ubuntu | `/etc/motd` (enlace o archivo) | `/etc/update-motd.d/` | (vacío) |
   | RHEL/Fedora | `/etc/motd` | `/etc/motd.d/` | `/etc/profile.d/mom-motd.sh` |
   | Arch | `/etc/motd` | (vacío) | `/etc/profile.d/mom-motd.sh` |
   | SUSE | `/etc/motd` | (vacío) | `/etc/profile.d/mom-motd.sh` |
   | Unknown | `/etc/motd` | (vacío) | `/etc/profile.d/mom-motd.sh` |

3. Para Debian/Ubuntu: el MOTD se sirve escribiendo un script numerado en `/etc/update-motd.d/` (ej: `99-mom`). El script simplemente hace `cat` del MOTD generado por `mom`. Esto garantiza que se ejecute en cada login.
4. Para RHEL/Fedora: `mom` crea un script en `/etc/profile.d/mom-motd.sh` que imprime el MOTD generado. Alternativamente, escribe directamente en `/etc/motd` si el mecanismo de `motd.d` no está disponible.
5. Para Arch/openSUSE/Unknown: Escribir directamente en `/etc/motd`. Adicionalmente, crear script en `/etc/profile.d/mom-motd.sh` para la visualización en login interactivo por si el mecanismo `pam_motd` no está configurado.

**Tareas en `internal/distro/detect_test.go`**:

1. Crear archivos `/etc/os-release` temporales simulando cada distro.
2. Test `TestDetect_Ubuntu`, `TestDetect_Fedora`, `TestDetect_Arch`, `TestDetect_OpenSUSE`, `TestDetect_Unknown` (archivo no existe).
3. Test `TestDetect_ID_LIKE` (ej: Linux Mint tiene `ID_LIKE=ubuntu debian`).

**Verificación**: Tests pasan. `go test ./internal/distro/...` verde.

---

### Fase 2: Sistema de Backups

**Objetivo**: Implementar backup automático, backup original inmutable, listado y restauración.

**Tareas en `internal/backup/backup.go`**:

1. Definir struct `Backup` con campos: `Timestamp time.Time`, `Path string`, `Distro string`, `Modules []string` (nombres de módulos activos).
2. Definir struct `Manager` con campo `BackupDir string` (default: `~/.config/mom/backups`).
3. Implementar `(m *Manager) Init() error`:
   - Crear `BackupDir` con `os.MkdirAll(0755)`.
   - Si no existe `original_*` en el directorio, hacer backup del MOTD actual y nombrarlo `original_<timestamp>`.
4. Implementar `(m *Manager) Backup(ctx context.Context, motdPath string, distro string, modules []string) (*Backup, error)`:
   - Leer archivo actual en `motdPath`.
   - Guardar copia en `BackupDir/backup_<timestamp>`.
   - Guardar metadatos (timestamp, distro, modules) en archivo `.meta` JSON junto al backup o en el nombre.
   - Si el archivo actual no existe, crear backup vacío (no error).
5. Implementar `(m *Manager) List() ([]Backup, error)`:
   - Leer `BackupDir`, filtrar archivos que empiecen con `backup_` u `original_`.
   - Parsear timestamps y metadatos.
   - Ordenar por timestamp descendente (más reciente primero).
6. Implementar `(m *Manager) Restore(ctx context.Context, backup *Backup, motdPath string) error`:
   - Requerir confirmación interactiva si se llama desde TUI (el caller maneja esto; Backup Manager solo restaura).
   - Copiar contenido del backup a `motdPath`.
   - NO modificar el archivo de backup (no se borra al restaurar).
7. Implementar `(m *Manager) GetOriginal() (*Backup, error)`:
   - Buscar archivo que empiece con `original_`.
   - Devolver el backup original inmutable.

**Tareas en `internal/backup/rollback.go`**:

1. Función `InteractiveRollback(ctx context.Context, backupDir string, motdPath string)` que:
   - Lista backups.
   - Pide al usuario seleccionar uno (por índice o timestamp). Esta función devuelve la lista; la selección la hace la TUI.
2. El rollback siempre hace un backup del estado actual ANTES de restaurar (para no perder el estado pre-rollback).

**Tests**:
- `TestBackup_Init_CreatesDir`
- `TestBackup_Init_SavesOriginal`
- `TestBackup_Backup_SavesCorrectly`
- `TestBackup_Backup_MultipleBackups`
- `TestBackup_List_ReturnsSorted`
- `TestBackup_Restore_RestoresContent`
- `TestBackup_Original_Immutable`

**Verificación**: `go test ./internal/backup/...` verde.

---

### Fase 3: Sistema de Configuración

**Objetivo**: Leer, escribir y gestionar `~/.config/mom/config.toml`.

**Tareas en `internal/config/config.go`**:

1. Definir struct `Config` que mapee exactamente la estructura TOML de la sección 4.5.
2. Definir struct `ModuleConfig` para las opciones de cada módulo.
3. Implementar `Load() (*Config, error)`:
   - Determinar ruta: `$XDG_CONFIG_HOME/mom/config.toml` o `~/.config/mom/config.toml`.
   - Si el archivo no existe, devolver `Defaults()`.
   - Parsear con `BurntSushi/toml`.
   - Validar: valores de enum (units: metric/imperial, mode: manual/template/auto/full-auto, runtime: auto/docker/podman, cowsay mode: cowsay/figlet/lolcat/random).
4. Implementar `Save(cfg *Config) error`:
   - Crear directorio padre si no existe (`os.MkdirAll`).
   - Serializar con `toml.Marshal`.
   - Escribir con permisos `0644`.
5. Implementar `Defaults() *Config`: devolver configuración con todos los módulos en `false` excepto `system` y `logo`.

**Validaciones**:
- Si `weather.units` no es `metric` ni `imperial` → forzar `metric`.
- Si `cowsay.mode` no es válido → forzar `cowsay`.
- Si `containers.runtime` no es válido → forzar `auto`.
- Si `mode.default` no es válido → forzar `manual`.

**Tests**: Load existente, Load inexistente (defaults), Save and reload, Validate bad values.

**Verificación**: `go test ./internal/config/...` verde.

---

### Fase 4: Interfaz de Módulo y Registry

**Objetivo**: Definir la interfaz que todos los módulos deben implementar, y el registry que los descubre y gestiona.

**Tareas en `internal/module/module.go`**:

1. Definir interfaz `Module`:

```go
type Module interface {
    // Name returns the unique module identifier (e.g. "system", "weather").
    Name() string
    // Title returns the human-readable display name (e.g. "System Information").
    Title() string
    // Description returns a one-line description for the TUI.
    Description() string
    // Dependencies returns external binaries required (e.g. ["cowsay"]).
    // Empty slice means no external deps.
    Dependencies() []string
    // Available returns true if all dependencies are found in PATH.
    Available() bool
    // Generate produces the MOTD output for this module.
    // Returns empty string if the module is disabled or unavailable.
    Generate(ctx context.Context) (string, error)
    // DefaultEnabled returns whether this module should be on by default
    // in auto-detection and full-auto modes.
    DefaultEnabled() bool
}
```

2. Definir struct `Registry` con campo `modules map[string]Module`.
3. Implementar `NewRegistry() *Registry`.
4. Implementar `(r *Registry) Register(m Module)`: añade módulo al mapa.
5. Implementar `(r *Registry) RegisterAll(modules ...Module)`: registro masivo.
6. Implementar `(r *Registry) Get(name string) (Module, bool)`.
7. Implementar `(r *Registry) All() []Module`: devuelve todos ordenados alfabéticamente por Name().
8. Implementar `(r *Registry) Available() []Module`: filtra los que tienen `Available() == true`.
9. Implementar `(r *Registry) Enabled(cfg *config.Config) []Module`: filtra los que están habilitados en la config.
10. Implementar `(r *Registry) CheckDependency(binary string) bool`: busca el binario en `$PATH` usando `exec.LookPath`.

**Tests**: Register, Get existente/inexistente, All sorted, Available filter, Enabled filter.

**Verificación**: `go test ./internal/module/...` verde.

---

### Fase 5: Implementación de los 12 Módulos

**Objetivo**: Implementar cada módulo concreto. Orden recomendado: empezar por los más simples (system, logo, quote) y avanzar hacia los más complejos (weather, updates, containers).

**Reglas comunes para todos los módulos**:
- Cada módulo en su propio archivo dentro de `internal/module/`.
- Cada módulo implementa la interfaz `Module`.
- `Generate(ctx)` nunca debe panicar. Si algo falla, devolver string vacío y loggear error.
- Timeout máximo por módulo: 3 segundos (usar `context.WithTimeout` dentro de `Generate`).
- La salida de `Generate` es texto plano con colores ANSI opcionales (usando Lipgloss o secuencias de escape).
- Si un módulo no está disponible, `Generate` devuelve `("", nil)`.

#### 5.1 Módulo `system` (`internal/module/system.go`)

- **Dependencias**: ninguna (puro Go).
- **Salida**: Hostname (`os.Hostname`), Kernel (`syscall.Uname` o `/proc/version`), Uptime (`/proc/uptime`), Shell (`os.Getenv("SHELL")`), Usuario actual.
- **Formato**:
  ```
  ┌─ System ─────────────────────────┐
  │ Hostname: myhost                  │
  │ Kernel:   Linux 6.8.0-45-generic  │
  │ Uptime:   3 days, 5 hours         │
  │ Shell:    /bin/zsh                │
  │ User:     ams                     │
  └───────────────────────────────────┘
  ```
- **DefaultEnabled**: true.

#### 5.2 Módulo `resources` (`internal/module/resources.go`)

- **Dependencias**: ninguna (lee `/proc/stat`, `/proc/meminfo`, `/proc/diskstats` o `syscall.Statfs`).
- **Salida**: CPU % (calcular delta entre dos lecturas con 1s de intervalo), RAM usada/total, Disco usado/total (montaje `/`).
- **DefaultEnabled**: false.

#### 5.3 Módulo `weather` (`internal/module/weather.go`)

- **Dependencias**: ninguna (usa `net/http`).
- **Salida**: Clima actual de la ciudad configurada (o auto-detectada por IP de wttr.in).
- **Implementación**:
  - GET a `https://wttr.in/<city>?format=3&m` (con `?m` para métrico si `units=metric`).
  - Si `city` está vacío en config, no pasar parámetro de ciudad (wttr.in usa la IP).
  - Timeout 5 segundos.
  - Si falla, devolver `""` (no bloquear el MOTD).
- **DefaultEnabled**: false.

#### 5.4 Módulo `cowsay` (`internal/module/cowsay.go`)

- **Dependencias**: `cowsay`, `figlet`, `lolcat` (según modo).
- **Salida**: Mensaje de bienvenida renderizado con cowsay/figlet/lolcat.
- **Implementación**:
  - `exec.Command(binary, message)` y capturar stdout.
  - Si el binario no existe, mostrar advertencia y devolver `""`.
  - Modo `random`: elige aleatoriamente entre cowsay y figlet (si ambos disponibles).
- **DefaultEnabled**: false.

#### 5.5 Módulo `network` (`internal/module/network.go`)

- **Dependencias**: ninguna.
- **Salida**: IPs privadas (interfaces de red) e IP pública.
- **IP pública**: GET a `https://ifconfig.me` o `https://api.ipify.org`. Timeout 3s.
- **IPs privadas**: `net.InterfaceAddrs()` filtrando IPv4 e IPv6, excluyendo loopback.
- **DefaultEnabled**: false.

#### 5.6 Módulo `containers` (`internal/module/containers.go`)

- **Dependencias**: `docker` y/o `podman` (según config `runtime` o auto-detección).
- **Salida**: Lista de contenedores en ejecución con nombre y estado.
- **Implementación**:
  - En modo `auto`: detectar si `docker` o `podman` están en PATH y usarlos.
  - Ejecutar `docker ps --format "table {{.Names}}\t{{.Status}}"` (similar para podman).
  - Limitar a 10 contenedores máximo.
- **DefaultEnabled**: false.

#### 5.7 Módulo `updates` (`internal/module/updates.go`)

- **Dependencias**: `apt`, `dnf`, `pacman`, o `zypper` (según distro).
- **Salida**: Número de paquetes actualizables.
- **Implementación**:
  - Detectar el package manager según la distro detectada.
  - Debian/Ubuntu: `apt list --upgradable 2>/dev/null | wc -l` (ajustar conteo).
  - RHEL/Fedora: `dnf check-update -q` o contar líneas.
  - Arch: `checkupdates` o `pacman -Qu`. Si `include_aur=true`, `yay -Qu` o `paru -Qu`.
  - SUSE: `zypper list-updates`.
  - Timeout 10s (estas operaciones pueden ser lentas).
  - Si falla, devolver `""`.
- **DefaultEnabled**: false.

#### 5.8 Módulo `logins` (`internal/module/logins.go`)

- **Dependencias**: `last`, `who` (prácticamente siempre disponibles).
- **Salida**: Últimos 5 logins + sesiones SSH activas.
- **Implementación**:
  - `last -5` (últimos logins) parseando la salida.
  - `who` o `/var/run/utmp` para sesiones activas.
  - Mostrar usuario, terminal, fecha, IP de origen.
- **DefaultEnabled**: false.

#### 5.9 Módulo `quote` (`internal/module/quote.go`)

- **Dependencias**: ninguna (frases embebidas).
- **Salida**: Frase aleatoria del día.
- **Implementación**:
  - Array de ~50 frases de tecnología, Linux, programación embebidas en el código.
  - Elegir una al azar (usando `math/rand/v2`).
  - Opcional: si hay conectividad, obtener de una API gratuita (ej: `https://api.quotable.io/random`) como fallback/alternativa.
- **DefaultEnabled**: false.

#### 5.10 Módulo `calendar` (`internal/module/calendar.go`)

- **Dependencias**: `cal` o `ncal` (parte de `util-linux`, prácticamente universal).
- **Salida**: Calendario del mes actual con el día resaltado.
- **Implementación**:
  - Ejecutar `cal` (o `ncal -b` para formato alternativo).
  - Opción de resaltar el día actual con colores ANSI (reemplazar el número del día con versión coloreada).
- **DefaultEnabled**: false.

#### 5.11 Módulo `services` (`internal/module/services.go`)

- **Dependencias**: `systemctl` (solo sistemas con systemd).
- **Salida**: Estado de servicios systemd preseleccionados.
- **Implementación**:
  - Lista de servicios comunes a monitorear: `sshd`, `nginx`, `docker`, `ufw`, `cron`, `fail2ban`.
  - Para cada uno, `systemctl is-active <service>`.
  - Mostrar con icono de color: verde (active), rojo (inactive), amarillo (failed).
  - Si el servicio no existe, omitirlo.
- **DefaultEnabled**: false.

#### 5.12 Módulo `logo` (`internal/module/logo.go`)

- **Dependencias**: `neofetch` o `fastfetch` (opcional). Archivos ASCII embebidos (primario).
- **Salida**: Logo ASCII de la distro.
- **Implementación**:
  - Usar archivos embebidos en `embed/logos/` con `//go:embed`.
  - Seleccionar según la distro detectada.
  - Si no hay logo específico, usar `default.txt`.
  - Opcionalmente, si `neofetch` o `fastfetch` están disponibles, delegar a ellos para el arte ASCII (más detallado).
- **DefaultEnabled**: true.

**Tests**: Test mínimo por módulo — `TestXxx_Name`, `TestXxx_Generate`, `TestXxx_Available` (con mock de PATH).

**Verificación**: `go test ./internal/module/...` verde.

---

### Fase 6: Generador de MOTD y Escritor

**Objetivo**: Ensamblar el MOTD completo a partir de los módulos habilitados y escribirlo en la ruta correcta.

**Tareas en `internal/generator/generator.go`**:

1. Definir struct `Generator` con campos: `Registry *module.Registry`, `Config *config.Config`, `Distro distro.Info`.
2. Implementar `(g *Generator) Generate(ctx context.Context) (string, error)`:
   - Obtener módulos habilitados de la config (`Registry.Enabled`).
   - Respetar el orden de los módulos (el orden en el TOML define el orden de salida).
   - Para cada módulo, llamar `Generate(ctx)` con timeout de 3s por módulo.
   - Insertar separador visual entre módulos (línea de `─` o ` `).
   - Si `header` no está vacío, insertarlo al inicio.
   - Si `footer` no está vacío, insertarlo al final.
   - Si ningún módulo genera salida, devolver `""`.
3. Implementar `(g *Generator) GenerateLive(ctx context.Context, moduleNames []string) (string, error)`:
   - Similar a `Generate` pero solo ejecuta módulos específicos (útil para previsualización en TUI).

**Tareas en `internal/generator/writer.go`**:

1. Definir struct `Writer` con campos: `BackupManager *backup.Manager`, `Paths distro.MotdPaths`, `Distro distro.Info`.
2. Implementar `(w *Writer) Write(ctx context.Context, content string, modules []string) error`:
   - Hacer backup del estado actual (llamar a `BackupManager.Backup`).
   - Según la distro:
     - **Debian/Ubuntu**: Crear/sobrescribir script en `/etc/update-motd.d/99-mom` con contenido:
       ```sh
       #!/bin/sh
       cat <<'MOMEOF'
       <contenido del MOTD>
       MOMEOF
       ```
       Dar permisos `0755` al script. También actualizar `/etc/motd` como symlink o copia (según configuración existente).
     - **RHEL/Fedora con /etc/motd.d/**: Crear script en `/etc/motd.d/mom.sh` (similar a Debian).
     - **Resto (Arch, SUSE, etc.)**: Escribir directamente en `/etc/motd`. Crear también script en `/etc/profile.d/mom-motd.sh`:
       ```sh
       #!/bin/sh
       if [ -f /etc/motd ]; then
           cat /etc/motd
       fi
       ```
   - Todas las escrituras en `/etc/` requieren sudo (usar `internal/permission/sudo.go`).
   - Verificar que el contenido se escribió correctamente (releer y comparar).
3. Implementar `(w *Writer) Remove(ctx context.Context) error`:
   - Desinstalar/limpiar los scripts/archivos creados por `mom`.
   - Restaurar backup original si existe.
   - Útil para `mom --uninstall`.

**Tests**: Test de generación con módulos mock. Test de escritura con directorios temporales simulando `/etc/`.

**Verificación**: `go test ./internal/generator/...` verde.

---

### Fase 7: Sistema de Plantillas

**Objetivo**: Definir plantillas predefinidas y permitir exportar/importar configuraciones como plantillas.

**Tareas en `internal/template/template.go`**:

1. Definir struct `Template` con campos: `Name string`, `Description string`, `Author string`, `ModuleConfig map[string]bool` (nombre de módulo → enabled), `Header string`, `Footer string`.
2. Implementar `Apply(t *Template, cfg *config.Config)`:
   - Sobrescribir `cfg.Motd.Header` y `cfg.Motd.Footer` con los de la plantilla.
   - Sobrescribir `cfg.Modules.*` con los valores de `ModuleConfig`.
   - NO modificar las configuraciones específicas de módulo (weather.city, weather.units, etc.) — solo los booleanos.

**Tareas en `internal/template/builtin.go`**:

Las plantillas built-in se embeben con `//go:embed` desde `embed/templates/*.toml`. Deben ser 5:

| Nombre | Descripción | Módulos activos |
|--------|-------------|-----------------|
| `minimal` | Solo lo esencial | `system`, `logo` |
| `sysadmin` | Enfocado en administración | `system`, `resources`, `updates`, `services`, `logins`, `logo` |
| `developer` | Para desarrolladores | `system`, `resources`, `network`, `containers`, `calendar`, `logo` |
| `hacker` | Estética hacker/cyberpunk | `system`, `network`, `cowsay` (figlet), `quote`, `weather`, `logo` (neofetch si disponible) |
| `full` | Todo activado | Todos los 12 módulos |

Cada archivo `.toml` en `embed/templates/` debe tener este formato:

```toml
name = "Minimal"
description = "Solo lo esencial: información del sistema y logo de la distro"

[motd]
header = ""
footer = ""

[modules]
system = true
logo = true
```

**Tareas en `internal/template/export.go`**:

1. Implementar `Export(cfg *config.Config, path string) error`:
   - Serializar solo las secciones `[motd]` y `[modules]` de la config actual a un archivo TOML.
   - Incluir metadatos: timestamp de exportación, versión de `mom`.
2. Implementar `Import(path string) (*Template, error)`:
   - Leer archivo TOML y parsear a `Template`.
   - Validar que los nombres de módulo existen en el registry.

**Tests**: Apply no modifica weather.units. Round-trip export/import. Built-in templates validan.

**Verificación**: `go test ./internal/template/...` verde.

---

### Fase 8: Manejo de Permisos (Elevación Sudo Puntual)

**Objetivo**: Ejecutar operaciones de escritura con sudo sin correr toda la TUI como root.

**Tareas en `internal/permission/sudo.go`**:

1. Implementar `WriteWithSudo(ctx context.Context, content string, path string, perms os.FileMode) error`:
   - Crear archivo temporal en `/tmp/mom-<random>.txt` con el contenido.
   - Ejecutar `sudo cp /tmp/mom-<random>.txt <path>` y luego `sudo chmod <perms> <path>`.
   - Eliminar el archivo temporal después (incluso si falla, con `defer`).
   - Usar `os/exec.CommandContext(ctx, "sudo", ...)`.
   - Capturar stderr para mostrar errores de sudo (contraseña incorrecta, etc.).
2. Implementar `MkdirWithSudo(ctx context.Context, path string, perms os.FileMode) error`:
   - `sudo mkdir -p <path> && sudo chmod <perms> <path>`.
3. Implementar `ChmodWithSudo(ctx context.Context, path string, perms os.FileMode) error`:
   - `sudo chmod <perms> <path>`.
4. Implementar `IsRoot() bool`:
   - `os.Geteuid() == 0`.
5. Implementar `CheckSudo(ctx context.Context) error`:
   - Ejecutar `sudo -n true` para verificar si hay credenciales cacheadas (sin pedir contraseña).
   - Si falla, informar al usuario que se requerirá contraseña sudo.

**IMPORTANTE**: La elevación NUNCA debe ejecutar shells arbitrarios con input del usuario. Solo comandos fijos con argumentos controlados.

**Tests**: Test de permisos con mocks (simular comandos sudo). Test de IsRoot.

**Verificación**: `go test ./internal/permission/...` verde.

---

### Fase 9: Automatización (Modos Manual, Plantillas, Auto-Detect, Full-Auto)

**Objetivo**: Implementar los 4 modos de operación.

Esta fase es mayormente integración de las fases anteriores con una capa de lógica.

**Tareas**:

1. **Modo Manual** (ya cubierto por la TUI): El usuario activa/desactiva módulos manualmente.
2. **Modo Plantillas** (ya cubierto por Phase 7): El usuario selecciona una plantilla y la aplica.
3. **Modo Auto-Detect**:
   - Escanear el sistema con `Registry.Available()` (detecta qué binarios existen).
   - Activar automáticamente los módulos cuyas dependencias están satisfechas.
   - Activar `system` y `logo` siempre (son `DefaultEnabled`).
   - Mostrar resultado al usuario para que confirme o ajuste.
4. **Modo Full-Auto**:
   - Igual que auto-detect, pero además configura el MOTD sin preguntar.
   - Ideal para scripts de provisioning: `mom --full-auto`.
   - Debe funcionar sin TUI (modo headless) cuando se pasa flag `--full-auto`.
   - Hacer backup y aplicar.

**Implementación**: Funciones helper en un nuevo archivo `internal/mode.go` o dentro del Generator. La TUI invoca estas funciones según el modo seleccionado.

**Verificación**: Test de auto-detect simulando PATH con distintos binarios.

---

### Fase 10: TUI — Aplicación Bubble Tea

**Objetivo**: Construir la interfaz de usuario completa con Bubble Tea, Lipgloss y Bubbles.

**IMPORTANTE**: Esta es la fase más grande. Dividir en sub-fases.

#### 10.1 Estilos y Tema (`internal/tui/components/stylist.go`)

Definir estilos globales con Lipgloss:

- `TitleStyle`: Negrita, color cian, padding 1.
- `HeadingStyle`: Negrita, color magenta.
- `SelectedStyle`: Fondo azul, texto blanco.
- `DisabledStyle`: Color gris oscuro (módulos no disponibles).
- `ErrorStyle`: Color rojo.
- `SuccessStyle`: Color verde.
- `InfoStyle`: Color amarillo.
- `HelpStyle`: Color gris, texto pequeño.
- `BorderStyle`: Bordes redondeados (Lipgloss `Border`).
- `SpinnerStyle`: Animación de carga (usar `spinner` de Bubbles).

#### 10.2 Modelo Principal (`internal/tui/app.go`)

Definir el modelo Bubble Tea principal:

```go
type Model struct {
    // Estado
    state        AppState       // enum: dashboard, modules, templates, preview, help
    width        int
    height       int

    // Dependencias
    registry     *module.Registry
    config       *config.Config
    generator    *generator.Generator
    writer       *generator.Writer
    backupMgr    *backup.Manager
    distroInfo   distro.Info

    // Componentes TUI
    moduleList   list.Model      // Lista de módulos con checkboxes
    templateList list.Model      // Lista de plantillas
    previewVP    viewport.Model  // Viewport para previsualizar MOTD
    spinner      spinner.Model   // Indicador de carga
    status       string          // Mensaje de barra de estado
    error        string          // Mensaje de error temporal

    // Keybindings
    keys         keyMap
}
```

Implementar:
- `NewModel(registry, config, etc.) Model`
- `Init() tea.Cmd`
- `Update(msg tea.Msg) (tea.Model, tea.Cmd)`
- `View() string`

#### 10.3 Navegación y Vistas

Implementar `AppState` como enum con estados:

- **Dashboard** (`internal/tui/views/dashboard.go`):
  - Menú principal con opciones:
    1. "Select Modules" → va a vista modules
    2. "Apply Template" → va a vista templates
    3. "Preview MOTD" → va a vista preview
    4. "Auto-Detect Modules" → ejecuta auto-detección
    5. "Full-Auto Setup" → aplica todo automáticamente
    6. "Save & Apply" → guarda config y escribe MOTD
    7. "Rollback" → lista backups y restaura
    8. "Quit" → sale de la aplicación
  - Mostrar distro detectada en la parte superior.
  - Mostrar resumen de módulos activos.

- **Vista Módulos** (`internal/tui/views/modules.go`):
  - Lista interactiva de 12 módulos con checkbox.
  - Cada item muestra: título, descripción, estado (enabled/disabled/available/unavailable).
  - Módulos no disponibles se muestran en gris con etiqueta "[missing: cowsay]".
  - Al seleccionar un módulo no disponible, ofrecer instalación asistida (ver 10.4).
  - Atajos: `Space` toggle, `a` activar todos los disponibles, `d` desactivar todos, `Esc` volver al dashboard.
  - Cambios se reflejan en `config.Modules` en memoria (no guardados a disco aún).

- **Vista Plantillas** (`internal/tui/views/templates.go`):
  - Lista de plantillas (5 built-in + plantillas importadas).
  - Cada item: nombre, descripción, número de módulos que activa.
  - Al aplicar: preguntar confirmación ("Esto sobrescribirá tu configuración actual").
  - Al aplicar: `template.Apply()` y volver al dashboard con mensaje de éxito.
  - Opción "Import template from file" → abre input para ruta, llama a `template.Import`.

- **Vista Preview** (`internal/tui/views/preview.go`):
  - Viewport con el MOTD generado en tiempo real.
  - Se actualiza cada vez que el usuario cambia módulos (re-generar con `Generator.GenerateLive`).
  - Soporte para scroll si el contenido es más largo que la pantalla.
  - Mostrar spinner mientras se genera.

- **Vista Ayuda** (`internal/tui/views/help.go`):
  - Lista de atajos de teclado disponibles.
  - `?` abre/cierra esta vista desde cualquier pantalla.

#### 10.4 Instalación Asistida de Dependencias (`internal/tui/views/install.go`)

- Cuando el usuario selecciona un módulo no disponible, mostrar diálogo:
  ```
  cowsay is not installed.
  Install it? [y/N]
  ```
- Detectar el package manager según la distro:
  - Debian/Ubuntu: `sudo apt install -y <package>`
  - RHEL/Fedora: `sudo dnf install -y <package>`
  - Arch: `sudo pacman -S --noconfirm <package>`
  - SUSE: `sudo zypper install -y <package>`
- Ejecutar instalación con spinner.
- Si tiene éxito, refrescar disponibilidad del módulo.

#### 10.5 Barra de Estado (`internal/tui/components/statusbar.go`)

- Línea inferior fija con:
  - Teclas de navegación: `[↑↓] Navigate  [Space] Toggle  [Tab] Next  [Esc] Back  [?] Help  [q] Quit`
  - Mensaje de estado temporal (se auto-oculta tras 3 segundos).
  - Indicador de cambios sin guardar: `[unsaved]`.

#### 10.6 Acción "Save & Apply"

Cuando el usuario selecciona "Save & Apply":
1. Guardar config actual con `config.Save(cfg)`.
2. Generar MOTD con `generator.Generate(ctx)`.
3. Escribir MOTD con `writer.Write(ctx, content, enabledModules)`.
4. El writer automáticamente:
   - Hace backup previo.
   - Eleva permisos con sudo (solo para la escritura).
   - Verifica la escritura.
5. Mostrar mensaje de éxito o error en la barra de estado.

#### 10.7 Acción "Rollback"

1. Listar backups con `backupMgr.List()`.
2. Mostrar lista interactiva (fecha, distro, módulos activos en ese backup).
3. Seleccionar uno → confirmar → `backupMgr.Restore(ctx, backup, motdPath)`.
4. Mostrar mensaje de éxito.

#### 10.8 Modo Headless (CLI flags)

Aunque la app es TUI, debe soportar flags para scripting:

```go
// En main.go, procesar flags ANTES de iniciar la TUI
flag.Bool("version", false, "Print version and exit")
flag.Bool("full-auto", false, "Run full-auto setup without TUI")
flag.String("apply-template", "", "Apply a built-in template by name without TUI")
flag.String("export-template", "", "Export current config as template to file")
flag.String("import-template", "", "Import a template from file and apply")
flag.Bool("rollback", false, "Restore a previous backup")
flag.Bool("uninstall", false, "Remove mom's changes and restore original MOTD")
flag.Parse()
```

Si se pasa cualquier flag CLI, ejecutar la acción correspondiente y salir. Si no, iniciar la TUI.

**Verificación**: Ejecutar `./bin/mom --version`. Iniciar TUI con `./bin/mom`. Navegar por todas las vistas. Activar/desactivar módulos. Aplicar plantilla. Previsualizar.

---

### Fase 11: Archivos Embebidos

**Objetivo**: Crear los archivos que se empaquetan en el binario con `//go:embed`.

**Tareas**:

1. Crear `embed/templates/minimal.toml` con la configuración de la plantilla minimal.
2. Crear `embed/templates/sysadmin.toml` con la plantilla sysadmin.
3. Crear `embed/templates/developer.toml` con la plantilla developer.
4. Crear `embed/templates/hacker.toml` con la plantilla hacker.
5. Crear `embed/templates/full.toml` con la plantilla full.
6. Crear logos ASCII para cada distro en `embed/logos/`:
   - `ubuntu.txt`: Logo de Ubuntu en ASCII art (pequeño, ~8 líneas).
   - `debian.txt`: Logo de Debian.
   - `arch.txt`: Logo de Arch Linux.
   - `fedora.txt`: Logo de Fedora.
   - `opensuse.txt`: Logo de openSUSE.
   - `default.txt`: Logo genérico de Linux (Tux).
7. El arte ASCII debe ser compacto (máximo 10 líneas, 40 columnas) para no ocupar demasiado espacio en el MOTD.
8. Usar `//go:embed` en `internal/template/builtin.go` y `internal/module/logo.go`.

---

### Fase 12: Pipeline de Build y Release

**Objetivo**: Configurar GoReleaser y scripts para generar binary releases.

**Tareas**:

1. Configurar `.goreleaser.yaml`:
   ```yaml
   builds:
     - id: mom
       main: ./cmd/mom
       binary: mom
       goos: [linux]
       goarch: [amd64, arm64, arm]
       goarm: [7]
       ldflags:
         - "-s -w -X main.version={{ .Version }} -X main.commit={{ .Commit }} -X main.date={{ .Date }}"
       env:
         - CGO_ENABLED=0

   archives:
     - format: tar.gz
       name_template: "mom_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

   checksum:
     name_template: "checksums.txt"

   nfpms:
     - package_name: mom
       file_name_template: "mom_{{ .Version }}_{{ .Arch }}"
       formats: [deb, rpm]
       vendor: "mom"
       homepage: "https://github.com/<user>/mom"
       maintainer: "<user>"
       description: "Interactive TUI MOTD manager for Linux"
       license: "MIT"

   changelog:
     sort: asc
   ```

2. Implementar `cmd/mom/main.go` con lectura de flags, inicialización de componentes, e inicio de TUI.
3. `Makefile` completo con todos los targets listados en Fase 0.

**Verificación**: `make build` produce binario funcional. `make test` pasa. El binario muestra versión correcta con `--version`.

---

### Fase 13: Testing y Pulido Final

**Objetivo**: Asegurar calidad de código y cobertura de tests.

**Tareas**:

1. **Tests unitarios** (mínimo):
   - `internal/distro/`: 5 tests de detección.
   - `internal/backup/`: 7 tests de backup/rollback.
   - `internal/config/`: 4 tests de load/save/defaults.
   - `internal/module/`: 3 tests de registry + 1 test mínimo por cada módulo.
   - `internal/generator/`: 2 tests de generación + 2 de escritura.
   - `internal/template/`: 3 tests de apply/export/import.
   - `internal/permission/`: 2 tests (IsRoot, mock sudo).

2. **Smoke test** (`scripts/smoke-test.sh`):
   - Compilar el binario.
   - Ejecutar `./bin/mom --version`.
   - Verificar que imprime la versión correctamente.
   - (Opcional) Simular un directorio `/etc` temporal y ejecutar `./bin/mom --full-auto` en un contenedor.

3. **Formateo y linting**:
   - `go fmt ./...` no debe producir cambios.
   - `go vet ./...` no debe reportar issues.
   - Si `staticcheck` está disponible: `staticcheck ./...` limpio.

4. **Actualizar `go.mod`**:
   - `go mod tidy` para limpiar dependencias.
   - `go mod verify` para verificar integridad.

5. **README.md** final:
   - Instalación (descargar binario, mover a `/usr/local/bin`).
   - Uso básico: `mom` abre TUI, `mom --full-auto` para automatizar.
   - Capturas de pantalla ASCII de la TUI (con `vhs` o similares, opcional).
   - Requisitos del sistema (Linux, Go no necesario si se usa binary release).

**Verificación FINAL**: `make test && make lint && make build` todo verde. El binario `./bin/mom` inicia la TUI correctamente.

---

## 6. Apéndices

### A. Orden de desarrollo recomendado para un agente de IA

```
Fase 0  → Scaffold
Fase 1  → Distro detection
Fase 2  → Backups
Fase 3  → Config
Fase 4  → Module interface + Registry
Fase 5  → 12 modules (system, logo, quote, calendar, resources, network, weather, cowsay, updates, logins, services, containers)
Fase 6  → Generator + Writer
Fase 7  → Templates (built-in + export/import)
Fase 8  → Sudo elevation
Fase 9  → Automation modes
Fase 10 → TUI (styles → model → dashboard → modules → templates → preview → install → apply → rollback → CLI flags)
Fase 11 → Embed files (logos + template TOML)
Fase 12 → Build pipeline
Fase 13 → Testing + polish
```

### B. Glosario

| Término | Definición |
|---------|-----------|
| MOTD | Message of the Day — mensaje que se muestra al iniciar sesión en Linux |
| TUI | Terminal User Interface — interfaz de usuario en terminal |
| Bubble Tea | Framework TUI para Go basado en Elm Architecture |
| Lipgloss | Librería de estilos para terminal en Go |
| Bubbles | Componentes TUI reutilizables para Bubble Tea |
| wttr.in | Servicio gratuito de clima sin API key |
| update-motd.d | Mecanismo de Debian/Ubuntu para MOTD dinámico mediante scripts |
| pam_motd | Módulo PAM que muestra el MOTD en el login |

### C. Notas de seguridad

- `mom` NUNCA debe modificar `/etc/motd` o archivos en `/etc/` sin el consentimiento explícito del usuario.
- Los backups son la primera línea de defensa. Siempre backup antes de escribir.
- La elevación con sudo es puntual: solo el comando `cp`/`mkdir`/`chmod` se ejecuta con sudo, no toda la aplicación.
- No se almacenan ni transmiten credenciales. La API de clima (wttr.in) no requiere API key.
- Las dependencias externas se instalan con confirmación explícita del usuario.
