package generator

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/msalexms/MoM/internal/backup"
	"github.com/msalexms/MoM/internal/config"
	"github.com/msalexms/MoM/internal/distro"
	"github.com/msalexms/MoM/internal/i18n"
	"github.com/msalexms/MoM/internal/permission"
)

// MotdPath is the static MOTD every distribution shows at login. mom does
// not write its output here: it only puts back the original content when
// an earlier version of mom had replaced it.
const MotdPath = "/etc/motd"

// Login hooks. Debian and Ubuntu run /etc/update-motd.d/* as root at every
// login (pam_motd); the other families source /etc/profile.d/*.sh in login
// shells. Either way the hook runs "mom --motd", so the MOTD is rendered
// with the data of that very moment.
const (
	debianHook  = "/etc/update-motd.d/99-mom"
	profileHook = "/etc/profile.d/mom-motd.sh"
)

// ErrLegacyBackup marks backups that hold text instead of a config: the
// original /etc/motd and backups made by older versions of mom. They can
// be viewed but not installed again.
var ErrLegacyBackup = errors.New("legacy backup")

// Writer installs and removes the login hook and its config.
type Writer struct {
	backups *backup.Manager
	family  distro.Family
	exe     string // absolute path of the mom binary the hook runs
}

// NewWriter creates a writer for the given distribution family.
func NewWriter(bm *backup.Manager, family distro.Family) *Writer {
	exe, err := os.Executable()
	if err == nil {
		if real, err := filepath.EvalSymlinks(exe); err == nil {
			exe = real
		}
	} else {
		exe = "/usr/local/bin/mom"
	}
	return &Writer{backups: bm, family: family, exe: exe}
}

func (w *Writer) hook() (path, script string, mode os.FileMode) {
	bin := "'" + strings.ReplaceAll(w.exe, "'", `'\''`) + "'"
	const header = "# Installed by mom (amstech.dev): shows the MOTD with live data at login.\n"
	if w.family == distro.FamilyDebian {
		return debianHook, "#!/bin/sh\n" + header +
			"[ -x " + bin + " ] || exit 0\n" +
			"exec " + bin + " --motd\n", 0o755
	}
	return profileHook, header +
		"case $- in *i*) [ -x " + bin + " ] && " + bin + " --motd ;; esac\n", 0o644
}

// Install puts cfg in place as the system config and installs the login
// hook. The config is also kept in the backup history.
func (w *Writer) Install(ctx context.Context, cfg *config.Config) error {
	data, err := config.Encode(cfg)
	if err != nil {
		return err
	}
	if err := permission.WriteFile(ctx, config.SystemPath, string(data), 0o644); err != nil {
		return fmt.Errorf(i18n.T("writing %s: %w"), config.SystemPath, err)
	}

	path, script, mode := w.hook()
	if err := permission.WriteFile(ctx, path, script, mode); err != nil {
		return fmt.Errorf(i18n.T("writing %s: %w"), path, err)
	}
	for _, other := range []string{debianHook, profileHook} {
		if other != path {
			if err := permission.Remove(ctx, other); err != nil {
				return fmt.Errorf(i18n.T("removing %s: %w"), other, err)
			}
		}
	}

	// Earlier versions wrote their output into /etc/motd; put the original
	// back so the MOTD is not shown twice.
	if err := w.restoreOriginalMotd(ctx); err != nil {
		return err
	}

	if _, err := w.backups.Snapshot(data, "installed", cfg.EnabledModuleNames()); err != nil {
		return fmt.Errorf(i18n.T("saving the backup: %w"), err)
	}
	return nil
}

// Reinstall installs the config stored in a backup and returns it.
func (w *Writer) Reinstall(ctx context.Context, b *backup.Backup) (*config.Config, error) {
	cfg, err := w.BackupConfig(b)
	if err != nil {
		return nil, err
	}
	return cfg, w.Install(ctx, cfg)
}

// BackupConfig decodes the config stored in a backup. It returns
// ErrLegacyBackup for backups that hold text instead of a config.
func (w *Writer) BackupConfig(b *backup.Backup) (*config.Config, error) {
	if b.IsOriginal {
		return nil, ErrLegacyBackup
	}
	data, err := w.backups.Read(b)
	if err != nil {
		return nil, err
	}
	if !bytes.Contains(data, []byte("[modules]")) {
		return nil, ErrLegacyBackup
	}
	cfg, err := config.Decode(data)
	if err != nil {
		return nil, ErrLegacyBackup
	}
	return cfg, nil
}

// InstalledConfig returns the raw system config, or nil if mom is not
// installed.
func (w *Writer) InstalledConfig() []byte {
	data, err := os.ReadFile(config.SystemPath)
	if err != nil {
		return nil
	}
	return data
}

// Uninstall removes the hook and the system config and puts the original
// /etc/motd back.
func (w *Writer) Uninstall(ctx context.Context) error {
	for _, f := range []string{debianHook, profileHook, config.SystemPath} {
		if err := permission.Remove(ctx, f); err != nil {
			return fmt.Errorf(i18n.T("removing %s: %w"), f, err)
		}
	}
	if err := permission.RemoveDirIfEmpty(ctx, filepath.Dir(config.SystemPath)); err != nil {
		return err
	}
	return w.restoreOriginalMotd(ctx)
}

func (w *Writer) restoreOriginalMotd(ctx context.Context) error {
	orig, err := w.backups.Original()
	if err != nil || orig == nil {
		return err
	}
	want, err := w.backups.Read(orig)
	if err != nil {
		return err
	}
	have, err := os.ReadFile(MotdPath)
	if (err == nil && bytes.Equal(have, want)) || (os.IsNotExist(err) && len(want) == 0) {
		return nil
	}
	if err := permission.WriteFile(ctx, MotdPath, string(want), 0o644); err != nil {
		return fmt.Errorf(i18n.T("writing %s: %w"), MotdPath, err)
	}
	return nil
}
