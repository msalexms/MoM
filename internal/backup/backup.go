// Package backup provides MOTD backup management with automatic snapshots,
// immutable original preservation, and rollback capabilities.
package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Backup represents a saved MOTD state.
type Backup struct {
	Timestamp  time.Time `json:"timestamp"`
	Path       string    `json:"path"`
	Distro     string    `json:"distro"`
	Modules    []string  `json:"modules"`
	IsOriginal bool      `json:"is_original"`
}

// Manager handles backup creation, listing, and restoration.
type Manager struct {
	BackupDir string
}

// NewManager creates a new backup manager with the given backup directory.
func NewManager(backupDir string) *Manager {
	return &Manager{BackupDir: backupDir}
}

// DefaultBackupDir returns the default backup directory path.
func DefaultBackupDir() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "mom", "backups")
}

// Init creates the backup directory and saves the original MOTD if it hasn't
// been saved before.
func (m *Manager) Init(ctx context.Context, motdPath string) error {
	if err := os.MkdirAll(m.BackupDir, 0755); err != nil {
		return fmt.Errorf("creating backup directory: %w", err)
	}

	// Check if original backup already exists
	original, _ := m.GetOriginal()
	if original != nil {
		return nil // Already initialized
	}

	// Save the current MOTD as the immutable original
	content, err := readFileOrEmpty(motdPath)
	if err != nil {
		return fmt.Errorf("reading current MOTD for original backup: %w", err)
	}

	ts := time.Now()
	backupFile := filepath.Join(m.BackupDir, fmt.Sprintf("original_%s", formatTimestamp(ts)))
	if err := os.WriteFile(backupFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing original backup: %w", err)
	}

	meta := Backup{
		Timestamp:  ts,
		Path:       backupFile,
		Distro:     "",
		Modules:    nil,
		IsOriginal: true,
	}
	if err := m.writeMeta(backupFile, &meta); err != nil {
		return fmt.Errorf("writing original metadata: %w", err)
	}

	return nil
}

// Backup creates a new backup of the current MOTD file before modification.
func (m *Manager) Backup(ctx context.Context, motdPath string, distro string, modules []string) (*Backup, error) {
	content, err := readFileOrEmpty(motdPath)
	if err != nil {
		return nil, fmt.Errorf("reading MOTD for backup: %w", err)
	}

	ts := time.Now()
	backupFile := filepath.Join(m.BackupDir, fmt.Sprintf("backup_%s", formatTimestamp(ts)))
	if err := os.WriteFile(backupFile, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("writing backup: %w", err)
	}

	backup := &Backup{
		Timestamp:  ts,
		Path:       backupFile,
		Distro:     distro,
		Modules:    modules,
		IsOriginal: false,
	}
	if err := m.writeMeta(backupFile, backup); err != nil {
		return nil, fmt.Errorf("writing backup metadata: %w", err)
	}

	return backup, nil
}

// List returns all backups sorted by timestamp descending (most recent first).
func (m *Manager) List() ([]Backup, error) {
	entries, err := os.ReadDir(m.BackupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading backup directory: %w", err)
	}

	var backups []Backup
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "backup_") && !strings.HasPrefix(name, "original_") {
			continue
		}
		// Skip metadata files
		if strings.HasSuffix(name, ".meta") {
			continue
		}

		backupPath := filepath.Join(m.BackupDir, name)
		meta, err := m.readMeta(backupPath)
		if err != nil {
			// If no metadata, create a basic entry from filename
			meta = &Backup{
				Path:       backupPath,
				IsOriginal: strings.HasPrefix(name, "original_"),
			}
		}
		backups = append(backups, *meta)
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	return backups, nil
}

// Restore copies the content of a backup to the target MOTD path.
// It creates a backup of the current state before restoring.
func (m *Manager) Restore(ctx context.Context, backup *Backup, motdPath string) error {
	content, err := os.ReadFile(backup.Path)
	if err != nil {
		return fmt.Errorf("reading backup file: %w", err)
	}

	// Backup current state before restoring (pre-rollback safety)
	if _, err := m.Backup(ctx, motdPath, "pre-rollback", nil); err != nil {
		return fmt.Errorf("pre-rollback backup: %w", err)
	}

	if err := os.WriteFile(motdPath, content, 0644); err != nil {
		return fmt.Errorf("restoring MOTD: %w", err)
	}

	return nil
}

// GetOriginal returns the immutable original backup, if it exists.
func (m *Manager) GetOriginal() (*Backup, error) {
	entries, err := os.ReadDir(m.BackupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading backup directory: %w", err)
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "original_") && !strings.HasSuffix(entry.Name(), ".meta") {
			backupPath := filepath.Join(m.BackupDir, entry.Name())
			meta, err := m.readMeta(backupPath)
			if err != nil {
				meta = &Backup{
					Path:       backupPath,
					IsOriginal: true,
				}
			}
			return meta, nil
		}
	}

	return nil, nil
}

// GetContent reads and returns the content of a backup file.
func (m *Manager) GetContent(backup *Backup) (string, error) {
	content, err := os.ReadFile(backup.Path)
	if err != nil {
		return "", fmt.Errorf("reading backup content: %w", err)
	}
	return string(content), nil
}

// writeMeta writes backup metadata alongside the backup file.
func (m *Manager) writeMeta(backupPath string, backup *Backup) error {
	metaPath := backupPath + ".meta"
	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(metaPath, data, 0644)
}

// readMeta reads backup metadata from the .meta file alongside the backup.
func (m *Manager) readMeta(backupPath string) (*Backup, error) {
	metaPath := backupPath + ".meta"
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}
	var backup Backup
	if err := json.Unmarshal(data, &backup); err != nil {
		return nil, err
	}
	return &backup, nil
}

// readFileOrEmpty reads a file, returning empty string if it doesn't exist.
func readFileOrEmpty(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(content), nil
}

// formatTimestamp returns a filesystem-safe timestamp string with nanosecond precision.
func formatTimestamp(t time.Time) string {
	return t.Format("20060102_150405.000000000")
}
