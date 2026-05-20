package theme

// Built-in themes. Each theme is registered via init() so they are available
// from any package as soon as `internal/theme` is imported. New themes can be
// added by creating a new file in this package and calling Register() in init().

// ANSI 16-color SGR codes used as building blocks.
const (
	ansiBlack         = "\033[30m"
	ansiRed           = "\033[31m"
	ansiGreen         = "\033[32m"
	ansiYellow        = "\033[33m"
	ansiBlue          = "\033[34m"
	ansiMagenta       = "\033[35m"
	ansiCyan          = "\033[36m"
	ansiWhite         = "\033[37m"
	ansiBrightBlack   = "\033[90m"
	ansiBrightRed     = "\033[91m"
	ansiBrightGreen   = "\033[92m"
	ansiBrightYellow  = "\033[93m"
	ansiBrightBlue    = "\033[94m"
	ansiBrightMagenta = "\033[95m"
	ansiBrightCyan    = "\033[96m"
	ansiBrightWhite   = "\033[97m"

	ansiBold      = "\033[1m"
	ansiDim       = "\033[2m"
	ansiItalic    = "\033[3m"
	ansiUnderline = "\033[4m"
)

// fg256 returns an ANSI 256-color foreground sequence for the given color
// index (0-255). Used by themes that target a specific palette.
func fg256(n int) string {
	return "\033[38;5;" + itoa(n) + "m"
}

// fgRGB returns a truecolor foreground sequence (24-bit). Most modern
// terminals support this.
func fgRGB(r, g, b int) string {
	return "\033[38;2;" + itoa(r) + ";" + itoa(g) + ";" + itoa(b) + "m"
}

// itoa is a tiny stack-allocating int-to-string for small positive numbers.
// Avoids pulling in strconv on package init().
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [4]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

func init() {
	Register(themeDefault())
	Register(themeDracula())
	Register(themeNord())
	Register(themeSolarizedDark())
	Register(themeMonochrome())
	Register(themeASCII())
}

// themeDefault is the original mom palette: bright 16-color ANSI. Looks good
// on most terminals out of the box and doesn't depend on truecolor support.
func themeDefault() *Theme {
	return &Theme{
		ID:          "default",
		Name:        "Default",
		Description: "Bright 16-color palette, works in any terminal",
		UseUnicode:  true,
		Attrs: Attrs{
			Bold: ansiBold, Dim: ansiDim, Italic: ansiItalic, Underline: ansiUnderline,
		},
		Palette: Palette{
			Foreground: ansiBrightWhite,
			Muted:      ansiWhite,
			Subtle:     ansiBrightBlack,
			Accent:     ansiBrightCyan,
			Secondary:  ansiBrightMagenta,
			Success:    ansiBrightGreen,
			Warning:    ansiBrightYellow,
			Danger:     ansiBrightRed,
			Info:       ansiBrightBlue,

			SectionSystem:    ansiBrightCyan,
			SectionResources: ansiBrightCyan,
			SectionNetwork:   ansiBrightBlue,
			SectionWeather:   ansiBrightYellow,
			SectionContainer: ansiBrightBlue,
			SectionService:   ansiBrightMagenta,
			SectionUpdate:    ansiBrightYellow,
			SectionLogin:     ansiBrightGreen,
			SectionCalendar:  ansiBrightCyan,
			SectionQuote:     ansiBrightMagenta,
			SectionArt:       ansiBrightCyan,
			SectionLogo:      ansiBrightWhite,

			GradientLow:  ansiBrightGreen,
			GradientMid:  ansiBrightYellow,
			GradientHigh: ansiBrightRed,
		},
	}
}

// themeDracula is the popular Dracula palette (https://draculatheme.com).
func themeDracula() *Theme {
	const (
		bg         = "" // we don't paint backgrounds
		foreground = "" // truecolor: f8f8f2
		comment    = ""
		cyan       = ""
		green      = ""
		orange     = ""
		pink       = ""
		purple     = ""
		red        = ""
		yellow     = ""
		_          = bg + foreground + comment + cyan + green + orange + pink + purple + red + yellow
	)
	return &Theme{
		ID:          "dracula",
		Name:        "Dracula",
		Description: "Dracula color scheme — vivid pink/purple on dark",
		UseUnicode:  true,
		Attrs: Attrs{
			Bold: ansiBold, Dim: ansiDim, Italic: ansiItalic, Underline: ansiUnderline,
		},
		Palette: Palette{
			Foreground: fgRGB(248, 248, 242),
			Muted:      fgRGB(189, 147, 249),
			Subtle:     fgRGB(98, 114, 164),

			Accent:    fgRGB(255, 121, 198),
			Secondary: fgRGB(189, 147, 249),
			Success:   fgRGB(80, 250, 123),
			Warning:   fgRGB(241, 250, 140),
			Danger:    fgRGB(255, 85, 85),
			Info:      fgRGB(139, 233, 253),

			SectionSystem:    fgRGB(139, 233, 253),
			SectionResources: fgRGB(80, 250, 123),
			SectionNetwork:   fgRGB(189, 147, 249),
			SectionWeather:   fgRGB(255, 184, 108),
			SectionContainer: fgRGB(139, 233, 253),
			SectionService:   fgRGB(255, 121, 198),
			SectionUpdate:    fgRGB(241, 250, 140),
			SectionLogin:     fgRGB(80, 250, 123),
			SectionCalendar:  fgRGB(189, 147, 249),
			SectionQuote:     fgRGB(255, 121, 198),
			SectionArt:       fgRGB(255, 184, 108),
			SectionLogo:      fgRGB(248, 248, 242),

			GradientLow:  fgRGB(80, 250, 123),
			GradientMid:  fgRGB(241, 250, 140),
			GradientHigh: fgRGB(255, 85, 85),
		},
	}
}

// themeNord is the Nord palette (https://www.nordtheme.com).
func themeNord() *Theme {
	return &Theme{
		ID:          "nord",
		Name:        "Nord",
		Description: "Nord palette — cold arctic blues",
		UseUnicode:  true,
		Attrs: Attrs{
			Bold: ansiBold, Dim: ansiDim, Italic: ansiItalic, Underline: ansiUnderline,
		},
		Palette: Palette{
			Foreground: fgRGB(216, 222, 233),
			Muted:      fgRGB(143, 188, 187),
			Subtle:     fgRGB(76, 86, 106),

			Accent:    fgRGB(136, 192, 208),
			Secondary: fgRGB(180, 142, 173),
			Success:   fgRGB(163, 190, 140),
			Warning:   fgRGB(235, 203, 139),
			Danger:    fgRGB(191, 97, 106),
			Info:      fgRGB(129, 161, 193),

			SectionSystem:    fgRGB(136, 192, 208),
			SectionResources: fgRGB(143, 188, 187),
			SectionNetwork:   fgRGB(129, 161, 193),
			SectionWeather:   fgRGB(235, 203, 139),
			SectionContainer: fgRGB(94, 129, 172),
			SectionService:   fgRGB(180, 142, 173),
			SectionUpdate:    fgRGB(208, 135, 112),
			SectionLogin:     fgRGB(163, 190, 140),
			SectionCalendar:  fgRGB(136, 192, 208),
			SectionQuote:     fgRGB(180, 142, 173),
			SectionArt:       fgRGB(143, 188, 187),
			SectionLogo:      fgRGB(216, 222, 233),

			GradientLow:  fgRGB(163, 190, 140),
			GradientMid:  fgRGB(235, 203, 139),
			GradientHigh: fgRGB(191, 97, 106),
		},
	}
}

// themeSolarizedDark is the canonical Solarized Dark palette.
func themeSolarizedDark() *Theme {
	return &Theme{
		ID:          "solarized-dark",
		Name:        "Solarized Dark",
		Description: "Ethan Schoonover's Solarized — dark variant",
		UseUnicode:  true,
		Attrs: Attrs{
			Bold: ansiBold, Dim: ansiDim, Italic: ansiItalic, Underline: ansiUnderline,
		},
		Palette: Palette{
			Foreground: fgRGB(147, 161, 161),
			Muted:      fgRGB(101, 123, 131),
			Subtle:     fgRGB(88, 110, 117),

			Accent:    fgRGB(38, 139, 210),  // blue
			Secondary: fgRGB(108, 113, 196), // violet
			Success:   fgRGB(133, 153, 0),   // green
			Warning:   fgRGB(181, 137, 0),   // yellow
			Danger:    fgRGB(220, 50, 47),   // red
			Info:      fgRGB(42, 161, 152),  // cyan

			SectionSystem:    fgRGB(38, 139, 210),
			SectionResources: fgRGB(42, 161, 152),
			SectionNetwork:   fgRGB(38, 139, 210),
			SectionWeather:   fgRGB(181, 137, 0),
			SectionContainer: fgRGB(108, 113, 196),
			SectionService:   fgRGB(211, 54, 130),
			SectionUpdate:    fgRGB(203, 75, 22),
			SectionLogin:     fgRGB(133, 153, 0),
			SectionCalendar:  fgRGB(42, 161, 152),
			SectionQuote:     fgRGB(108, 113, 196),
			SectionArt:       fgRGB(211, 54, 130),
			SectionLogo:      fgRGB(147, 161, 161),

			GradientLow:  fgRGB(133, 153, 0),
			GradientMid:  fgRGB(181, 137, 0),
			GradientHigh: fgRGB(220, 50, 47),
		},
	}
}

// themeMonochrome uses only attribute changes (bold/dim/italic) and the
// terminal's default foreground. Works on monochrome terminals or when the
// user wants a completely neutral MOTD.
func themeMonochrome() *Theme {
	return &Theme{
		ID:          "monochrome",
		Name:        "Monochrome",
		Description: "No colors, only bold/dim/italic — friendly to color-blind users",
		UseUnicode:  true,
		Attrs: Attrs{
			Bold: ansiBold, Dim: ansiDim, Italic: ansiItalic, Underline: ansiUnderline,
		},
		Palette: Palette{
			Foreground: "",
			Muted:      ansiDim,
			Subtle:     ansiDim,
			Accent:     ansiBold,
			Secondary:  ansiBold,
			Success:    ansiBold,
			Warning:    ansiBold + ansiUnderline,
			Danger:     ansiBold + ansiUnderline,
			Info:       ansiItalic,

			SectionSystem:    ansiBold,
			SectionResources: ansiBold,
			SectionNetwork:   ansiBold,
			SectionWeather:   ansiBold,
			SectionContainer: ansiBold,
			SectionService:   ansiBold,
			SectionUpdate:    ansiBold,
			SectionLogin:     ansiBold,
			SectionCalendar:  ansiBold,
			SectionQuote:     ansiItalic,
			SectionArt:       ansiBold,
			SectionLogo:      ansiBold,

			GradientLow:  "",
			GradientMid:  ansiDim,
			GradientHigh: ansiBold,
		},
	}
}

// themeASCII strips all color and decoration. Output is plain ASCII suitable
// for logs, ssh sessions to dumb terminals, or piping to a file.
func themeASCII() *Theme {
	return &Theme{
		ID:          "ascii",
		Name:        "ASCII",
		Description: "No colors, no Unicode — pure plain text",
		UseUnicode:  false,
		Attrs:       Attrs{},   // all empty
		Palette:     Palette{}, // all empty
	}
}
