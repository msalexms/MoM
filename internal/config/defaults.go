package config

// Defaults returns the configuration used before the user saves anything:
// the distro logo and the system summary, AMSTech theme.
func Defaults() *Config {
	return &Config{
		Modules: ModulesConfig{
			System: true,
			Logo:   true,

			WeatherConfig:    WeatherModuleConfig{Units: "metric"},
			CowsayConfig:     CowsayModuleConfig{Mode: "ascii-art", Message: "Hello"},
			ContainersConfig: ContainersModuleConfig{Runtime: "auto"},
			GitConfig:        GitModuleConfig{MaxRepos: 5},
		},
		Mode: ModeConfig{
			Language: "en",
			Theme:    DefaultTheme,
			Variant:  "default",
		},
	}
}
