package backup

import (
	"context"
	"fmt"
)

// RollbackList returns the list of backups available for rollback.
// The TUI handles selection; this just provides the data.
func (m *Manager) RollbackList() ([]Backup, error) {
	return m.List()
}

// RollbackToOriginal restores the original MOTD backup.
func (m *Manager) RollbackToOriginal(ctx context.Context, motdPath string) error {
	original, err := m.GetOriginal()
	if err != nil {
		return err
	}
	if original == nil {
		return fmt.Errorf("no original backup found")
	}
	return m.Restore(ctx, original, motdPath)
}
