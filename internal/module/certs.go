package module

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ams/mom/internal/module/render"
)

// CertsModule displays TLS certificates nearing expiration.
type CertsModule struct{}

func (m *CertsModule) Name() string           { return "certs" }
func (m *CertsModule) Title() string          { return "TLS Certificates" }
func (m *CertsModule) Description() string    { return "Certificates expiring within 30 days" }
func (m *CertsModule) Dependencies() []string { return nil }
func (m *CertsModule) DefaultEnabled() bool   { return false }
func (m *CertsModule) Available() bool {
	_, err := os.Stat("/etc/letsencrypt/live")
	return err == nil
}
func (m *CertsModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *CertsModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *CertsModule) Settings() []SettingDef         { return nil }

func (m *CertsModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

type certInfo struct {
	domain  string
	expires time.Time
	days    int
}

func (m *CertsModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	certs := scanCerts()
	if len(certs) == 0 {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()
	var sb strings.Builder

	switch r.Variant() {
	case render.VariantCompact:
		sb.WriteString(r.Header("Certs", "certs"))
		sb.WriteString("\n    ")
		var parts []string
		for _, c := range certs {
			color := th.Palette.Success
			if c.days < 7 {
				color = th.Palette.Danger
			} else if c.days < 14 {
				color = th.Palette.Warning
			}
			parts = append(parts, fmt.Sprintf("%s:%s", c.domain, th.Color(fmt.Sprintf("%dd", c.days), color)))
		}
		sb.WriteString(strings.Join(parts, "  "))
	case render.VariantBoxed:
		var content strings.Builder
		for _, c := range certs {
			color := th.Palette.Success
			if c.days < 7 {
				color = th.Palette.Danger
			} else if c.days < 14 {
				color = th.Palette.Warning
			}
			content.WriteString(fmt.Sprintf("%-20s  %s\n", c.domain, th.Color(fmt.Sprintf("%d days left", c.days), color)))
		}
		sb.WriteString(render.Indent(r.Box(strings.TrimRight(content.String(), "\n"), "TLS Certs"), "  "))
	default:
		sb.WriteString(r.Header("TLS Certificates", "certs"))
		sb.WriteString("\n\n")
		for _, c := range certs {
			color := th.Palette.Success
			if c.days < 7 {
				color = th.Palette.Danger
			} else if c.days < 14 {
				color = th.Palette.Warning
			}
			sb.WriteString(fmt.Sprintf("    %-20s  %s\n", th.Color(c.domain, th.Palette.Warning), th.Color(fmt.Sprintf("%d days", c.days), color)))
		}
	}
	return sb.String(), nil
}

func scanCerts() []certInfo {
	base := "/etc/letsencrypt/live"
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil
	}
	var certs []certInfo
	now := time.Now()
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		certPath := filepath.Join(base, e.Name(), "cert.pem")
		data, err := os.ReadFile(certPath)
		if err != nil {
			continue
		}
		block, _ := pem.Decode(data)
		if block == nil {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}
		days := int(cert.NotAfter.Sub(now).Hours() / 24)
		if days <= 30 {
			certs = append(certs, certInfo{domain: e.Name(), expires: cert.NotAfter, days: days})
		}
	}
	return certs
}
