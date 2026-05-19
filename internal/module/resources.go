package module

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/ams/mom/internal/module/render"
)

// ResourcesModule displays CPU, RAM, and disk usage.
type ResourcesModule struct {
	ShowTemp bool
}

func (m *ResourcesModule) Name() string           { return "resources" }
func (m *ResourcesModule) Title() string          { return "Resources" }
func (m *ResourcesModule) Description() string    { return "CPU load, RAM, disk usage with progress bars" }
func (m *ResourcesModule) Dependencies() []string { return nil }
func (m *ResourcesModule) Available() bool        { return true }
func (m *ResourcesModule) DefaultEnabled() bool   { return false }

func (m *ResourcesModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantMinimal, render.VariantPowerline, render.VariantCards}
}
func (m *ResourcesModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *ResourcesModule) Settings() []SettingDef {
	return []SettingDef{
		{Key: "show_temp", Label: "Show CPU temperature", Type: SettingBool, Default: false},
	}
}

func (m *ResourcesModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

func (m *ResourcesModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	r := render.New(opts)
	th := r.Theme()
	const barWidth = 28

	load1, load5, load15 := readLoadAvgValues()
	cpus := cpuCount()
	memTotal, memAvail := readMemInfo()
	memUsed := memTotal - memAvail
	memPercent := 0.0
	if memTotal > 0 {
		memPercent = float64(memUsed) / float64(memTotal) * 100
	}

	var sb strings.Builder

	switch r.Variant() {
	case render.VariantCompact:
		sb.WriteString(r.Header("Resources", "resources"))
		sb.WriteString("\n")
		// Sparkline-style compact view
		loads := []float64{load1 * 100 / cpus, load5 * 100 / cpus, load15 * 100 / cpus}
		sb.WriteString(fmt.Sprintf("    %s %s %s  %s %s %.0f%%  ",
			r.Icon("cpu"), th.Color("load", th.Palette.Warning), r.Sparkline(loads),
			r.Icon("ram"), th.Color("ram", th.Palette.Warning), memPercent))
		// Disk summary
		partitions := getMountedPartitions()
		for _, p := range partitions {
			total, free := readDiskUsage(p.mountpoint)
			if total == 0 {
				continue
			}
			used := total - free
			pct := float64(used) / float64(total) * 100
			sb.WriteString(fmt.Sprintf("%s %.0f%% ", p.mountpoint, pct))
			break // only root in compact
		}

	case render.VariantBoxed:
		var content strings.Builder
		content.WriteString(fmt.Sprintf("%-6s  %s\n", "cpu 1m", r.ProgressBar(load1*100/cpus, 22, "")))
		content.WriteString(fmt.Sprintf("%-6s  %s\n", "cpu 5m", r.ProgressBar(load5*100/cpus, 22, "")))
		content.WriteString(fmt.Sprintf("%-6s  %s\n", "cpu15m", r.ProgressBar(load15*100/cpus, 22, "")))
		content.WriteString(r.SectionSeparator() + "\n")
		content.WriteString(fmt.Sprintf("%-6s  %s\n", "ram",
			r.ProgressBar(memPercent, 22, fmt.Sprintf("%s/%s", render.FormatBytes(memUsed*1024), render.FormatBytes(memTotal*1024)))))

		swapTotal, swapFree := readSwapInfo()
		if swapTotal > 0 {
			swapUsed := swapTotal - swapFree
			swapPct := float64(swapUsed) / float64(swapTotal) * 100
			content.WriteString(fmt.Sprintf("%-6s  %s\n", "swap",
				r.ProgressBar(swapPct, 22, fmt.Sprintf("%s/%s", render.FormatBytes(swapUsed*1024), render.FormatBytes(swapTotal*1024)))))
		}

		content.WriteString(r.SectionSeparator() + "\n")
		partitions := getMountedPartitions()
		for _, p := range partitions {
			total, free := readDiskUsage(p.mountpoint)
			if total == 0 {
				continue
			}
			used := total - free
			pct := float64(used) / float64(total) * 100
			label := p.mountpoint
			if len(label) > 6 {
				label = label[:6]
			}
			content.WriteString(fmt.Sprintf("%-6s  %s\n", label,
				r.ProgressBar(pct, 22, fmt.Sprintf("%s/%s", render.FormatBytes(used), render.FormatBytes(total)))))
		}

		if m.ShowTemp {
			if temp := readCPUTemp(); temp != "" {
				content.WriteString(fmt.Sprintf("%-6s  %s", "temp", th.Color(temp, th.Palette.Warning)))
			}
		}

		sb.WriteString(render.Indent(r.Box(strings.TrimRight(content.String(), "\n"), "Resources"), "  "))

	case render.VariantMinimal:
		sb.WriteString(r.Header("Resources", "resources"))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("    load %.1f/%.1f/%.1f  ram %.0f%%  ",
			load1, load5, load15, memPercent))
		partitions := getMountedPartitions()
		for _, p := range partitions {
			total, free := readDiskUsage(p.mountpoint)
			if total == 0 {
				continue
			}
			used := total - free
			pct := float64(used) / float64(total) * 100
			sb.WriteString(fmt.Sprintf("%s %.0f%%  ", p.mountpoint, pct))
		}

	case render.VariantPowerline:
		sb.WriteString(r.Header("Resources", "resources"))
		sb.WriteString("\n\n")
		sb.WriteString(fmt.Sprintf("    %s %-8s %s\n",
			th.Color("▌", th.Palette.Accent), th.Color("load", th.Palette.Warning),
			fmt.Sprintf("%.1f / %.1f / %.1f", load1, load5, load15)))
		sb.WriteString(fmt.Sprintf("    %s %-8s %s\n",
			th.Color("▌", th.Palette.Accent), th.Color("ram", th.Palette.Warning),
			r.ProgressBar(memPercent, 20, fmt.Sprintf("%s/%s", render.FormatBytes(memUsed*1024), render.FormatBytes(memTotal*1024)))))
		swapTotal, swapFree := readSwapInfo()
		if swapTotal > 0 {
			swapUsed := swapTotal - swapFree
			swapPct := float64(swapUsed) / float64(swapTotal) * 100
			sb.WriteString(fmt.Sprintf("    %s %-8s %s\n",
				th.Color("▌", th.Palette.Accent), th.Color("swap", th.Palette.Warning),
				r.ProgressBar(swapPct, 20, "")))
		}
		partitions := getMountedPartitions()
		for _, p := range partitions {
			total, free := readDiskUsage(p.mountpoint)
			if total == 0 {
				continue
			}
			used := total - free
			pct := float64(used) / float64(total) * 100
			label := p.mountpoint
			if len(label) > 8 {
				label = label[:8]
			}
			sb.WriteString(fmt.Sprintf("    %s %-8s %s\n",
				th.Color("▌", th.Palette.Accent), th.Color(label, th.Palette.Warning),
				r.ProgressBar(pct, 20, fmt.Sprintf("%s/%s", render.FormatBytes(used), render.FormatBytes(total)))))
		}

	case render.VariantCards:
		var content strings.Builder
		content.WriteString(fmt.Sprintf("  %-8s  %s\n", "load", fmt.Sprintf("%.1f / %.1f / %.1f", load1, load5, load15)))
		content.WriteString(fmt.Sprintf("  %-8s  %s\n", "ram", r.ProgressBar(memPercent, 20, fmt.Sprintf("%s/%s", render.FormatBytes(memUsed*1024), render.FormatBytes(memTotal*1024)))))
		swapTotal2, swapFree2 := readSwapInfo()
		if swapTotal2 > 0 {
			swapUsed2 := swapTotal2 - swapFree2
			swapPct2 := float64(swapUsed2) / float64(swapTotal2) * 100
			content.WriteString(fmt.Sprintf("  %-8s  %s\n", "swap", r.ProgressBar(swapPct2, 20, "")))
		}
		partitions := getMountedPartitions()
		for _, p := range partitions {
			total, free := readDiskUsage(p.mountpoint)
			if total == 0 {
				continue
			}
			used := total - free
			pct := float64(used) / float64(total) * 100
			label := p.mountpoint
			if len(label) > 8 {
				label = label[:8]
			}
			content.WriteString(fmt.Sprintf("  %-8s  %s\n", label, r.ProgressBar(pct, 20, fmt.Sprintf("%s/%s", render.FormatBytes(used), render.FormatBytes(total)))))
		}
		sb.WriteString(render.Indent(r.Card(strings.TrimRight(content.String(), "\n"), "Resources"), "  "))

	default:
		sb.WriteString(r.Header("Resources", "resources"))
		sb.WriteString("\n\n")
		sb.WriteString(fmt.Sprintf("    %-10s  %s\n", "cpu 1m", r.ProgressBar(load1*100/cpus, barWidth, "")))
		sb.WriteString(fmt.Sprintf("    %-10s  %s\n", "cpu 5m", r.ProgressBar(load5*100/cpus, barWidth, "")))
		sb.WriteString(fmt.Sprintf("    %-10s  %s\n", "cpu 15m", r.ProgressBar(load15*100/cpus, barWidth, "")))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("    %-10s  %s\n", "ram",
			r.ProgressBar(memPercent, barWidth,
				fmt.Sprintf("%s / %s", render.FormatBytes(memUsed*1024), render.FormatBytes(memTotal*1024)))))
		sb.WriteString("\n")

		swapTotal, swapFree := readSwapInfo()
		if swapTotal > 0 {
			swapUsed := swapTotal - swapFree
			swapPct := float64(swapUsed) / float64(swapTotal) * 100
			sb.WriteString(fmt.Sprintf("    %-10s  %s\n", "swap",
				r.ProgressBar(swapPct, barWidth,
					fmt.Sprintf("%s / %s", render.FormatBytes(swapUsed*1024), render.FormatBytes(swapTotal*1024)))))
			sb.WriteString("\n")
		}

		partitions := getMountedPartitions()
		for _, p := range partitions {
			total, free := readDiskUsage(p.mountpoint)
			if total == 0 {
				continue
			}
			used := total - free
			pct := float64(used) / float64(total) * 100
			info := fmt.Sprintf("%s / %s", render.FormatBytes(used), render.FormatBytes(total))
			sb.WriteString(fmt.Sprintf("    %-10s  %s\n", p.mountpoint, r.ProgressBar(pct, barWidth, info)))
		}

		if m.ShowTemp {
			if temp := readCPUTemp(); temp != "" {
				sb.WriteString("\n" + r.KeyValue("temp", temp))
			}
		}
	}

	return sb.String(), nil
}

type partition struct {
	device     string
	mountpoint string
	fstype     string
}

func getMountedPartitions() []partition {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return []partition{{mountpoint: "/"}}
	}
	defer f.Close()

	var parts []partition
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}
		device := fields[0]
		mount := fields[1]
		fstype := fields[2]

		if !strings.HasPrefix(device, "/dev/") {
			continue
		}
		if strings.Contains(mount, "/snap/") || strings.Contains(mount, "/docker/") {
			continue
		}
		if seen[device] {
			continue
		}
		seen[device] = true
		parts = append(parts, partition{device: device, mountpoint: mount, fstype: fstype})
	}

	if len(parts) == 0 {
		return []partition{{mountpoint: "/"}}
	}
	return parts
}

func readLoadAvgValues() (load1, load5, load15 float64) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, 0, 0
	}
	fmt.Sscanf(string(data), "%f %f %f", &load1, &load5, &load15)
	return
}

func readMemInfo() (total, available uint64) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fmt.Sscanf(line, "MemTotal: %d kB", &total)
		}
		if strings.HasPrefix(line, "MemAvailable:") {
			fmt.Sscanf(line, "MemAvailable: %d kB", &available)
		}
	}
	return total, available
}

func cpuCount() float64 {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return 1
	}
	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "processor") {
			count++
		}
	}
	if count == 0 {
		return 1
	}
	return float64(count)
}

func readSwapInfo() (total, free uint64) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "SwapTotal:") {
			fmt.Sscanf(line, "SwapTotal: %d kB", &total)
		}
		if strings.HasPrefix(line, "SwapFree:") {
			fmt.Sscanf(line, "SwapFree: %d kB", &free)
		}
	}
	return total, free
}

func readDiskUsage(path string) (total, free uint64) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0
	}
	total = stat.Blocks * uint64(stat.Bsize)
	free = stat.Bavail * uint64(stat.Bsize)
	return total, free
}

func readCPUTemp() string {
	paths := []string{
		"/sys/class/thermal/thermal_zone0/temp",
		"/sys/class/hwmon/hwmon0/temp1_input",
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var millideg int
		fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &millideg)
		if millideg > 0 {
			return fmt.Sprintf("%.1f°C", float64(millideg)/1000.0)
		}
	}
	return ""
}
