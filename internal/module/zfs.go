package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/msalexms/MoM/internal/module/render"
)

// ZFSModule displays ZFS pool status.
type ZFSModule struct{}

func (m *ZFSModule) Name() string           { return "zfs" }
func (m *ZFSModule) Title() string          { return "ZFS Pools" }
func (m *ZFSModule) Description() string    { return "ZFS pool health and usage" }
func (m *ZFSModule) Dependencies() []string { return []string{"zpool"} }
func (m *ZFSModule) Available() bool        { return CheckDependency("zpool") }
func (m *ZFSModule) DefaultEnabled() bool   { return false }
func (m *ZFSModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *ZFSModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *ZFSModule) Settings() []SettingDef         { return nil }

func (m *ZFSModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

type poolInfo struct {
	name   string
	health string
	used   string
	avail  string
}

func (m *ZFSModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "zpool", "list", "-H", "-o", "name,health,alloc,free")
	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	var pools []poolInfo
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		pools = append(pools, poolInfo{fields[0], fields[1], fields[2], fields[3]})
	}
	if len(pools) == 0 {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()

	var lines []string
	var compactParts []string
	for _, p := range pools {
		color := th.Palette.Success
		if p.health != "ONLINE" {
			color = th.Palette.Danger
		}
		lines = append(lines, fmt.Sprintf("%-12s  %s  used:%s  free:%s",
			th.Color(p.name, th.Palette.Warning), th.Color(p.health, color), p.used, p.avail))
		compactParts = append(compactParts, fmt.Sprintf("%s:%s", p.name, th.Color(p.health, color)))
	}

	compact := strings.Join(compactParts, "  ")
	return r.Section("ZFS Pools", "zfs", compact, lines), nil
}
