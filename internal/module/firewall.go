package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ams/mom/internal/module/render"
)

// FirewallModule displays active firewall rules summary.
type FirewallModule struct{}

func (m *FirewallModule) Name() string           { return "firewall" }
func (m *FirewallModule) Title() string          { return "Firewall" }
func (m *FirewallModule) Description() string    { return "Active firewall rules (ufw/nftables)" }
func (m *FirewallModule) Dependencies() []string { return nil }
func (m *FirewallModule) DefaultEnabled() bool   { return false }
func (m *FirewallModule) Available() bool {
	return CheckDependency("ufw") || CheckDependency("nft")
}
func (m *FirewallModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *FirewallModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *FirewallModule) Settings() []SettingDef         { return nil }

func (m *FirewallModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

func (m *FirewallModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var status string
	var rules []string

	if CheckDependency("ufw") {
		status, rules = getUFWStatus(ctx)
	} else if CheckDependency("nft") {
		status, rules = getNFTStatus(ctx)
	}

	if status == "" {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()
	var sb strings.Builder

	statusColor := th.Palette.Success
	if status == "inactive" {
		statusColor = th.Palette.Danger
	}

	switch r.Variant() {
	case render.VariantCompact:
		sb.WriteString(r.Header("Firewall", "firewall"))
		sb.WriteString(fmt.Sprintf("\n    %s  %d rules", th.Color(status, statusColor), len(rules)))
	case render.VariantBoxed:
		var c strings.Builder
		c.WriteString(fmt.Sprintf("%-8s  %s\n", "status", th.Color(status, statusColor)))
		for _, rule := range rules {
			c.WriteString(truncate(rule, 40) + "\n")
		}
		sb.WriteString(render.Indent(r.Box(strings.TrimRight(c.String(), "\n"), "Firewall"), "  "))
	default:
		sb.WriteString(r.Header("Firewall", "firewall"))
		sb.WriteString("\n\n")
		sb.WriteString(r.KeyValue("status", th.Color(status, statusColor)) + "\n")
		for _, rule := range rules {
			sb.WriteString("    " + th.Dim(truncate(rule, 44)) + "\n")
		}
	}
	return sb.String(), nil
}

func getUFWStatus(ctx context.Context) (string, []string) {
	cmd := exec.CommandContext(ctx, "ufw", "status")
	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	status := "inactive"
	var rules []string
	for _, line := range lines {
		if strings.Contains(line, "Status: active") {
			status = "active"
		}
		if strings.Contains(line, "ALLOW") || strings.Contains(line, "DENY") || strings.Contains(line, "REJECT") {
			rules = append(rules, strings.TrimSpace(line))
		}
	}
	if len(rules) > 6 {
		rules = rules[:6]
	}
	return status, rules
}

func getNFTStatus(ctx context.Context) (string, []string) {
	cmd := exec.CommandContext(ctx, "nft", "list", "ruleset")
	output, err := cmd.Output()
	if err != nil {
		return "inactive", nil
	}
	lines := strings.Split(string(output), "\n")
	var chains []string
	for _, line := range lines {
		if strings.Contains(line, "chain") {
			chains = append(chains, strings.TrimSpace(line))
		}
	}
	if len(chains) > 6 {
		chains = chains[:6]
	}
	return "active", chains
}
