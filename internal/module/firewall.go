package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/msalexms/MoM/internal/i18n"

	"github.com/msalexms/MoM/internal/module/render"
)

// FirewallModule displays active firewall rules summary.
type FirewallModule struct{}

func (m *FirewallModule) Name() string           { return "firewall" }
func (m *FirewallModule) Title() string          { return "Firewall" }
func (m *FirewallModule) Description() string    { return "Active firewall rules (ufw/nftables)" }
func (m *FirewallModule) Dependencies() []string { return []string{"ufw", "nft"} }
func (m *FirewallModule) Available() bool {
	return CheckDependency("ufw") || CheckDependency("nft")
}

func (m *FirewallModule) Generate(ctx context.Context, opts render.Options) (string, error) {
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

	statusColor := th.Palette.Success
	if status == "inactive" {
		statusColor = th.Palette.Danger
	}

	var lines []string
	lines = append(lines, r.Field("status", th.Color(status, statusColor)))
	for _, rule := range rules {
		lines = append(lines, th.Dim(truncate(rule, 44)))
	}

	compact := fmt.Sprintf(i18n.T("%s  %d rules"), th.Color(status, statusColor), len(rules))
	return r.Section("Firewall", "firewall", compact, lines), nil
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

// NeedsRoot implements Privileged.
func (m *FirewallModule) NeedsRoot() bool { return true }
