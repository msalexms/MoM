package module

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ams/mom/internal/module/render"
)

// TopProcsModule displays the top processes by memory usage.
type TopProcsModule struct{}

func (m *TopProcsModule) Name() string           { return "procs" }
func (m *TopProcsModule) Title() string          { return "Top Processes" }
func (m *TopProcsModule) Description() string    { return "Top processes by memory usage" }
func (m *TopProcsModule) Dependencies() []string { return nil }
func (m *TopProcsModule) Available() bool        { return true }
func (m *TopProcsModule) DefaultEnabled() bool   { return false }

func (m *TopProcsModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *TopProcsModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *TopProcsModule) Settings() []SettingDef         { return nil }

func (m *TopProcsModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

type procInfo struct {
	name string
	mem  float64 // percentage
}

func (m *TopProcsModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	procs := getTopProcs(5)
	if len(procs) == 0 {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()

	var lines []string
	var compactParts []string
	for _, p := range procs {
		lines = append(lines, fmt.Sprintf("%-16s  %s", p.name, th.Color(fmt.Sprintf("%.1f%%", p.mem), th.PercentColor(p.mem))))
		compactParts = append(compactParts, fmt.Sprintf("%s %.0f%%", p.name, p.mem))
	}

	compact := strings.Join(compactParts, th.Color(" · ", th.Palette.Subtle))
	return r.Section("Top Processes", "procs", compact, lines), nil
}

func getTopProcs(n int) []procInfo {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}

	memTotal, _ := readMemInfo()
	if memTotal == 0 {
		return nil
	}

	type rawProc struct {
		name string
		rss  uint64
	}
	var all []rawProc

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if len(name) == 0 || name[0] < '0' || name[0] > '9' {
			continue
		}

		status, err := os.ReadFile("/proc/" + name + "/status")
		if err != nil {
			continue
		}

		var pname string
		var rss uint64
		for _, line := range strings.Split(string(status), "\n") {
			if strings.HasPrefix(line, "Name:") {
				pname = strings.TrimSpace(strings.TrimPrefix(line, "Name:"))
			}
			if strings.HasPrefix(line, "VmRSS:") {
				fmt.Sscanf(strings.TrimPrefix(line, "VmRSS:"), "%d", &rss)
			}
		}
		if pname != "" && rss > 0 {
			all = append(all, rawProc{pname, rss})
		}
	}

	// Aggregate by name
	agg := make(map[string]uint64)
	for _, p := range all {
		agg[p.name] += p.rss
	}

	var procs []procInfo
	for name, rss := range agg {
		pct := float64(rss) / float64(memTotal) * 100
		procs = append(procs, procInfo{name, pct})
	}

	sort.Slice(procs, func(i, j int) bool { return procs[i].mem > procs[j].mem })

	if len(procs) > n {
		procs = procs[:n]
	}
	return procs
}
