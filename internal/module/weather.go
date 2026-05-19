package module

import (
	"context"
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
	return []render.Variant{render.VariantDefault, render.VariantCompact}
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

	url := "https://wttr.in/"
	if m.City != "" {
		url += m.City
	}
	url += "?format=%l:+%c+%C+%t+%w"
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

	raw := strings.TrimSpace(string(body))
	if raw == "" {
		return "", nil
	}

	location := m.City
	weather := raw
	if parts := strings.SplitN(raw, ": ", 2); len(parts) == 2 {
		location = strings.TrimSpace(parts[0])
		weather = strings.TrimSpace(parts[1])
	}
	if location == "" {
		location = "Unknown"
	}

	r := render.New(opts)
	var sb strings.Builder
	sb.WriteString(r.Header("Weather", "weather"))
	sb.WriteString("\n\n")
	sb.WriteString(r.KeyValue("location", location) + "\n")
	sb.WriteString(r.KeyValue("current", weather))

	return sb.String(), nil
}
