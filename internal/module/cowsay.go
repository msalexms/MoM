package module

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os/exec"
	"strings"
	"time"
)

// CowsayModule renders messages with cowsay, figlet, or lolcat.
type CowsayModule struct {
	Mode    string // cowsay | figlet | lolcat | random
	Message string
}

func (m *CowsayModule) Name() string        { return "cowsay" }
func (m *CowsayModule) Title() string       { return "Cowsay / Figlet" }
func (m *CowsayModule) Description() string { return "Fun message with cowsay, figlet, or lolcat" }

func (m *CowsayModule) Dependencies() []string {
	switch m.Mode {
	case "cowsay":
		return []string{"cowsay"}
	case "figlet":
		return []string{"figlet"}
	case "lolcat":
		return []string{"lolcat"}
	case "random":
		return []string{"cowsay", "figlet"}
	default:
		return []string{"cowsay"}
	}
}

func (m *CowsayModule) Available() bool {
	// Available if at least one of cowsay or figlet exists
	return CheckDependency("cowsay") || CheckDependency("figlet")
}

func (m *CowsayModule) DefaultEnabled() bool { return false }

func (m *CowsayModule) Generate(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	message := m.Message
	if message == "" {
		message = "Welcome back!"
	}

	mode := m.Mode
	if mode == "random" {
		options := []string{}
		if CheckDependency("cowsay") {
			options = append(options, "cowsay")
		}
		if CheckDependency("figlet") {
			options = append(options, "figlet")
		}
		if len(options) == 0 {
			return "", nil
		}
		mode = options[rand.IntN(len(options))]
	}

	var cmd *exec.Cmd
	switch mode {
	case "cowsay":
		if !CheckDependency("cowsay") {
			return "", nil
		}
		cmd = exec.CommandContext(ctx, "cowsay", message)
	case "figlet":
		if !CheckDependency("figlet") {
			return "", nil
		}
		cmd = exec.CommandContext(ctx, "figlet", message)
	case "lolcat":
		if !CheckDependency("lolcat") {
			return "", nil
		}
		// lolcat colorizes stdin, so pipe figlet or echo into it
		if CheckDependency("figlet") {
			cmd = exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("figlet '%s' | lolcat -f", message))
		} else {
			cmd = exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("echo '%s' | lolcat -f", message))
		}
	default:
		return "", nil
	}

	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	return strings.TrimRight(string(output), "\n"), nil
}
