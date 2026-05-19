package config

// Defaults returns a Config with sensible default values.
// Only system and logo modules are enabled by default.
func Defaults() *Config {
	return &Config{
		Motd: MotdConfig{
			Header: "",
			Footer: "",
		},
		Modules: ModulesConfig{
			System:     true,
			Resources:  false,
			Weather:    false,
			Cowsay:     false,
			Network:    false,
			Containers: false,
			Updates:    false,
			Logins:     false,
			Quote:      false,
			Calendar:   false,
			Services:   false,
			Logo:       true,

			WeatherConfig: WeatherModuleConfig{
				City:  "",
				Units: "metric",
			},
			ResourcesConfig: ResourcesModuleConfig{
				ShowTemp: false,
			},
			CowsayConfig: CowsayModuleConfig{
				Mode:    "cowsay",
				Message: "Welcome back!",
			},
			UpdatesConfig: UpdatesModuleConfig{
				IncludeAUR:   false,
				IncludeSnaps: false,
			},
			ContainersConfig: ContainersModuleConfig{
				Runtime: "auto",
			},
			ServicesConfig: ServicesModuleConfig{
				Services: nil,
			},
		},
		Mode: ModeConfig{
			Default:      "manual",
			LastTemplate: "",
			Theme:        "default",
			Variant:      "default",
		},
	}
}
