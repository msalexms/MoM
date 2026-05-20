package module

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ams/mom/internal/module/render"
)

// WeatherModule displays current weather from wttr.in.
type WeatherModule struct {
	City  string
	Units string // "metric" or "imperial"
}

func (m *WeatherModule) Name() string           { return "weather" }
func (m *WeatherModule) Title() string          { return "Weather" }
func (m *WeatherModule) Description() string    { return "Current weather via wttr.in (no API key)" }
func (m *WeatherModule) Dependencies() []string { return nil }
func (m *WeatherModule) Available() bool        { return true }
func (m *WeatherModule) DefaultEnabled() bool   { return false }

func (m *WeatherModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *WeatherModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *WeatherModule) Settings() []SettingDef {
	return []SettingDef{
		{Key: "city", Label: "City", Type: SettingString, Default: ""},
		{Key: "units", Label: "Units", Type: SettingEnum, Default: "metric", Options: []string{"metric", "imperial"}},
	}
}

func (m *WeatherModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

func (m *WeatherModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
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
	url += "?format=%C+%t+%w"
	if m.Units == "imperial" {
		url += "&u"
	} else {
		url += "&m"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", nil
	}
	req.Header.Set("User-Agent", "mom-motd/1.0")

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
		fmt.Sprintf("%-10s  %s", th.Color("location", th.Palette.Warning), th.Color(location, th.Palette.Accent)),
		fmt.Sprintf("%-10s  %s", th.Color("current", th.Palette.Warning), weather),
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
	req.Header.Set("User-Agent", "mom-motd/1.0")

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
