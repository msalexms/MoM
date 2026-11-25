package theme

// themeAmstech is the AMSTech brand theme, built from the "noche" tokens
// of the AMSTech design system (manual de marca v1.5, amstech.dev): the
// indigo brand color as accent, warm light text, the signal colors for
// status, and the series colors (homelab, IA, visión, cripto) to tell the
// module sections apart.
func themeAmstech() *Theme {
	var (
		texto   = fgRGB(0xE9, 0xE6, 0xE0) // --ams-texto
		texto2  = fgRGB(0xC4, 0xC0, 0xCB) // --ams-texto-2
		gris    = fgRGB(0x9E, 0x9A, 0xA8) // --ams-gris
		marca   = fgRGB(0x7F, 0x89, 0xE8) // --ams-marca
		thor    = fgRGB(0x9E, 0xA6, 0xF2) // --ams-thor
		ok      = fgRGB(0x8F, 0xD3, 0xA4) // --ams-ok
		mal     = fgRGB(0xF0, 0xA0, 0x94) // --ams-mal
		senal   = fgRGB(0xE0, 0x7A, 0x62) // --ams-senal
		amarill = fgRGB(0xD9, 0xC2, 0x6A) // --ams-amarillo
		homelab = fgRGB(0x6F, 0xD0, 0x8C) // --ams-serie-homelab
		ia      = fgRGB(0xB4, 0x8C, 0xFF) // --ams-serie-ia
		ia2     = fgRGB(0xFF, 0x8F, 0xB8) // --ams-serie-ia-2
		vision  = fgRGB(0x7C, 0xC4, 0xFF) // --ams-serie-vision
		cripto  = fgRGB(0xFF, 0x9F, 0x4A) // --ams-serie-cripto
		cripto2 = fgRGB(0xFF, 0xD2, 0x3F) // --ams-serie-cripto-2
	)

	return &Theme{
		ID:          "amstech",
		Name:        "AMSTech",
		Description: "The amstech.dev brand: indigo, Thor and series colors",
		UseUnicode:  true,
		Attrs: Attrs{
			Bold: ansiBold, Dim: ansiDim, Italic: ansiItalic, Underline: ansiUnderline,
		},
		Palette: Palette{
			Foreground: texto,
			Muted:      texto2,
			Subtle:     gris,

			Accent:    marca,
			Secondary: thor,
			Success:   ok,
			Warning:   amarill,
			Danger:    mal,
			Info:      vision,

			SectionSystem:    marca,
			SectionResources: homelab,
			SectionNetwork:   vision,
			SectionWeather:   cripto2,
			SectionContainer: vision,
			SectionService:   ia,
			SectionUpdate:    cripto,
			SectionLogin:     ia2,
			SectionCalendar:  marca,
			SectionQuote:     ia,
			SectionArt:       thor,
			SectionLogo:      thor,

			GradientLow:  ok,
			GradientMid:  amarill,
			GradientHigh: senal,
		},
	}
}
