package module

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ams/mom/internal/module/render"
)

// DiskIOModule displays disk I/O statistics from /proc/diskstats.
type DiskIOModule struct{}

func (m *DiskIOModule) Name() string           { return "diskio" }
func (m *DiskIOModule) Title() string          { return "Disk I/O" }
func (m *DiskIOModule) Description() string    { return "Disk read/write totals and IO time" }
func (m *DiskIOModule) Dependencies() []string { return nil }
func (m *DiskIOModule) Available() bool        { return true }
func (m *DiskIOModule) DefaultEnabled() bool   { return false }

func (m *DiskIOModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *DiskIOModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *DiskIOModule) Settings() []SettingDef         { return nil }

func (m *DiskIOModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

type diskStat struct {
	name     string
	readMB   float64
	writeMB  float64
	ioTimeMs uint64
}

func (m *DiskIOModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	stats := getDiskStats()
	if len(stats) == 0 {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()

	var lines []string
	var compactParts []string
	for _, d := range stats {
		ioSec := float64(d.ioTimeMs) / 1000.0
		lines = append(lines, fmt.Sprintf("%-6s  R %s  W %s  IO %s",
			th.Color(d.name, th.Palette.Warning),
			th.Color(fmt.Sprintf("%.1fMB", d.readMB), th.Palette.Success),
			th.Color(fmt.Sprintf("%.1fMB", d.writeMB), th.Palette.Info),
			th.Dim(fmt.Sprintf("%.1fs", ioSec))))
		compactParts = append(compactParts, fmt.Sprintf("%s R:%.0fMB W:%.0fMB", d.name, d.readMB, d.writeMB))
	}

	compact := strings.Join(compactParts, th.Color(" │ ", th.Palette.Subtle))
	return r.Section("Disk I/O", "diskio", compact, lines), nil
}

func getDiskStats() []diskStat {
	data, err := os.ReadFile("/proc/diskstats")
	if err != nil {
		return nil
	}

	var stats []diskStat
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 14 {
			continue
		}
		name := fields[2]
		// Only show whole disks (sda, nvme0n1, vda) not partitions
		if strings.HasSuffix(name, "1") || strings.HasSuffix(name, "2") || strings.HasSuffix(name, "3") {
			// Check if it's a partition (has a digit before the last digit)
			if len(name) > 1 && name[len(name)-2] >= '0' && name[len(name)-2] <= '9' {
				continue
			}
			if strings.Contains(name, "p") && name[len(name)-1] >= '1' && name[len(name)-1] <= '9' {
				continue
			}
		}
		// Skip loop, ram, dm devices
		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") || strings.HasPrefix(name, "dm-") {
			continue
		}

		var sectorsRead, sectorsWritten, ioTime uint64
		fmt.Sscanf(fields[5], "%d", &sectorsRead)
		fmt.Sscanf(fields[9], "%d", &sectorsWritten)
		fmt.Sscanf(fields[12], "%d", &ioTime)

		if sectorsRead == 0 && sectorsWritten == 0 {
			continue
		}

		// Sectors are typically 512 bytes
		readMB := float64(sectorsRead) * 512 / (1024 * 1024)
		writeMB := float64(sectorsWritten) * 512 / (1024 * 1024)

		stats = append(stats, diskStat{name, readMB, writeMB, ioTime})
	}
	return stats
}
