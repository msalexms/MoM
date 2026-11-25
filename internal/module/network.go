package module

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/msalexms/MoM/internal/i18n"

	"github.com/msalexms/MoM/internal/module/render"
)

// NetworkModule displays public and private IP addresses.
type NetworkModule struct{}

func (m *NetworkModule) Name() string           { return "network" }
func (m *NetworkModule) Title() string          { return "Network" }
func (m *NetworkModule) Description() string    { return "Public and private IPs" }
func (m *NetworkModule) Dependencies() []string { return nil }
func (m *NetworkModule) Available() bool        { return true }

func (m *NetworkModule) Generate(ctx context.Context, opts render.Options) (string, error) {
	r := render.New(opts)
	th := r.Theme()

	privates := getPrivateIPs()
	publicIP := getPublicIP(ctx)

	var lines []string
	for _, ip := range privates {
		lines = append(lines, r.Field("local", ip))
	}
	if len(privates) == 0 {
		lines = append(lines, r.Field("local", i18n.T("no interface")))
	}
	lines = append(lines, r.Field("public", th.Color(publicIP, th.Palette.Success)))

	compact := ""
	if len(privates) > 0 {
		compact = fmt.Sprintf("%s %s  %s %s", r.Icon("net"), privates[0], r.Icon("globe"), publicIP)
	} else {
		compact = fmt.Sprintf("%s %s", r.Icon("globe"), publicIP)
	}

	return r.Section("Network", "network", compact, lines), nil
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
