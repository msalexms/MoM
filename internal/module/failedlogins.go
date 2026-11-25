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

// FailedLoginsModule displays failed SSH login attempts in the last 24h.
type FailedLoginsModule struct{}

func (m *FailedLoginsModule) Name() string  { return "failed-logins" }
func (m *FailedLoginsModule) Title() string { return "Failed logins" }
func (m *FailedLoginsModule) Description() string {
	return "Failed SSH attempts in the last 24 h"
}
func (m *FailedLoginsModule) Dependencies() []string { return []string{"journalctl"} }
func (m *FailedLoginsModule) Available() bool        { return CheckDependency("journalctl") }

func (m *FailedLoginsModule) Generate(ctx context.Context, opts render.Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "journalctl", "-t", "sshd", "-t", "sshd-session", "--since", "24h ago", "--no-pager", "-o", "short")
	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	// Count failed attempts and extract IPs
	ipCount := make(map[string]int)
	total := 0
	for _, line := range strings.Split(string(output), "\n") {
		if !strings.Contains(line, "Failed") && !strings.Contains(line, "Invalid user") {
			continue
		}
		total++
		// Extract IP: look for "from X.X.X.X"
		if idx := strings.Index(line, "from "); idx >= 0 {
			rest := line[idx+5:]
			fields := strings.Fields(rest)
			if len(fields) > 0 {
				ipCount[fields[0]]++
			}
		}
	}

	if total == 0 {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()

	// Top offending IPs
	type ipEntry struct {
		ip    string
		count int
	}
	var topIPs []ipEntry
	for ip, count := range ipCount {
		topIPs = append(topIPs, ipEntry{ip, count})
	}
	// Simple sort by count desc
	for i := range topIPs {
		for j := i + 1; j < len(topIPs); j++ {
			if topIPs[j].count > topIPs[i].count {
				topIPs[i], topIPs[j] = topIPs[j], topIPs[i]
			}
		}
	}
	if len(topIPs) > 5 {
		topIPs = topIPs[:5]
	}

	var lines []string
	lines = append(lines, r.Field("total", th.Color(fmt.Sprintf(i18n.T("%d attempts"), total), th.Palette.Danger)))
	for _, ip := range topIPs {
		lines = append(lines, fmt.Sprintf("%-16s  %d", ip.ip, ip.count))
	}

	compact := fmt.Sprintf(i18n.T("%s %d attempts from %d IPs"), th.Color("⚠", th.Palette.Danger), total, len(ipCount))
	return r.Section("Failed logins", "failed-logins", compact, lines), nil
}

// NeedsRoot implements Privileged.
func (m *FailedLoginsModule) NeedsRoot() bool { return true }
