package module

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/msalexms/MoM/internal/module/render"
)

// GPUModule displays GPU utilization and memory usage.
type GPUModule struct{}

func (m *GPUModule) Name() string           { return "gpu" }
func (m *GPUModule) Title() string          { return "GPU" }
func (m *GPUModule) Description() string    { return "GPU utilization, memory, temperature" }
func (m *GPUModule) Dependencies() []string { return nil }
func (m *GPUModule) DefaultEnabled() bool   { return false }

func (m *GPUModule) Available() bool {
	return CheckDependency("nvidia-smi") || hasAMDGPU()
}

func (m *GPUModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *GPUModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *GPUModule) Settings() []SettingDef         { return nil }

func (m *GPUModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

type gpuInfo struct {
	name     string
	util     int // percentage
	memUsed  uint64
	memTotal uint64
	temp     int
}

func (m *GPUModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	info := getGPUInfo(ctx)
	if info.name == "" {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()

	memPct := 0.0
	if info.memTotal > 0 {
		memPct = float64(info.memUsed) / float64(info.memTotal) * 100
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("%-10s  %s", th.Color("model", th.Palette.Warning), info.name))
	lines = append(lines, fmt.Sprintf("%-10s  %s", th.Color("util", th.Palette.Warning), r.ProgressBar(float64(info.util), 24, fmt.Sprintf("%d%%", info.util))))
	lines = append(lines, fmt.Sprintf("%-10s  %s", th.Color("vram", th.Palette.Warning), r.ProgressBar(memPct, 24, fmt.Sprintf("%s/%s", render.FormatBytes(info.memUsed*1024*1024), render.FormatBytes(info.memTotal*1024*1024)))))
	if info.temp > 0 {
		lines = append(lines, fmt.Sprintf("%-10s  %s", th.Color("temp", th.Palette.Warning), th.Color(fmt.Sprintf("%d°C", info.temp), th.PercentColor(float64(info.temp)))))
	}

	compact := fmt.Sprintf("%s  util %d%%  mem %.0f%%  %d°C", info.name, info.util, memPct, info.temp)
	return r.Section("GPU", "gpu", compact, lines), nil
}

func getGPUInfo(ctx context.Context) gpuInfo {
	if CheckDependency("nvidia-smi") {
		return getNvidiaInfo(ctx)
	}
	return getAMDInfo()
}

func getNvidiaInfo(ctx context.Context) gpuInfo {
	cmd := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=name,utilization.gpu,memory.used,memory.total,temperature.gpu", "--format=csv,noheader,nounits")
	out, err := cmd.Output()
	if err != nil {
		return gpuInfo{}
	}
	line := strings.TrimSpace(string(out))
	parts := strings.Split(line, ", ")
	if len(parts) < 5 {
		return gpuInfo{}
	}
	var info gpuInfo
	info.name = strings.TrimSpace(parts[0])
	fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &info.util)
	fmt.Sscanf(strings.TrimSpace(parts[2]), "%d", &info.memUsed)
	fmt.Sscanf(strings.TrimSpace(parts[3]), "%d", &info.memTotal)
	fmt.Sscanf(strings.TrimSpace(parts[4]), "%d", &info.temp)
	return info
}

func hasAMDGPU() bool {
	entries, err := os.ReadDir("/sys/class/drm")
	if err != nil {
		return false
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "card") && !strings.Contains(e.Name(), "-") {
			if _, err := os.Stat("/sys/class/drm/" + e.Name() + "/device/gpu_busy_percent"); err == nil {
				return true
			}
		}
	}
	return false
}

func getAMDInfo() gpuInfo {
	entries, _ := os.ReadDir("/sys/class/drm")
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "card") || strings.Contains(e.Name(), "-") {
			continue
		}
		base := "/sys/class/drm/" + e.Name() + "/device/"
		util := readSysInt(base + "gpu_busy_percent")
		if util < 0 {
			continue
		}
		var info gpuInfo
		info.name = strings.TrimSpace(readFileStr(base + "product_name"))
		if info.name == "" {
			info.name = "AMD GPU"
		}
		info.util = util
		info.temp = readSysInt(base+"hwmon/hwmon0/temp1_input") / 1000
		memUsed := readSysInt(base + "mem_info_vram_used")
		memTotal := readSysInt(base + "mem_info_vram_total")
		if memUsed > 0 {
			info.memUsed = uint64(memUsed) / (1024 * 1024)
		}
		if memTotal > 0 {
			info.memTotal = uint64(memTotal) / (1024 * 1024)
		}
		return info
	}
	return gpuInfo{}
}

func readSysInt(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return -1
	}
	var v int
	fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &v)
	return v
}
