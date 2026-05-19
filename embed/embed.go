// Package embed provides embedded static assets for the mom binary.
package embeds

import "embed"

// LogosFS contains the embedded ASCII art logos.
//
//go:embed logos/*
var LogosFS embed.FS

// TemplatesFS contains the embedded template TOML files.
//
//go:embed templates/*
var TemplatesFS embed.FS
