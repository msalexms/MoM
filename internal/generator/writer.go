package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ams/mom/internal/backup"
	"github.com/ams/mom/internal/distro"
	"github.com/ams/mom/internal/permission"
)

// Writer handles writing MOTD content to the appropriate system path.
type Writer struct {
	BackupManager *backup.Manager
	Paths         distro.MotdPaths
	Distro        distro.Info
}

// NewWriter creates a new MOTD writer.
func NewWriter(bm *backup.Manager, paths distro.MotdPaths, di distro.Info) *Writer {
	return &Writer{
		BackupManager: bm,
		Paths:         paths,
		Distro:        di,
	}
}

// Write saves the MOTD content to the system, creating backups before modification.
func (w *Writer) Write(ctx context.Context, content string, modules []string) error {
	// 1. Backup current state
	if _, err := w.BackupManager.Backup(ctx, w.Paths.MotdFile, string(w.Distro.Family), modules); err != nil {
		return fmt.Errorf("pre-write backup: %w", err)
	}

	// 2. Write based on distribution family
	switch w.Distro.Family {
	case distro.FamilyDebian:
		return w.writeDebian(ctx, content)
	case distro.FamilyRHEL:
		return w.writeRHEL(ctx, content)
	default:
		return w.writeGeneric(ctx, content)
	}
}

// writeDebian creates a script in /etc/update-motd.d/ for dynamic MOTD.
func (w *Writer) writeDebian(ctx context.Context, content string) error {
	script := fmt.Sprintf("#!/bin/sh\ncat <<'MOMEOF'\n%s\nMOMEOF\n", content)
	scriptPath := filepath.Join(w.Paths.MotdDir, w.Paths.ScriptName)

	// Ensure directory exists
	if err := permission.MkdirWithSudo(ctx, w.Paths.MotdDir, 0755); err != nil {
		return fmt.Errorf("creating motd dir: %w", err)
	}

	if err := permission.WriteWithSudo(ctx, script, scriptPath, 0755); err != nil {
		return fmt.Errorf("writing motd script: %w", err)
	}

	// Also write static /etc/motd as fallback
	if err := permission.WriteWithSudo(ctx, content+"\n", w.Paths.MotdFile, 0644); err != nil {
		// Non-fatal: dynamic script is the primary mechanism
		_ = err
	}

	return nil
}

// writeRHEL writes to /etc/motd and creates a profile.d script.
func (w *Writer) writeRHEL(ctx context.Context, content string) error {
	// Write /etc/motd
	if err := permission.WriteWithSudo(ctx, content+"\n", w.Paths.MotdFile, 0644); err != nil {
		return fmt.Errorf("writing motd: %w", err)
	}

	// Create profile.d script for interactive shells
	if w.Paths.ProfileScript != "" {
		profileScript := "#!/bin/sh\nif [ -f /etc/motd ]; then\n    cat /etc/motd\nfi\n"
		if err := permission.WriteWithSudo(ctx, profileScript, w.Paths.ProfileScript, 0755); err != nil {
			return fmt.Errorf("writing profile script: %w", err)
		}
	}

	return nil
}

// writeGeneric writes directly to /etc/motd and optionally creates a profile.d script.
func (w *Writer) writeGeneric(ctx context.Context, content string) error {
	// Write /etc/motd
	if err := permission.WriteWithSudo(ctx, content+"\n", w.Paths.MotdFile, 0644); err != nil {
		return fmt.Errorf("writing motd: %w", err)
	}

	// Create profile.d script
	if w.Paths.ProfileScript != "" {
		profileScript := "#!/bin/sh\nif [ -f /etc/motd ]; then\n    cat /etc/motd\nfi\n"
		if err := permission.WriteWithSudo(ctx, profileScript, w.Paths.ProfileScript, 0755); err != nil {
			return fmt.Errorf("writing profile script: %w", err)
		}
	}

	return nil
}

// Remove uninstalls all mom-created files and restores the original MOTD.
func (w *Writer) Remove(ctx context.Context) error {
	// Remove dynamic script if exists
	if w.Paths.MotdDir != "" && w.Paths.ScriptName != "" {
		scriptPath := filepath.Join(w.Paths.MotdDir, w.Paths.ScriptName)
		if _, err := os.Stat(scriptPath); err == nil {
			permission.RemoveWithSudo(ctx, scriptPath)
		}
	}

	// Remove profile.d script if exists
	if w.Paths.ProfileScript != "" {
		if _, err := os.Stat(w.Paths.ProfileScript); err == nil {
			permission.RemoveWithSudo(ctx, w.Paths.ProfileScript)
		}
	}

	// Restore original MOTD
	if err := w.BackupManager.RollbackToOriginal(ctx, w.Paths.MotdFile); err != nil {
		return fmt.Errorf("restoring original MOTD: %w", err)
	}

	return nil
}
