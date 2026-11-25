package generator

import (
	"strings"
	"testing"

	"github.com/msalexms/MoM/internal/distro"
)

func TestHookRendersLiveMotd(t *testing.T) {
	w := &Writer{exe: "/opt/it's/mom"}

	w.family = distro.FamilyDebian
	path, script, mode := w.hook()
	if path != debianHook || mode != 0o755 || !strings.HasPrefix(script, "#!/bin/sh\n") {
		t.Errorf("debian hook: %s %o\n%s", path, mode, script)
	}
	if !strings.Contains(script, `exec '/opt/it'\''s/mom' --motd`) {
		t.Errorf("debian hook must exec mom --motd with a quoted path:\n%s", script)
	}

	w.family = distro.FamilyArch
	path, script, _ = w.hook()
	if path != profileHook || !strings.Contains(script, "case $- in *i*)") || !strings.Contains(script, "--motd") {
		t.Errorf("profile hook must only run in interactive shells:\n%s", script)
	}
}
