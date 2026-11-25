package module

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os/exec"
	"strings"
	"time"

	"github.com/msalexms/MoM/internal/config"
	"github.com/msalexms/MoM/internal/module/render"
)

// CowsayModule renders messages with cowsay, figlet, or built-in ASCII art.
type CowsayModule struct {
	Mode    string // cowsay | figlet | ascii-art | random
	Message string
}

func (m *CowsayModule) Name() string  { return "cowsay" }
func (m *CowsayModule) Title() string { return "ASCII art" }

// UpdateFromConfig implements ConfigSync.
func (m *CowsayModule) UpdateFromConfig(cfg *config.Config) {
	m.Mode = cfg.Modules.CowsayConfig.Mode
	m.Message = cfg.Modules.CowsayConfig.Message
}
func (m *CowsayModule) Description() string { return "Your own text as ASCII art, cowsay or figlet" }

func (m *CowsayModule) Dependencies() []string {
	switch m.Mode {
	case "cowsay":
		return []string{"cowsay"}
	case "figlet":
		return []string{"figlet"}
	case "random":
		return []string{"cowsay", "figlet"}
	default:
		return nil
	}
}

func (m *CowsayModule) Available() bool { return true }
func (m *CowsayModule) Settings() []Setting {
	return []Setting{
		enumSetting("mode", "Mode", "ascii-art needs nothing installed; random picks one each time",
			cowsayModes, func(c *config.Config) *string { return &c.Modules.CowsayConfig.Mode }),
		stringSetting("message", "Message", "Text drawn in the MOTD",
			func(c *config.Config) *string { return &c.Modules.CowsayConfig.Message }),
	}
}

var cowsayModes = []string{"cowsay", "figlet", "ascii-art", "random"}

func (m *CowsayModule) Generate(ctx context.Context, opts render.Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	message := m.Message
	if message == "" {
		return "", nil
	}

	mode := m.Mode
	if mode == "random" {
		options := []string{"ascii-art"}
		if CheckDependency("figlet") {
			options = append(options, "figlet")
		}
		if CheckDependency("cowsay") {
			options = append(options, "cowsay")
		}
		mode = options[rand.IntN(len(options))]
	}

	r := render.New(opts)
	var output string
	var err error

	switch mode {
	case "cowsay":
		output, err = m.runCowsay(ctx, message)
	case "figlet":
		output, err = m.runFiglet(ctx, message)
	default:
		output = r.AsciiBanner(message)
	}

	if err != nil || output == "" {
		output = r.AsciiBanner(message)
	}

	var sb strings.Builder
	sb.WriteString(r.Header("ASCII art", "cowsay"))
	sb.WriteString("\n\n")
	sb.WriteString(render.Indent(output, "    "))
	sb.WriteString("\n")

	return sb.String(), nil
}

func (m *CowsayModule) runCowsay(ctx context.Context, message string) (string, error) {
	if !CheckDependency("cowsay") {
		return "", fmt.Errorf("cowsay not installed")
	}
	cmd := exec.CommandContext(ctx, "cowsay", message)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(out), "\n"), nil
}

func (m *CowsayModule) runFiglet(ctx context.Context, message string) (string, error) {
	if !CheckDependency("figlet") {
		return "", fmt.Errorf("figlet not installed")
	}
	fonts := []string{"small", "standard", "slant", "mini", "banner3"}
	font := fonts[rand.IntN(len(fonts))]

	cmd := exec.CommandContext(ctx, "figlet", "-f", font, message)
	out, err := cmd.Output()
	if err != nil {
		cmd = exec.CommandContext(ctx, "figlet", message)
		out, err = cmd.Output()
		if err != nil {
			return "", err
		}
	}
	return strings.TrimRight(string(out), "\n"), nil
}
