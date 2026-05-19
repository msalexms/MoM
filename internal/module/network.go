package module

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ams/mom/internal/module/render"
)

// NetworkModule displays public and private IP addresses.
type NetworkModule struct{}

func (m *NetworkModule) Name() string           { return "network" }
func (m *NetworkModule) Title() string          { return "Network" }
func (m *NetworkModule) Description() string    { return "Public and private IP addresses" }
func (m *NetworkModule) Dependencies() []string { return nil }
func (m *NetworkModule) Available() bool        { return true }
func (m *NetworkModule) DefaultEnabled() bool   { return false }

func (m *NetworkModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact}
}
func (m *NetworkModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *NetworkModule) Settings() []SettingDef         { return nil }

func (m *NetworkModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

func (m *NetworkModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	r := render.New(opts)
	var sb strings.Builder

	sb.WriteString(r.Header("Network", "network"))
	sb.WriteString("\n\n")

	privates := getPrivateIPs()
	for _, ip := range privates {
		sb.WriteString(r.KeyValue("local", ip) + "\n")
	}
	if len(privates) == 0 {
		sb.WriteString(r.KeyValue("local", "no interface") + "\n")
	}

	publicIP := getPublicIP(ctx)
	sb.WriteString(r.KeyValue("public", publicIP))

	return sb.String(), nil
}

func getPrivateIPs() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.IsLoopback() {
				continue
			}
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}
	return ips
}

func getPublicIP(ctx context.Context) string {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.ipify.org", nil)
	if err != nil {
		return "N/A"
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "N/A"
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "N/A"
	}
	return strings.TrimSpace(string(body))
}
