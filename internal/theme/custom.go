package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// themeFile is the TOML structure for a custom theme file.
type themeFile struct {
	ID          string `toml:"id"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	UseUnicode  *bool  `toml:"use_unicode"`

	Colors struct {
		Foreground string `toml:"foreground"`
		Muted      string `toml:"muted"`
		Subtle     string `toml:"subtle"`
		Accent     string `toml:"accent"`
		Secondary  string `toml:"secondary"`
		Success    string `toml:"success"`
		Warning    string `toml:"warning"`
		Danger     string `toml:"danger"`
		Info       string `toml:"info"`
	} `toml:"colors"`

	Gradient struct {
		Low  string `toml:"low"`
		Mid  string `toml:"mid"`
		High string `toml:"high"`
	} `toml:"gradient"`
}

// LoadCustomThemes reads all .toml files from ~/.config/mom/themes/ and
// registers them. Existing built-in themes are not overwritten.
func LoadCustomThemes() error {
	dir := customThemesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading themes dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".toml") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		t, err := loadThemeFile(path)
		if err != nil {
			continue // skip invalid themes silently
		}
		// Don't overwrite built-in themes
		if _, exists := registry[t.ID]; !exists {
			Register(t)
		}
	}
	return nil
}

func loadThemeFile(path string) (*Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tf themeFile
	if _, err := toml.Decode(string(data), &tf); err != nil {
		return nil, fmt.Errorf("parsing theme %s: %w", path, err)
	}

	if tf.ID == "" || tf.Name == "" {
		return nil, fmt.Errorf("theme missing id or name: %s", path)
	}

	unicode := true
	if tf.UseUnicode != nil {
		unicode = *tf.UseUnicode
	}

	t := &Theme{
		ID:          tf.ID,
		Name:        tf.Name,
		Description: tf.Description,
		UseUnicode:  unicode,
		Attrs: Attrs{
			Bold:      "\033[1m",
			Dim:       "\033[2m",
			Italic:    "\033[3m",
			Underline: "\033[4m",
		},
		Palette: Palette{
			Foreground:  parseColor(tf.Colors.Foreground),
			Muted:       parseColor(tf.Colors.Muted),
			Subtle:      parseColor(tf.Colors.Subtle),
			Accent:      parseColor(tf.Colors.Accent),
			Secondary:   parseColor(tf.Colors.Secondary),
			Success:     parseColor(tf.Colors.Success),
			Warning:     parseColor(tf.Colors.Warning),
			Danger:      parseColor(tf.Colors.Danger),
			Info:        parseColor(tf.Colors.Info),
			GradientLow: parseColor(tf.Gradient.Low),
			GradientMid: parseColor(tf.Gradient.Mid),
			GradientHigh: parseColor(tf.Gradient.High),
		},
	}

	// Fill section colors from accent/secondary
	t.Palette.SectionSystem = t.Palette.Accent
	t.Palette.SectionResources = t.Palette.Accent
	t.Palette.SectionNetwork = t.Palette.Info
	t.Palette.SectionWeather = t.Palette.Warning
	t.Palette.SectionContainer = t.Palette.Info
	t.Palette.SectionService = t.Palette.Secondary
	t.Palette.SectionUpdate = t.Palette.Warning
	t.Palette.SectionLogin = t.Palette.Success
	t.Palette.SectionCalendar = t.Palette.Accent
	t.Palette.SectionQuote = t.Palette.Secondary
	t.Palette.SectionArt = t.Palette.Accent
	t.Palette.SectionLogo = t.Palette.Foreground

	// Default gradient if not specified
	if t.Palette.GradientLow == "" {
		t.Palette.GradientLow = t.Palette.Success
	}
	if t.Palette.GradientMid == "" {
		t.Palette.GradientMid = t.Palette.Warning
	}
	if t.Palette.GradientHigh == "" {
		t.Palette.GradientHigh = t.Palette.Danger
	}

	return t, nil
}

// parseColor converts a hex color (#RRGGBB) or ANSI code to an escape sequence.
func parseColor(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// Already an escape sequence
	if strings.HasPrefix(s, "\033[") {
		return s
	}
	// Hex color #RRGGBB
	if strings.HasPrefix(s, "#") && len(s) == 7 {
		var r, g, b int
		fmt.Sscanf(s[1:], "%02x%02x%02x", &r, &g, &b)
		return fgRGB(r, g, b)
	}
	return ""
}

func customThemesDir() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "mom", "themes")
}
