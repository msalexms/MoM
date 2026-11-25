// Package backup keeps the history of what mom installed: the untouched
// /etc/motd from before mom (the "original", saved the first time mom runs
// and never modified) plus a copy of the config of every install, so any
// of them can be previewed and installed again.
package backup

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/msalexms/MoM/internal/i18n"
)

// maxSnapshots is how many regular snapshots are kept; older ones are
// pruned. The original is never pruned.
const maxSnapshots = 30

const (
	prefixOriginal = "original_"
	prefixSnapshot = "backup_"
	metaSuffix     = ".meta"
)

// Backup describes one saved copy of the MOTD.
type Backup struct {
	Timestamp  time.Time `json:"timestamp"`
	Path       string    `json:"path"`
	Note       string    `json:"note,omitempty"`
	Modules    []string  `json:"modules,omitempty"`
	IsOriginal bool      `json:"is_original"`
}

// Manager stores backups as plain files plus a JSON .meta sidecar.
type Manager struct {
	dir string
}

// NewManager creates a manager that stores backups in dir.
func NewManager(dir string) *Manager {
	return &Manager{dir: dir}
}

// DefaultDir returns $XDG_CONFIG_HOME/mom/backups (or ~/.config/mom/backups).
func DefaultDir() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "mom", "backups")
}

// Init creates the backup directory and, the first time only, saves the
// current MOTD as the immutable original.
func (m *Manager) Init(motdPath string) error {
	if err := os.MkdirAll(m.dir, 0o755); err != nil {
		return fmt.Errorf(i18n.T("creating the backup directory: %w"), err)
	}
	orig, err := m.Original()
	if err != nil {
		return err
	}
	if orig != nil {
		return nil
	}
	content, err := os.ReadFile(motdPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf(i18n.T("reading %s: %w"), motdPath, err)
	}
	_, err = m.save(content, prefixOriginal, Backup{IsOriginal: true, Note: "original"})
	return err
}

// Snapshot stores content (the config of an install) in the history.
func (m *Manager) Snapshot(content []byte, note string, modules []string) (*Backup, error) {
	b, err := m.save(content, prefixSnapshot, Backup{Note: note, Modules: modules})
	if err != nil {
		return nil, err
	}
	m.prune()
	return b, nil
}

func (m *Manager) save(content []byte, prefix string, meta Backup) (*Backup, error) {
	meta.Timestamp = time.Now()
	meta.Path = filepath.Join(m.dir, prefix+meta.Timestamp.Format("20060102_150405.000000000"))
	if err := os.WriteFile(meta.Path, content, 0o644); err != nil {
		return nil, fmt.Errorf(i18n.T("saving the backup: %w"), err)
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(meta.Path+metaSuffix, data, 0o644); err != nil {
		return nil, fmt.Errorf(i18n.T("saving the backup metadata: %w"), err)
	}
	return &meta, nil
}

// List returns every backup, newest first.
func (m *Manager) List() ([]Backup, error) {
	entries, err := os.ReadDir(m.dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf(i18n.T("reading the backup directory: %w"), err)
	}

	var out []Backup
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || strings.HasSuffix(name, metaSuffix) {
			continue
		}
		isOriginal := strings.HasPrefix(name, prefixOriginal)
		if !isOriginal && !strings.HasPrefix(name, prefixSnapshot) {
			continue
		}
		out = append(out, m.load(filepath.Join(m.dir, name), isOriginal))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Timestamp.After(out[j].Timestamp) })
	return out, nil
}

// load reads a backup's metadata, falling back to what the file itself
// tells us when the .meta sidecar is missing or unreadable.
func (m *Manager) load(path string, isOriginal bool) Backup {
	if data, err := os.ReadFile(path + metaSuffix); err == nil {
		var b Backup
		if json.Unmarshal(data, &b) == nil {
			b.Path = path
			return b
		}
	}
	b := Backup{Path: path, IsOriginal: isOriginal}
	if info, err := os.Stat(path); err == nil {
		b.Timestamp = info.ModTime()
	}
	return b
}

// Original returns the immutable original backup, or nil if none exists.
func (m *Manager) Original() (*Backup, error) {
	all, err := m.List()
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].IsOriginal {
			return &all[i], nil
		}
	}
	return nil, nil
}

// Read returns the content of a backup.
func (m *Manager) Read(b *Backup) ([]byte, error) {
	data, err := os.ReadFile(b.Path)
	if err != nil {
		return nil, fmt.Errorf(i18n.T("reading the backup: %w"), err)
	}
	return data, nil
}

// prune removes the oldest regular snapshots beyond maxSnapshots.
func (m *Manager) prune() {
	all, err := m.List()
	if err != nil {
		return
	}
	kept := 0
	for _, b := range all {
		if b.IsOriginal {
			continue
		}
		kept++
		if kept > maxSnapshots {
			os.Remove(b.Path)
			os.Remove(b.Path + metaSuffix)
		}
	}
}
