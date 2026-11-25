// Package version holds the build metadata injected via -ldflags.
// The same values are used by --version output and template exports.
package version

// These are set at build time:
//
//	-X github.com/msalexms/MoM/internal/version.Version=1.2.3
var (
	Version = "dev"
	Commit  = "dev"
	Date    = "unknown"
)
