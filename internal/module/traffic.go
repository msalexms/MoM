package module

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ams/mom/internal/module/render"
)

// NetTrafficModule displays network TX/RX bytes per interface since boot.
type NetTrafficModule struct{}

func (m *NetTrafficModule) Name() string           { return "traffic" }
func (m *NetTrafficModule) Title() string          { return "Network Traffic" }
func (m *NetTrafficModule) Description() string    { return "TX/RX bytes per interface since boot" }
func (m *NetTrafficModule) Dependencies() []string { return nil }
func (m *NetTrafficModule) Available() bool        { return true }
func (m *NetTrafficModule) DefaultEnabled() bool   { return false }

func (m *NetTrafficModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *NetTrafficModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *NetTrafficModule) Settings() []SettingDef         { return nil }

func (m *NetTrafficModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

type ifaceTraffic struct {
	name string
	rx   uint64
	tx   uint64
}

func (m *NetTrafficModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	ifaces := getInterfaceTraffic()
	if len(ifaces) == 0 {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()
	var sb strings.Builder

	switch r.Variant() {
	case render.VariantCompact:
		sb.WriteString(r.Header("Traffic", "traffic"))
		sb.WriteString("\n    ")
		var parts []string
		for _, iface := range ifaces {
			parts = append(parts, fmt.Sprintf("%s ↓%s ↑%s", iface.name, render.FormatBytes(iface.rx), render.FormatBytes(iface.tx)))
		}
		sb.WriteString(strings.Join(parts, th.Color(" │ ", th.Palette.Subtle)))

	case render.VariantBoxed:
		var content strings.Builder
		for _, iface := range ifaces {
			content.WriteString(fmt.Sprintf("%-8s  ↓ %-10s  ↑ %s\n", iface.name, render.FormatBytes(iface.rx), render.FormatBytes(iface.tx)))
		}
		sb.WriteString(render.Indent(r.Box(strings.TrimRight(content.String(), "\n"), "Network Traffic"), "  "))

	case render.VariantPowerline:
		sb.WriteString(r.Header("Traffic", "traffic"))
		sb.WriteString("\n\n")
		for _, iface := range ifaces {
			sb.WriteString(fmt.Sprintf("    %s %-8s ↓ %-10s  ↑ %s\n",
				th.Color("▌", th.Palette.Accent),
				th.Color(iface.name, th.Palette.Warning),
				th.Color(render.FormatBytes(iface.rx), th.Palette.Success),
				th.Color(render.FormatBytes(iface.tx), th.Palette.Info)))
		}

	case render.VariantCards:
		var content strings.Builder
		for _, iface := range ifaces {
			content.WriteString(fmt.Sprintf("  %-8s  ↓ %-10s  ↑ %s\n", iface.name, render.FormatBytes(iface.rx), render.FormatBytes(iface.tx)))
		}
		sb.WriteString(render.Indent(r.Card(strings.TrimRight(content.String(), "\n"), "Network Traffic"), "  "))

	default:
		sb.WriteString(r.Header("Network Traffic", "traffic"))
		sb.WriteString("\n\n")
		for _, iface := range ifaces {
			sb.WriteString(fmt.Sprintf("    %-8s  ↓ %-10s  ↑ %s\n",
				th.Color(iface.name, th.Palette.Warning),
				th.Color(render.FormatBytes(iface.rx), th.Palette.Success),
				th.Color(render.FormatBytes(iface.tx), th.Palette.Info)))
		}
	}

	return sb.String(), nil
}

func getInterfaceTraffic() []ifaceTraffic {
	entries, err := os.ReadDir("/sys/class/net")
	if err != nil {
		return nil
	}

	var ifaces []ifaceTraffic
	for _, e := range entries {
		name := e.Name()
		if name == "lo" {
			continue
		}
		rxPath := "/sys/class/net/" + name + "/statistics/rx_bytes"
		txPath := "/sys/class/net/" + name + "/statistics/tx_bytes"

		rxData, err := os.ReadFile(rxPath)
		if err != nil {
			continue
		}
		txData, err := os.ReadFile(txPath)
		if err != nil {
			continue
		}

		var rx, tx uint64
		fmt.Sscanf(strings.TrimSpace(string(rxData)), "%d", &rx)
		fmt.Sscanf(strings.TrimSpace(string(txData)), "%d", &tx)

		if rx == 0 && tx == 0 {
			continue
		}

		ifaces = append(ifaces, ifaceTraffic{name, rx, tx})
	}
	return ifaces
}
