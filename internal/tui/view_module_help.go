package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/msalexms/MoM/internal/module"
)

func (m Model) viewModuleHelp() string {
	var sb strings.Builder

	allModules := m.registry.Ordered()
	if m.cursor >= len(allModules) {
		return "  No module selected"
	}
	mod := allModules[m.cursor]

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00BFFF"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Width(14)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CCCCCC"))

	sb.WriteString(titleStyle.Render(fmt.Sprintf("  :: %s — Help", mod.Title())) + "\n")
	sb.WriteString(viewSeparator() + "\n\n")

	// Basic info
	sb.WriteString(fmt.Sprintf("  %s%s\n", labelStyle.Render("Name"), valueStyle.Render(mod.Name())))
	sb.WriteString(fmt.Sprintf("  %s%s\n", labelStyle.Render("Description"), valueStyle.Render(mod.Description())))

	// Availability
	avail := "✓ Available"
	availColor := lipgloss.Color("#00FF7F")
	if !mod.Available() {
		avail = "✗ Not available"
		availColor = lipgloss.Color("#FF4444")
	}
	sb.WriteString(fmt.Sprintf("  %s%s\n", labelStyle.Render("Status"),
		lipgloss.NewStyle().Foreground(availColor).Render(avail)))

	// Dependencies
	deps := mod.Dependencies()
	if len(deps) > 0 {
		sb.WriteString(fmt.Sprintf("  %s%s\n", labelStyle.Render("Requires"),
			valueStyle.Render(strings.Join(deps, ", "))))
	} else {
		sb.WriteString(fmt.Sprintf("  %s%s\n", labelStyle.Render("Requires"),
			dimStyle.Render("none")))
	}

	// Variants (if Configurable)
	if cfg, ok := mod.(module.Configurable); ok {
		variants := cfg.Variants()
		if len(variants) > 0 {
			var vs []string
			for _, v := range variants {
				vs = append(vs, string(v))
			}
			sb.WriteString(fmt.Sprintf("  %s%s\n", labelStyle.Render("Variants"),
				valueStyle.Render(strings.Join(vs, ", "))))
		}

		// Settings
		settings := cfg.Settings()
		if len(settings) > 0 {
			sb.WriteString("\n")
			sb.WriteString(sectionStyle.Render("  Configuration") + "\n")
			sb.WriteString(dimStyle.Render("  Set in ~/.config/mom/config.toml") + "\n\n")

			for _, s := range settings {
				typeStr := settingTypeStr(s.Type)
				sb.WriteString(fmt.Sprintf("  %s%s\n",
					labelStyle.Render(s.Key),
					valueStyle.Render(s.Label)))
				sb.WriteString(fmt.Sprintf("  %s%s",
					labelStyle.Render(""),
					dimStyle.Render(fmt.Sprintf("Type: %s", typeStr))))
				if len(s.Options) > 0 {
					sb.WriteString(dimStyle.Render(fmt.Sprintf("  Options: %s", strings.Join(s.Options, ", "))))
				}
				if s.Default != nil {
					sb.WriteString(dimStyle.Render(fmt.Sprintf("  Default: %v", s.Default)))
				}
				sb.WriteString("\n")
				if s.Description != "" {
					sb.WriteString(fmt.Sprintf("  %s%s\n",
						labelStyle.Render(""),
						dimStyle.Render(s.Description)))
				}
			}
		} else {
			sb.WriteString("\n")
			sb.WriteString(dimStyle.Render("  No configuration needed — works out of the box") + "\n")
		}
	} else {
		sb.WriteString("\n")
		sb.WriteString(dimStyle.Render("  No configuration needed — works out of the box") + "\n")
	}

	// Config example
	sb.WriteString("\n")
	sb.WriteString(sectionStyle.Render("  TOML Example") + "\n")
	sb.WriteString(dimStyle.Render(fmt.Sprintf("  [modules]\n  %s = true", mod.Name())) + "\n")

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
		"  [Esc] Back to modules"))

	return sb.String()
}

func settingTypeStr(t module.SettingType) string {
	switch t {
	case module.SettingBool:
		return "bool"
	case module.SettingString:
		return "string"
	case module.SettingEnum:
		return "enum"
	case module.SettingList:
		return "list"
	case module.SettingInt:
		return "int"
	default:
		return "unknown"
	}
}
