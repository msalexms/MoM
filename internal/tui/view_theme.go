package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ams/mom/internal/module/render"
	"github.com/ams/mom/internal/theme"
)

func (m Model) viewTheme() string {
	var sb strings.Builder

	sb.WriteString(viewTitleStyle("#FF00FF").Render("  :: Theme & Style") + "\n")
	sb.WriteString(viewSeparator() + "\n\n")

	// Section 1: Themes
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CCCCCC")).Render("  Themes") + "\n")
	themes := theme.All()
	currentID := m.config.ThemeID()

	// Section 2: Variants
	variants := []struct {
		id   string
		desc string
	}{
		{"default", "Classic — separators and key:value pairs"},
		{"compact", "One-liner per module — sparklines, dots"},
		{"boxed", "Unicode boxes ╭─╮ with rounded corners"},
		{"powerline", "Powerline blocks ▌ with arrow separators"},
		{"cards", "Double-border cards ╔═╗ with shadow"},
		{"minimal", "Data only — no decoration"},
	}
	currentVariant := m.config.GlobalVariant()

	totalItems := len(themes) + 1 + len(variants) // +1 for the "Variants" header
	_ = totalItems

	idx := 0
	for i, th := range themes {
		cursor := "  "
		if idx == m.cursor {
			cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF00FF")).Bold(true).Render("▸ ")
		}

		nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
		if idx == m.cursor {
			nameStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF00FF"))
		}

		active := ""
		if th.ID == currentID {
			active = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Render(" ●")
		}

		sb.WriteString(fmt.Sprintf("%s%-14s %s%s\n", cursor,
			nameStyle.Render(th.Name),
			descStyle.Render(th.Description),
			active))
		_ = i
		idx++
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CCCCCC")).Render("  Style Variant") + "\n")
	idx++ // skip the header line in cursor math

	for _, v := range variants {
		cursor := "  "
		if idx == m.cursor {
			cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF00FF")).Bold(true).Render("▸ ")
		}

		nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
		if idx == m.cursor {
			nameStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF00FF"))
		}

		active := ""
		if v.id == currentVariant {
			active = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Render(" ●")
		}

		sb.WriteString(fmt.Sprintf("%s%-14s %s%s\n", cursor,
			nameStyle.Render(v.id),
			descStyle.Render(v.desc),
			active))
		idx++
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
		"  [Enter] Select  [Esc] Back"))

	return sb.String()
}

func (m Model) updateTheme(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	themes := theme.All()
	variants := []string{"default", "compact", "boxed", "powerline", "cards", "minimal"}
	maxIdx := len(themes) + 1 + len(variants) - 1 // +1 for header gap

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
		// Skip the "Variants" header line
		if m.cursor == len(themes) {
			m.cursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.cursor < maxIdx {
			m.cursor++
		}
		// Skip the "Variants" header line
		if m.cursor == len(themes) {
			m.cursor++
		}
	case key.Matches(msg, m.keys.Enter):
		if m.cursor < len(themes) {
			// Theme selection
			selected := themes[m.cursor]
			m.config.Mode.Theme = selected.ID
			m.generator.RenderOpts = render.Options{
				Theme:   selected,
				Variant: render.Variant(m.config.GlobalVariant()),
			}
			m.unsaved = true
			m.status = fmt.Sprintf("Theme → %s", selected.Name)
			m.state = StateDashboard
			m.cursor = 0
		} else if m.cursor > len(themes) {
			// Variant selection
			varIdx := m.cursor - len(themes) - 1
			if varIdx >= 0 && varIdx < len(variants) {
				m.config.Mode.Variant = variants[varIdx]
				m.generator.RenderOpts = render.Options{
					Theme:   theme.MustGet(m.config.ThemeID()),
					Variant: render.Variant(variants[varIdx]),
				}
				m.unsaved = true
				m.status = fmt.Sprintf("Style → %s", variants[varIdx])
				m.state = StateDashboard
				m.cursor = 0
			}
		}
	}

	return m, nil
}
