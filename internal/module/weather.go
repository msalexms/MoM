package module

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/msalexms/MoM/internal/config"
	"github.com/msalexms/MoM/internal/i18n"
	"github.com/msalexms/MoM/internal/module/render"
)

// WeatherModule displays current weather from wttr.in.
type WeatherModule struct {
	City  string
	Units string // "metric" or "imperial"
}

func (m *WeatherModule) Name() string  { return "weather" }
func (m *WeatherModule) Title() string { return "Weather" }

// UpdateFromConfig implements ConfigSync.
func (m *WeatherModule) UpdateFromConfig(cfg *config.Config) {
	m.City = cfg.Modules.WeatherConfig.City
	m.Units = cfg.Modules.WeatherConfig.Units
}
func (m *WeatherModule) Description() string    { return "Current weather from wttr.in (no API key)" }
func (m *WeatherModule) Dependencies() []string { return nil }
func (m *WeatherModule) Available() bool        { return true }
func (m *WeatherModule) Settings() []Setting {
	return []Setting{
		stringSetting("city", "City", "Empty: detected from your IP",
			func(c *config.Config) *string { return &c.Modules.WeatherConfig.City }),
		enumSetting("units", "Units", "metric (°C, km/h) or imperial (°F, mph)",
			[]string{"metric", "imperial"},
			func(c *config.Config) *string { return &c.Modules.WeatherConfig.Units }),
	}
}

func (m *WeatherModule) Generate(ctx context.Context, opts render.Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Build URL: use city name format to avoid coordinates in output
	city := m.City
	url := "https://wttr.in/"
	if city != "" {
		url += city
	}
	// %l = location name, %C = condition text, %t = temp, %w = wind
	// When city is empty, wttr.in auto-detects but %l may return coords.
	// We use a separate request for the location name if needed.
	url += "?format=%C+%t+%w&lang=" + string(i18n.Current())
	if m.Units == "imperial" {
		url += "&u"
	} else {
		url += "&m"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", nil
	}
	req.Header.Set("User-Agent", "mom/1.0 (github.com/msalexms/MoM)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil
	}

	weather := strings.TrimSpace(string(body))
	if weather == "" || strings.Contains(weather, "Unknown") {
		return "", nil
	}

	// Determine location display name
	location := city
	if location == "" {
		location = getLocationName(ctx)
	}

	r := render.New(opts)
	th := r.Theme()

	lines := []string{
		r.Field("location", th.Color(location, th.Palette.Accent)),
		r.Field("now", weather),
	}

	compact := location + " " + weather
	return r.Section("Weather", "weather", compact, lines), nil
}

// getLocationName fetches the city name from wttr.in using the %l format.
// Falls back to "Local" if it looks like coordinates.
func getLocationName(ctx context.Context) string {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://wttr.in/?format=%l", nil)
	if err != nil {
		return "Local"
	}
	req.Header.Set("User-Agent", "mom/1.0 (github.com/msalexms/MoM)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "Local"
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Local"
	}

	loc := strings.TrimSpace(string(body))
	// If it looks like coordinates (contains comma + numbers), extract city part
	if strings.Contains(loc, ",") {
		parts := strings.SplitN(loc, ",", 2)
		city := strings.TrimSpace(parts[0])
		// If the first part is numeric (lat), it's coords — return "Local"
		if len(city) > 0 && (city[0] == '-' || (city[0] >= '0' && city[0] <= '9')) {
			return "Local"
		}
		return city
	}
	if loc == "" {
		return "Local"
	}
	return loc
}
