package module

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// WeatherModule displays current weather from wttr.in.
type WeatherModule struct {
	City  string
	Units string // "metric" or "imperial"
}

func (m *WeatherModule) Name() string        { return "weather" }
func (m *WeatherModule) Title() string       { return "Weather" }
func (m *WeatherModule) Description() string { return "Current weather via wttr.in (no API key)" }
func (m *WeatherModule) Dependencies() []string { return nil }
func (m *WeatherModule) Available() bool     { return true }
func (m *WeatherModule) DefaultEnabled() bool { return false }

func (m *WeatherModule) Generate(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	url := "https://wttr.in/"
	if m.City != "" {
		url += m.City
	}
	url += "?format=3"
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
		return "", nil // Silently fail on network errors
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
	if weather == "" {
		return "", nil
	}

	return fmt.Sprintf("┌─ Weather ────────────────────────────┐\n│ %-37s │\n└───────────────────────────────────────┘", truncate(weather, 37)), nil
}
