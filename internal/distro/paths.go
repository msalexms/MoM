package distro

// MotdPaths holds the filesystem paths where MOTD content should be written
// depending on the Linux distribution family.
type MotdPaths struct {
	// MotdFile is the path to the static MOTD file (e.g. /etc/motd).
	MotdFile string

	// MotdDir is the directory for dynamic MOTD scripts (e.g. /etc/update-motd.d/).
	// Empty string means the distro does not use this mechanism.
	MotdDir string

	// ProfileScript is the path to a profile.d script for displaying MOTD on login.
	// Empty string means the distro uses another mechanism (e.g. pam_motd with update-motd.d).
	ProfileScript string

	// ScriptName is the filename of the script mom creates in MotdDir.
	ScriptName string
}

// GetPaths returns the appropriate MOTD paths for the given distribution family.
func GetPaths(family Family) MotdPaths {
	switch family {
	case FamilyDebian:
		return MotdPaths{
			MotdFile:      "/etc/motd",
			MotdDir:       "/etc/update-motd.d/",
			ProfileScript: "",
			ScriptName:    "99-mom",
		}
	case FamilyRHEL:
		return MotdPaths{
			MotdFile:      "/etc/motd",
			MotdDir:       "/etc/motd.d/",
			ProfileScript: "/etc/profile.d/mom-motd.sh",
			ScriptName:    "mom.sh",
		}
	case FamilyArch:
		return MotdPaths{
			MotdFile:      "/etc/motd",
			MotdDir:       "",
			ProfileScript: "/etc/profile.d/mom-motd.sh",
			ScriptName:    "",
		}
	case FamilySUSE:
		return MotdPaths{
			MotdFile:      "/etc/motd",
			MotdDir:       "",
			ProfileScript: "/etc/profile.d/mom-motd.sh",
			ScriptName:    "",
		}
	default: // FamilyUnknown
		return MotdPaths{
			MotdFile:      "/etc/motd",
			MotdDir:       "",
			ProfileScript: "/etc/profile.d/mom-motd.sh",
			ScriptName:    "",
		}
	}
}
