package backup

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupTestManager(t *testing.T) (*Manager, string) {
	t.Helper()
	dir := t.TempDir()
	backupDir := filepath.Join(dir, "backups")
	motdFile := filepath.Join(dir, "motd")
	os.WriteFile(motdFile, []byte("Original MOTD content"), 0644)
	return NewManager(backupDir), motdFile
}

func TestBackup_Init_CreatesDir(t *testing.T) {
	mgr, motdFile := setupTestManager(t)

	err := mgr.Init(context.Background(), motdFile)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	info, err := os.Stat(mgr.BackupDir)
	if err != nil {
		t.Fatalf("backup dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("backup path is not a directory")
	}
}

func TestBackup_Init_SavesOriginal(t *testing.T) {
	mgr, motdFile := setupTestManager(t)

	err := mgr.Init(context.Background(), motdFile)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	original, err := mgr.GetOriginal()
	if err != nil {
		t.Fatalf("GetOriginal failed: %v", err)
	}
	if original == nil {
		t.Fatal("expected original backup, got nil")
	}
	if !original.IsOriginal {
		t.Error("expected IsOriginal=true")
	}

	content, err := mgr.GetContent(original)
	if err != nil {
		t.Fatalf("GetContent failed: %v", err)
	}
	if content != "Original MOTD content" {
		t.Errorf("expected 'Original MOTD content', got %q", content)
	}
}

func TestBackup_Init_DoesNotOverwriteOriginal(t *testing.T) {
	mgr, motdFile := setupTestManager(t)

	// First init
	err := mgr.Init(context.Background(), motdFile)
	if err != nil {
		t.Fatalf("first Init failed: %v", err)
	}

	// Modify MOTD
	os.WriteFile(motdFile, []byte("Modified content"), 0644)

	// Second init should NOT overwrite the original
	err = mgr.Init(context.Background(), motdFile)
	if err != nil {
		t.Fatalf("second Init failed: %v", err)
	}

	original, _ := mgr.GetOriginal()
	content, _ := mgr.GetContent(original)
	if content != "Original MOTD content" {
		t.Errorf("original was overwritten, got %q", content)
	}
}

func TestBackup_Backup_SavesCorrectly(t *testing.T) {
	mgr, motdFile := setupTestManager(t)
	os.MkdirAll(mgr.BackupDir, 0755)

	backup, err := mgr.Backup(context.Background(), motdFile, "ubuntu", []string{"system", "logo"})
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	if backup.Distro != "ubuntu" {
		t.Errorf("expected distro 'ubuntu', got %q", backup.Distro)
	}
	if len(backup.Modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(backup.Modules))
	}

	content, err := mgr.GetContent(backup)
	if err != nil {
		t.Fatalf("GetContent failed: %v", err)
	}
	if content != "Original MOTD content" {
		t.Errorf("backup content mismatch: %q", content)
	}
}

func TestBackup_Backup_MultipleBackups(t *testing.T) {
	mgr, motdFile := setupTestManager(t)
	os.MkdirAll(mgr.BackupDir, 0755)

	// Create multiple backups with slight delay to ensure different timestamps
	_, err := mgr.Backup(context.Background(), motdFile, "ubuntu", []string{"system"})
	if err != nil {
		t.Fatalf("first backup failed: %v", err)
	}

	time.Sleep(10 * time.Millisecond)
	os.WriteFile(motdFile, []byte("Second version"), 0644)

	_, err = mgr.Backup(context.Background(), motdFile, "ubuntu", []string{"system", "logo"})
	if err != nil {
		t.Fatalf("second backup failed: %v", err)
	}

	backups, err := mgr.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(backups) < 2 {
		t.Errorf("expected at least 2 backups, got %d", len(backups))
	}
}

func TestBackup_List_ReturnsSorted(t *testing.T) {
	mgr, motdFile := setupTestManager(t)
	os.MkdirAll(mgr.BackupDir, 0755)

	_, _ = mgr.Backup(context.Background(), motdFile, "arch", nil)
	time.Sleep(10 * time.Millisecond)
	_, _ = mgr.Backup(context.Background(), motdFile, "fedora", nil)

	backups, err := mgr.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(backups) < 2 {
		t.Fatalf("expected at least 2 backups, got %d", len(backups))
	}

	// Most recent first
	if backups[0].Timestamp.Before(backups[1].Timestamp) {
		t.Error("expected backups sorted most recent first")
	}
}

func TestBackup_Restore_RestoresContent(t *testing.T) {
	mgr, motdFile := setupTestManager(t)
	os.MkdirAll(mgr.BackupDir, 0755)

	// Backup original
	backup, _ := mgr.Backup(context.Background(), motdFile, "ubuntu", nil)

	// Modify MOTD
	os.WriteFile(motdFile, []byte("New MOTD content"), 0644)

	// Restore
	err := mgr.Restore(context.Background(), backup, motdFile)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	content, _ := os.ReadFile(motdFile)
	if string(content) != "Original MOTD content" {
		t.Errorf("expected restored content 'Original MOTD content', got %q", string(content))
	}
}

func TestBackup_Original_Immutable(t *testing.T) {
	mgr, motdFile := setupTestManager(t)
	mgr.Init(context.Background(), motdFile)

	original, _ := mgr.GetOriginal()
	if original == nil {
		t.Fatal("expected original backup")
	}

	// Verify original file still has original content after multiple operations
	os.WriteFile(motdFile, []byte("Changed"), 0644)
	mgr.Backup(context.Background(), motdFile, "test", nil)

	content, _ := mgr.GetContent(original)
	if content != "Original MOTD content" {
		t.Errorf("original backup was modified, content: %q", content)
	}
}

func TestBackup_RollbackToOriginal(t *testing.T) {
	mgr, motdFile := setupTestManager(t)
	mgr.Init(context.Background(), motdFile)

	// Modify MOTD
	os.WriteFile(motdFile, []byte("Modified"), 0644)

	err := mgr.RollbackToOriginal(context.Background(), motdFile)
	if err != nil {
		t.Fatalf("RollbackToOriginal failed: %v", err)
	}

	content, _ := os.ReadFile(motdFile)
	if string(content) != "Original MOTD content" {
		t.Errorf("expected original content, got %q", string(content))
	}
}

func TestBackup_NonExistentMotd(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(filepath.Join(dir, "backups"))
	nonExistentMotd := filepath.Join(dir, "no-such-file")

	err := mgr.Init(context.Background(), nonExistentMotd)
	if err != nil {
		t.Fatalf("Init with non-existent MOTD should not fail: %v", err)
	}

	original, _ := mgr.GetOriginal()
	content, _ := mgr.GetContent(original)
	if content != "" {
		t.Errorf("expected empty content for non-existent MOTD, got %q", content)
	}
}

func TestBackup_MetadataFile(t *testing.T) {
	mgr, motdFile := setupTestManager(t)
	os.MkdirAll(mgr.BackupDir, 0755)

	backup, _ := mgr.Backup(context.Background(), motdFile, "fedora", []string{"system", "weather"})

	// Verify .meta file exists
	metaPath := backup.Path + ".meta"
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Error("expected .meta file to exist")
	}

	// Read back the metadata
	meta, err := mgr.readMeta(backup.Path)
	if err != nil {
		t.Fatalf("readMeta failed: %v", err)
	}
	if meta.Distro != "fedora" {
		t.Errorf("expected distro 'fedora', got %q", meta.Distro)
	}
	if len(meta.Modules) != 2 || meta.Modules[0] != "system" {
		t.Errorf("modules mismatch: %v", meta.Modules)
	}
}

func TestBackup_ListFiltersMetaFiles(t *testing.T) {
	mgr, motdFile := setupTestManager(t)
	os.MkdirAll(mgr.BackupDir, 0755)

	mgr.Backup(context.Background(), motdFile, "ubuntu", nil)

	backups, _ := mgr.List()
	for _, b := range backups {
		if strings.HasSuffix(b.Path, ".meta") {
			t.Error("List should not include .meta files")
		}
	}
}
