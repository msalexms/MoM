package module

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/msalexms/MoM/internal/i18n"

	"github.com/msalexms/MoM/internal/config"
)

// Configurable is implemented by modules that expose user settings. Each
// setting is bound to its config field, so the TUI can read and edit any
// module generically without knowing the config layout.
type Configurable interface {
	Settings() []Setting
}

// SettingType identifies the kind of value a setting holds and therefore
// how the TUI edits it.
type SettingType int

const (
	SettingBool   SettingType = iota // toggled in place
	SettingString                    // free text
	SettingEnum                      // cycles through Options
	SettingList                      // list of strings
	SettingInt                       // positive integer
)

// Setting describes one user-editable value of a module.
type Setting struct {
	Key         string // config key, as it appears in config.toml
	Label       string // human-readable label
	Description string // one-line help text
	Type        SettingType
	Options     []string // valid choices for SettingEnum

	get func(*config.Config) string
	set func(*config.Config, string) error
}

// Get returns the current value rendered as text (lists are comma-separated).
func (s Setting) Get(cfg *config.Config) string { return s.get(cfg) }

// Set parses text and stores it in cfg.
func (s Setting) Set(cfg *config.Config, value string) error { return s.set(cfg, value) }

// Values returns a list setting's items. It is empty for other types.
func (s Setting) Values(cfg *config.Config) []string {
	if s.Type != SettingList {
		return nil
	}
	return splitList(s.get(cfg))
}

// Next returns the value that follows the current one: the opposite for a
// bool and the next option (wrapping around) for an enum.
func (s Setting) Next(cfg *config.Config) string {
	cur := s.get(cfg)
	switch s.Type {
	case SettingBool:
		return strconv.FormatBool(cur != "true")
	case SettingEnum:
		i := slices.Index(s.Options, cur)
		return s.Options[(i+1)%len(s.Options)]
	}
	return cur
}

func stringSetting(key, label, desc string, field func(*config.Config) *string) Setting {
	return Setting{
		Key: key, Label: label, Description: desc, Type: SettingString,
		get: func(c *config.Config) string { return *field(c) },
		set: func(c *config.Config, v string) error {
			*field(c) = strings.TrimSpace(v)
			return nil
		},
	}
}

func enumSetting(key, label, desc string, options []string, field func(*config.Config) *string) Setting {
	return Setting{
		Key: key, Label: label, Description: desc, Type: SettingEnum, Options: options,
		get: func(c *config.Config) string { return *field(c) },
		set: func(c *config.Config, v string) error {
			if !slices.Contains(options, v) {
				return fmt.Errorf(i18n.T("invalid value %q (options: %s)"), v, strings.Join(options, ", "))
			}
			*field(c) = v
			return nil
		},
	}
}

func boolSetting(key, label, desc string, field func(*config.Config) *bool) Setting {
	return Setting{
		Key: key, Label: label, Description: desc, Type: SettingBool,
		get: func(c *config.Config) string { return strconv.FormatBool(*field(c)) },
		set: func(c *config.Config, v string) error {
			b, err := strconv.ParseBool(strings.TrimSpace(v))
			if err != nil {
				return fmt.Errorf(i18n.T("expected true or false, got %q"), v)
			}
			*field(c) = b
			return nil
		},
	}
}

func intSetting(key, label, desc string, field func(*config.Config) *int) Setting {
	return Setting{
		Key: key, Label: label, Description: desc, Type: SettingInt,
		get: func(c *config.Config) string { return strconv.Itoa(*field(c)) },
		set: func(c *config.Config, v string) error {
			n, err := strconv.Atoi(strings.TrimSpace(v))
			if err != nil || n <= 0 {
				return fmt.Errorf(i18n.T("expected a positive number, got %q"), v)
			}
			*field(c) = n
			return nil
		},
	}
}

func listSetting(key, label, desc string, field func(*config.Config) *[]string) Setting {
	return Setting{
		Key: key, Label: label, Description: desc, Type: SettingList,
		get: func(c *config.Config) string { return strings.Join(*field(c), ", ") },
		set: func(c *config.Config, v string) error {
			*field(c) = splitList(v)
			return nil
		},
	}
}

// splitList parses a comma-separated list, dropping blanks and duplicates.
func splitList(s string) []string {
	var out []string
	for _, item := range strings.Split(s, ",") {
		item = strings.TrimSpace(item)
		if item != "" && !slices.Contains(out, item) {
			out = append(out, item)
		}
	}
	return out
}
