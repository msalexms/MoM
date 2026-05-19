package module

import (
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
)

// ResourcesModule displays CPU, RAM, and disk usage.
type ResourcesModule struct{}

func (m *ResourcesModule) Name() string        { return "resources" }
func (m *ResourcesModule) Title() string       { return "System Resources" }
func (m *ResourcesModule) Description() string { return "CPU load, RAM usage, disk usage" }
func (m *ResourcesModule) Dependencies() []string { return nil }
func (m *ResourcesModule) Available() bool     { return true }
func (m *ResourcesModule) DefaultEnabled() bool { return false }

func (m *ResourcesModule) Generate(ctx context.Context) (string, error) {
	var sb strings.Builder
	sb.WriteString("┌─ Resources ──────────────────────────┐\n")

	// Load average (instant alternative to CPU delta)
	load := readLoadAvg()
	sb.WriteString(fmt.Sprintf("│ Load:   %-29s │\n", load))

	// RAM
	memTotal, memAvail := readMemInfo()
	memUsed := memTotal - memAvail
	memPercent := 0.0
	if memTotal > 0 {
		memPercent = float64(memUsed) / float64(memTotal) * 100
	}
	sb.WriteString(fmt.Sprintf("│ RAM:    %dMB / %dMB (%.0f%%)%s│\n",
		memUsed/1024, memTotal/1024, memPercent,
		padding(fmt.Sprintf("%dMB / %dMB (%.0f%%)", memUsed/1024, memTotal/1024, memPercent), 29)))

	// Disk
	diskTotal, diskFree := readDiskUsage("/")
	diskUsed := diskTotal - diskFree
	diskPercent := 0.0
	if diskTotal > 0 {
		diskPercent = float64(diskUsed) / float64(diskTotal) * 100
	}
	sb.WriteString(fmt.Sprintf("│ Disk:   %dGB / %dGB (%.0f%%)%s│\n",
		diskUsed/(1024*1024*1024), diskTotal/(1024*1024*1024), diskPercent,
		padding(fmt.Sprintf("%dGB / %dGB (%.0f%%)", diskUsed/(1024*1024*1024), diskTotal/(1024*1024*1024), diskPercent), 29)))

	sb.WriteString("└───────────────────────────────────────┘")
	return sb.String(), nil
}

func readLoadAvg() string {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return "N/A"
	}
	parts := strings.Fields(string(data))
	if len(parts) >= 3 {
		return fmt.Sprintf("%s %s %s", parts[0], parts[1], parts[2])
	}
	return "N/A"
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

func readDiskUsage(path string) (total, free uint64) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0
	}
	total = stat.Blocks * uint64(stat.Bsize)
	free = stat.Bavail * uint64(stat.Bsize)
	return total, free
}

func padding(s string, width int) string {
	if len(s) >= width {
		return " "
	}
	return strings.Repeat(" ", width-len(s)) + " "
}
