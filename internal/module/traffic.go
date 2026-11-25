package module

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/msalexms/MoM/internal/module/render"
)

// NetTrafficModule displays network TX/RX bytes per interface since boot.
type NetTrafficModule struct{}

func (m *NetTrafficModule) Name() string  { return "traffic" }
func (m *NetTrafficModule) Title() string { return "Network traffic" }
func (m *NetTrafficModule) Description() string {
	return "Bytes sent and received per interface since boot"
}
func (m *NetTrafficModule) Dependencies() []string { return nil }
func (m *NetTrafficModule) Available() bool        { return true }

type ifaceTraffic struct {
	name string
	rx   uint64
	tx   uint64
}

func (m *NetTrafficModule) Generate(ctx context.Context, opts render.Options) (string, error) {
	ifaces := getInterfaceTraffic()
	if len(ifaces) == 0 {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()

	var lines []string
	var compactParts []string
	for _, iface := range ifaces {
		lines = append(lines, fmt.Sprintf("%s  ↓ %s  ↑ %s",
			r.Col(iface.name, 8, th.Palette.Warning),
			r.Col(render.FormatBytes(iface.rx), 10, th.Palette.Success),
			th.Color(render.FormatBytes(iface.tx), th.Palette.Info)))
		compactParts = append(compactParts, fmt.Sprintf("%s ↓%s ↑%s", iface.name, render.FormatBytes(iface.rx), render.FormatBytes(iface.tx)))
	}

	compact := strings.Join(compactParts, th.Color(" │ ", th.Palette.Subtle))
	return r.Section("Network traffic", "traffic", compact, lines), nil
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
