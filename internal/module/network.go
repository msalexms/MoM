package module

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// NetworkModule displays public and private IP addresses.
type NetworkModule struct{}

func (m *NetworkModule) Name() string        { return "network" }
func (m *NetworkModule) Title() string       { return "Network" }
func (m *NetworkModule) Description() string { return "Public and private IP addresses" }
func (m *NetworkModule) Dependencies() []string { return nil }
func (m *NetworkModule) Available() bool     { return true }
func (m *NetworkModule) DefaultEnabled() bool { return false }

func (m *NetworkModule) Generate(ctx context.Context) (string, error) {
	var sb strings.Builder
	sb.WriteString("┌─ Network ────────────────────────────┐\n")

	// Private IPs
	privates := getPrivateIPs()
	for _, ip := range privates {
		sb.WriteString(fmt.Sprintf("│ Local:  %-29s │\n", ip))
	}
	if len(privates) == 0 {
		sb.WriteString("│ Local:  N/A                           │\n")
	}

	// Public IP
	publicIP := getPublicIP(ctx)
	sb.WriteString(fmt.Sprintf("│ Public: %-29s │\n", publicIP))
	sb.WriteString("└───────────────────────────────────────┘")

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
