// Package permission provides sudo elevation utilities for writing system files.
package permission

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
)

// IsRoot returns true if the current process is running as root.
func IsRoot() bool {
	return os.Geteuid() == 0
}

// CheckSudo verifies whether cached sudo credentials are available.
// Returns nil if sudo can be used without prompting for a password.
func CheckSudo(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "sudo", "-n", "true")
	return cmd.Run()
}

// WriteWithSudo writes content to a system path using sudo.
// It creates a temporary file and copies it with sudo to the target path.
func WriteWithSudo(ctx context.Context, content string, path string, perms os.FileMode) error {
	if IsRoot() {
		// Already root, write directly
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory: %w", err)
		}
		return os.WriteFile(path, []byte(content), perms)
	}

	// Create temporary file
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("mom-%d.tmp", rand.Int64()))
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	// Copy with sudo
	cmd := exec.CommandContext(ctx, "sudo", "cp", tmpFile, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sudo cp: %w", err)
	}

	// Set permissions
	cmd = exec.CommandContext(ctx, "sudo", "chmod", fmt.Sprintf("%o", perms), path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sudo chmod: %w", err)
	}

	return nil
}

// MkdirWithSudo creates a directory with sudo if needed.
func MkdirWithSudo(ctx context.Context, path string, perms os.FileMode) error {
	if IsRoot() {
		return os.MkdirAll(path, perms)
	}

	// Check if directory already exists
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return nil
	}

	cmd := exec.CommandContext(ctx, "sudo", "mkdir", "-p", path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sudo mkdir: %w", err)
	}

	cmd = exec.CommandContext(ctx, "sudo", "chmod", fmt.Sprintf("%o", perms), path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sudo chmod: %w", err)
	}

	return nil
}

// ChmodWithSudo sets file permissions with sudo.
func ChmodWithSudo(ctx context.Context, path string, perms os.FileMode) error {
	if IsRoot() {
		return os.Chmod(path, perms)
	}

	cmd := exec.CommandContext(ctx, "sudo", "chmod", fmt.Sprintf("%o", perms), path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RemoveWithSudo removes a file with sudo.
func RemoveWithSudo(ctx context.Context, path string) error {
	if IsRoot() {
		return os.Remove(path)
	}

	cmd := exec.CommandContext(ctx, "sudo", "rm", "-f", path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
