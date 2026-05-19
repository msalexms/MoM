package distro

import (
	"os"
	"path/filepath"
	"testing"
)

func createTempOsRelease(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "os-release")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp os-release: %v", err)
	}
	return path
}

func TestDetect_Ubuntu(t *testing.T) {
	path := createTempOsRelease(t, `NAME="Ubuntu"
PRETTY_NAME="Ubuntu 24.04.1 LTS"
VERSION_ID="24.04"
ID=ubuntu
ID_LIKE=debian
`)
	info, err := DetectFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Family != FamilyDebian {
		t.Errorf("expected FamilyDebian, got %q", info.Family)
	}
	if info.ID != "ubuntu" {
		t.Errorf("expected ID 'ubuntu', got %q", info.ID)
	}
	if info.Version != "24.04" {
		t.Errorf("expected version '24.04', got %q", info.Version)
	}
}

func TestDetect_Fedora(t *testing.T) {
	path := createTempOsRelease(t, `NAME="Fedora Linux"
PRETTY_NAME="Fedora Linux 40 (Workstation Edition)"
VERSION_ID="40"
ID=fedora
`)
	info, err := DetectFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Family != FamilyRHEL {
		t.Errorf("expected FamilyRHEL, got %q", info.Family)
	}
	if info.ID != "fedora" {
		t.Errorf("expected ID 'fedora', got %q", info.ID)
	}
}

func TestDetect_Arch(t *testing.T) {
	path := createTempOsRelease(t, `NAME="Arch Linux"
PRETTY_NAME="Arch Linux"
ID=arch
`)
	info, err := DetectFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Family != FamilyArch {
		t.Errorf("expected FamilyArch, got %q", info.Family)
	}
}

func TestDetect_OpenSUSE(t *testing.T) {
	path := createTempOsRelease(t, `NAME="openSUSE Tumbleweed"
PRETTY_NAME="openSUSE Tumbleweed"
ID=opensuse-tumbleweed
ID_LIKE="opensuse suse"
VERSION_ID="20240501"
`)
	info, err := DetectFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Family != FamilySUSE {
		t.Errorf("expected FamilySUSE, got %q", info.Family)
	}
}

func TestDetect_Unknown(t *testing.T) {
	// Non-existent file
	info, err := DetectFrom("/tmp/nonexistent-os-release-file-mom-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Family != FamilyUnknown {
		t.Errorf("expected FamilyUnknown, got %q", info.Family)
	}
}

func TestDetect_IDLike(t *testing.T) {
	// Linux Mint has ID_LIKE=ubuntu debian
	path := createTempOsRelease(t, `NAME="Linux Mint"
PRETTY_NAME="Linux Mint 21.3"
VERSION_ID="21.3"
ID=linuxmint
ID_LIKE="ubuntu debian"
`)
	info, err := DetectFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Family != FamilyDebian {
		t.Errorf("expected FamilyDebian, got %q", info.Family)
	}
}

func TestDetect_RockyFromIDLike(t *testing.T) {
	// Rocky Linux has ID_LIKE with rhel
	path := createTempOsRelease(t, `NAME="Rocky Linux"
PRETTY_NAME="Rocky Linux 9.3 (Blue Onyx)"
VERSION_ID="9.3"
ID=rocky
ID_LIKE="rhel centos fedora"
`)
	info, err := DetectFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Family != FamilyRHEL {
		t.Errorf("expected FamilyRHEL, got %q", info.Family)
	}
}

func TestGetPaths_Debian(t *testing.T) {
	paths := GetPaths(FamilyDebian)
	if paths.MotdDir != "/etc/update-motd.d/" {
		t.Errorf("expected /etc/update-motd.d/, got %q", paths.MotdDir)
	}
	if paths.ProfileScript != "" {
		t.Errorf("expected empty ProfileScript for Debian, got %q", paths.ProfileScript)
	}
	if paths.ScriptName != "99-mom" {
		t.Errorf("expected ScriptName '99-mom', got %q", paths.ScriptName)
	}
}

func TestGetPaths_Arch(t *testing.T) {
	paths := GetPaths(FamilyArch)
	if paths.MotdDir != "" {
		t.Errorf("expected empty MotdDir for Arch, got %q", paths.MotdDir)
	}
	if paths.ProfileScript != "/etc/profile.d/mom-motd.sh" {
		t.Errorf("expected /etc/profile.d/mom-motd.sh, got %q", paths.ProfileScript)
	}
}

func TestGetPaths_RHEL(t *testing.T) {
	paths := GetPaths(FamilyRHEL)
	if paths.MotdDir != "/etc/motd.d/" {
		t.Errorf("expected /etc/motd.d/, got %q", paths.MotdDir)
	}
	if paths.ProfileScript != "/etc/profile.d/mom-motd.sh" {
		t.Errorf("expected /etc/profile.d/mom-motd.sh, got %q", paths.ProfileScript)
	}
}
