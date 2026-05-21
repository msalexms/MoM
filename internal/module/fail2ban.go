package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/msalexms/MoM/internal/module/render"
)

// Fail2banModule displays fail2ban jail status and banned IPs.
type Fail2banModule struct{}

func (m *Fail2banModule) Name() string           { return "fail2ban" }
func (m *Fail2banModule) Title() string          { return "Fail2ban" }
func (m *Fail2banModule) Description() string    { return "Active jails and banned IPs" }
func (m *Fail2banModule) Dependencies() []string { return []string{"fail2ban-client"} }
func (m *Fail2banModule) Available() bool        { return CheckDependency("fail2ban-client") }
func (m *Fail2banModule) DefaultEnabled() bool   { return false }

func (m *Fail2banModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *Fail2banModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *Fail2banModule) Settings() []SettingDef         { return nil }

func (m *Fail2banModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

type jailInfo struct {
	name   string
	banned int
}

func (m *Fail2banModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	jails := getJails(ctx)
	if len(jails) == 0 {
		return "", nil
	}

	totalBanned := 0
	for _, j := range jails {
		totalBanned += j.banned
	}

	r := render.New(opts)
	th := r.Theme()

	var lines []string
	for _, j := range jails {
		color := th.Palette.Success
		if j.banned > 0 {
			color = th.Palette.Danger
		}
		lines = append(lines, fmt.Sprintf("%-14s  %s", j.name, th.Color(fmt.Sprintf("%d banned", j.banned), color)))
	}

	compact := fmt.Sprintf("%d jails, %d banned", len(jails), totalBanned)
	return r.Section("Fail2ban", "fail2ban", compact, lines), nil
}

func getJails(ctx context.Context) []jailInfo {
	cmd := exec.CommandContext(ctx, "fail2ban-client", "status")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var jailNames []string
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "Jail list:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				for _, j := range strings.Split(parts[1], ",") {
					j = strings.TrimSpace(j)
					if j != "" {
						jailNames = append(jailNames, j)
					}
				}
			}
		}
	}

	var jails []jailInfo
	for _, name := range jailNames {
		banned := getJailBanned(ctx, name)
		jails = append(jails, jailInfo{name, banned})
	}
	return jails
}

func getJailBanned(ctx context.Context, jail string) int {
	cmd := exec.CommandContext(ctx, "fail2ban-client", "status", jail)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "Currently banned:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				var n int
				fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &n)
				return n
			}
		}
	}
	return 0
}
