// Package distro provides Linux distribution detection and MOTD path resolution.
package distro

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Family represents a Linux distribution family.
type Family string

const (
	FamilyDebian  Family = "debian"
	FamilyRHEL    Family = "rhel"
	FamilyArch    Family = "arch"
	FamilySUSE    Family = "suse"
	FamilyUnknown Family = "unknown"
)

// Info holds detected distribution information.
type Info struct {
	Family  Family
	Name    string // e.g. "Ubuntu 24.04"
	Version string // e.g. "24.04"
	ID      string // e.g. "ubuntu"
}

// osReleasePath is the default path to os-release. Overridable for testing.
var osReleasePath = "/etc/os-release"

// debianIDs maps distribution IDs to the Debian family.
var debianIDs = map[string]bool{
	"debian":     true,
	"ubuntu":     true,
	"linuxmint":  true,
	"pop":        true,
	"elementary": true,
	"zorin":      true,
	"kali":       true,
	"raspbian":   true,
}

// rhelIDs maps distribution IDs to the RHEL family.
var rhelIDs = map[string]bool{
	"rhel":      true,
	"fedora":    true,
	"centos":    true,
	"rocky":     true,
	"almalinux": true,
	"ol":        true,
}

// archIDs maps distribution IDs to the Arch family.
var archIDs = map[string]bool{
	"arch":        true,
	"manjaro":     true,
	"endeavouros": true,
	"garuda":      true,
	"artix":       true,
}

// suseIDs maps distribution IDs to the SUSE family.
var suseIDs = map[string]bool{
	"opensuse":        true,
	"opensuse-leap":   true,
	"opensuse-tumbleweed": true,
	"suse":            true,
	"sles":            true,
}

// Detect reads /etc/os-release and determines the Linux distribution family.
func Detect() (Info, error) {
	return DetectFrom(osReleasePath)
}

// DetectFrom reads a given os-release file and determines the distribution family.
// This allows testing with custom files.
func DetectFrom(path string) (Info, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Info{Family: FamilyUnknown}, nil
		}
		return Info{}, fmt.Errorf("opening os-release: %w", err)
	}
	defer f.Close()

	fields := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := strings.Trim(parts[1], "\"")
		fields[key] = value
	}
	if err := scanner.Err(); err != nil {
		return Info{}, fmt.Errorf("reading os-release: %w", err)
	}

	id := strings.ToLower(fields["ID"])
	idLike := strings.ToLower(fields["ID_LIKE"])
	name := fields["PRETTY_NAME"]
	if name == "" {
		name = fields["NAME"]
	}
	version := fields["VERSION_ID"]

	info := Info{
		ID:      id,
		Name:    name,
		Version: version,
	}

	// Determine family from ID first
	info.Family = resolveFamily(id)
	if info.Family != FamilyUnknown {
		return info, nil
	}

	// Try ID_LIKE (space-separated list)
	for _, like := range strings.Fields(idLike) {
		family := resolveFamily(like)
		if family != FamilyUnknown {
			info.Family = family
			return info, nil
		}
	}

	return info, nil
}

// resolveFamily maps a single distribution ID to its family.
func resolveFamily(id string) Family {
	if debianIDs[id] {
		return FamilyDebian
	}
	if rhelIDs[id] {
		return FamilyRHEL
	}
	if archIDs[id] {
		return FamilyArch
	}
	if suseIDs[id] {
		return FamilySUSE
	}
	return FamilyUnknown
}
