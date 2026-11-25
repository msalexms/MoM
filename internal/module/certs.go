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

	"github.com/msalexms/MoM/internal/i18n"

	"github.com/msalexms/MoM/internal/module/render"
)

// CertsModule displays TLS certificates nearing expiration.
type CertsModule struct{}

func (m *CertsModule) Name() string           { return "certs" }
func (m *CertsModule) Title() string          { return "TLS certificates" }
func (m *CertsModule) Description() string    { return "Certificates expiring within 30 days" }
func (m *CertsModule) Dependencies() []string { return nil }
func (m *CertsModule) Available() bool {
	_, err := os.Stat("/etc/letsencrypt/live")
	return err == nil
}

type certInfo struct {
	domain  string
	expires time.Time
	days    int
}

func (m *CertsModule) Generate(ctx context.Context, opts render.Options) (string, error) {
	certs := scanCerts()
	if len(certs) == 0 {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()

	var lines []string
	var compactParts []string
	for _, c := range certs {
		color := th.Palette.Success
		if c.days < 7 {
			color = th.Palette.Danger
		} else if c.days < 14 {
			color = th.Palette.Warning
		}
		lines = append(lines, fmt.Sprintf("%s  %s", r.Col(c.domain, 20, th.Palette.Warning), th.Color(fmt.Sprintf(i18n.T("%d days"), c.days), color)))
		compactParts = append(compactParts, fmt.Sprintf("%s:%s", c.domain, th.Color(fmt.Sprintf("%dd", c.days), color)))
	}

	compact := strings.Join(compactParts, "  ")
	return r.Section("TLS certificates", "certs", compact, lines), nil
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
