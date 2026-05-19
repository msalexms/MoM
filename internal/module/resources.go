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
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantDetailed}
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
	const barWidth = 28

	var sb strings.Builder
	sb.WriteString(r.Header("Resources", "resources"))
	sb.WriteString("\n\n")

	load1, load5, load15 := readLoadAvgValues()
	cpus := cpuCount()

	if r.Variant() == render.VariantCompact {
		memTotal, memAvail := readMemInfo()
		memPercent := 0.0
		if memTotal > 0 {
			memPercent = float64(memTotal-memAvail) / float64(memTotal) * 100
		}
		sb.WriteString(fmt.Sprintf("    load: %.1f/%.1f/%.1f  ram: %.0f%%",
			load1, load5, load15, memPercent))
		return sb.String(), nil
	}

	sb.WriteString(fmt.Sprintf("    %-10s  %s\n", "cpu 1m", r.ProgressBar(load1*100/cpus, barWidth, "")))
	sb.WriteString(fmt.Sprintf("    %-10s  %s\n", "cpu 5m", r.ProgressBar(load5*100/cpus, barWidth, "")))
	sb.WriteString(fmt.Sprintf("    %-10s  %s\n", "cpu 15m", r.ProgressBar(load15*100/cpus, barWidth, "")))
	sb.WriteString("\n")

	memTotal, memAvail := readMemInfo()
	memUsed := memTotal - memAvail
	memPercent := 0.0
	if memTotal > 0 {
		memPercent = float64(memUsed) / float64(memTotal) * 100
	}
	sb.WriteString(fmt.Sprintf("    %-10s  %s\n", "ram",
		r.ProgressBar(memPercent, barWidth,
			fmt.Sprintf("%s / %s", render.FormatBytes(memUsed*1024), render.FormatBytes(memTotal*1024)))))
	sb.WriteString("\n")

	swapTotal, swapFree := readSwapInfo()
	if swapTotal > 0 {
		swapUsed := swapTotal - swapFree
		swapPercent := float64(swapUsed) / float64(swapTotal) * 100
		sb.WriteString(fmt.Sprintf("    %-10s  %s\n", "swap",
			r.ProgressBar(swapPercent, barWidth,
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
		percent := float64(used) / float64(total) * 100
		info := fmt.Sprintf("%s / %s", render.FormatBytes(used), render.FormatBytes(total))
		sb.WriteString(fmt.Sprintf("    %-10s  %s\n", p.mountpoint, r.ProgressBar(percent, barWidth, info)))
	}

	if m.ShowTemp {
		temp := readCPUTemp()
		if temp != "" {
			sb.WriteString("\n" + r.KeyValue("temp", temp))
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
