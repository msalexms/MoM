package module

import (
	"context"

	"github.com/ams/mom/internal/module/render"
)

// Configurable is an optional interface that modules may implement to expose
// per-module settings to the TUI and config system. Modules that don't
// implement it are treated as having no configurable settings.
type Configurable interface {
	// Variants returns the list of render variants this module supports.
	// Must include at least render.VariantDefault.
	Variants() []render.Variant

	// DefaultVariant returns the variant used when the user hasn't chosen one.
	DefaultVariant() render.Variant

	// Settings returns the list of user-configurable settings for this module.
	// The TUI uses this to dynamically build a settings form.
	Settings() []SettingDef
}

// SettingType identifies the kind of value a setting holds.
type SettingType int

const (
	SettingBool   SettingType = iota // toggle
	SettingString                    // free text
	SettingEnum                      // pick one from a list
	SettingList                      // pick many from a list ([]string)
	SettingInt                       // numeric
)

// SettingDef describes a single configurable setting for a module.
type SettingDef struct {
	Key         string // config key (e.g. "city", "runtime")
	Label       string // human-readable label for the TUI
	Description string // one-line help text
	Type        SettingType
	Default     any      // default value (same type as the setting)
	Options     []string // valid choices for SettingEnum / SettingList
}

// Themeable is an optional interface that modules may implement to accept
// render options (theme + variant) at generation time. Modules that don't
// implement it use the default theme and variant.
type Themeable interface {
	// GenerateThemed is like Generate but receives render options so the
	// module can adapt its output to the active theme and variant.
	GenerateThemed(ctx context.Context, opts render.Options) (string, error)
}
