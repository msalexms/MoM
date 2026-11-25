package backup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setup(t *testing.T) (*Manager, string) {
	t.Helper()
	dir := t.TempDir()
	motd := filepath.Join(dir, "motd")
	if err := os.WriteFile(motd, []byte("Original MOTD content"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := NewManager(filepath.Join(dir, "backups"))
	if err := m.Init(motd); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return m, motd
}

func read(t *testing.T, m *Manager, b *Backup) string {
	t.Helper()
	s, err := m.Read(b)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	return string(s)
}

func TestInitSavesOriginalOnce(t *testing.T) {
	m, motd := setup(t)

	os.WriteFile(motd, []byte("Changed"), 0o644)
	if err := m.Init(motd); err != nil {
		t.Fatalf("second Init: %v", err)
	}

	orig, err := m.Original()
	if err != nil || orig == nil {
		t.Fatalf("Original: %v, %v", orig, err)
	}
	if got := read(t, m, orig); got != "Original MOTD content" {
		t.Errorf("original overwritten: %q", got)
	}
}

func TestInitWithMissingMotd(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(filepath.Join(dir, "backups"))
	if err := m.Init(filepath.Join(dir, "missing")); err != nil {
		t.Fatalf("Init: %v", err)
	}
	orig, _ := m.Original()
	if got := read(t, m, orig); got != "" {
		t.Errorf("expected empty original, got %q", got)
	}
}

func TestSnapshotKeepsContentAndMetadata(t *testing.T) {
	m, _ := setup(t)

	b, err := m.Snapshot([]byte("v2"), "antes de guardar", []string{"system", "weather"})
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	if got := read(t, m, b); got != "v2" {
		t.Errorf("content = %q", got)
	}

	all, _ := m.List()
	var found *Backup
	for i := range all {
		if all[i].Path == b.Path {
			found = &all[i]
		}
		if strings.HasSuffix(all[i].Path, metaSuffix) {
			t.Error("List returned a .meta file")
		}
	}
	if found == nil {
		t.Fatal("snapshot not listed")
	}
	if found.Note != "antes de guardar" || len(found.Modules) != 2 || found.IsOriginal {
		t.Errorf("metadata mismatch: %+v", found)
	}
}

func TestListNewestFirst(t *testing.T) {
	m, _ := setup(t)
	m.Snapshot([]byte("a"), "a", nil)
	m.Snapshot([]byte("b"), "b", nil)

	all, _ := m.List()
	if len(all) != 3 {
		t.Fatalf("expected 3 backups, got %d", len(all))
	}
	for i := 1; i < len(all); i++ {
		if all[i-1].Timestamp.Before(all[i].Timestamp) {
			t.Fatal("backups not sorted newest first")
		}
	}
}

func TestPruneKeepsOriginal(t *testing.T) {
	m, _ := setup(t)
	for range maxSnapshots + 5 {
		if _, err := m.Snapshot([]byte("x"), "", nil); err != nil {
			t.Fatal(err)
		}
	}

	all, _ := m.List()
	if len(all) != maxSnapshots+1 {
		t.Errorf("expected %d backups after pruning, got %d", maxSnapshots+1, len(all))
	}
	if orig, _ := m.Original(); orig == nil {
		t.Error("original was pruned")
	}
}
