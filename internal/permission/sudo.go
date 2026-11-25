// Package permission writes root-owned files through sudo.
//
// Every privileged command runs with "sudo -n" (never prompts) and its
// output is captured, so nothing ever writes to a terminal that a TUI
// owns. Callers authenticate once up front: the CLI with Authenticate, the
// TUI by running AuthCommand through tea.ExecProcess, which hands sudo the
// real terminal for the password prompt.
package permission

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/msalexms/MoM/internal/i18n"
)

// IsRoot reports whether the process already runs as root.
func IsRoot() bool { return os.Geteuid() == 0 }

// NeedsAuth reports whether a password prompt is required before the
// privileged commands can run non-interactively.
func NeedsAuth(ctx context.Context) bool {
	if IsRoot() {
		return false
	}
	return exec.CommandContext(ctx, "sudo", "-n", "true").Run() != nil
}

// AuthCommand returns the command that asks for the sudo password and
// caches the credentials ("sudo -v").
func AuthCommand() *exec.Cmd {
	return exec.Command("sudo", "-v", "-p", i18n.T("[sudo] password for %u to write the MOTD: "))
}

// Authenticate runs AuthCommand attached to the current terminal. It is
// meant for the headless CLI paths; it is a no-op when no prompt is needed.
func Authenticate(ctx context.Context) error {
	if !NeedsAuth(ctx) {
		return nil
	}
	cmd := AuthCommand()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(i18n.T("sudo: authentication failed: %w"), err)
	}
	return nil
}

// WriteFile atomically replaces path with content and the given mode.
func WriteFile(ctx context.Context, path, content string, perm os.FileMode) error {
	if IsRoot() {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		tmp := path + ".mom-tmp"
		if err := os.WriteFile(tmp, []byte(content), perm); err != nil {
			return err
		}
		return os.Rename(tmp, path)
	}

	tmp, err := os.CreateTemp("", "mom-*")
	if err != nil {
		return fmt.Errorf(i18n.T("creating the temporary file: %w"), err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		return fmt.Errorf(i18n.T("writing the temporary file: %w"), err)
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	// install creates the parent directory, copies, sets owner and mode
	// and replaces the target in one step.
	return sudo(ctx, "install", "-D", "-o", "root", "-g", "root", "-m", fmt.Sprintf("%o", perm), tmp.Name(), path)
}

// Remove deletes path if it exists.
func Remove(ctx context.Context, path string) error {
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return nil
	}
	if IsRoot() {
		return os.Remove(path)
	}
	return sudo(ctx, "rm", "-f", filepath.Clean(path))
}

// RemoveDirIfEmpty deletes dir when it exists and is empty.
func RemoveDirIfEmpty(ctx context.Context, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) > 0 {
		return nil
	}
	if IsRoot() {
		return os.Remove(dir)
	}
	return sudo(ctx, "rmdir", filepath.Clean(dir))
}

func sudo(ctx context.Context, args ...string) error {
	out, err := exec.CommandContext(ctx, "sudo", append([]string{"-n"}, args...)...).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("sudo %s: %s", args[0], msg)
	}
	return nil
}
