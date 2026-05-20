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

	idx := 0
	for i, th := range themes {
		active := idx == m.cursor
		cursor := listCursor(active, colMagenta)

		nameColor := colWhite
		if active {
			nameColor = colMagenta
		}
		name := fixedCol(th.Name, 18, nameColor)
		desc := dimText(th.Description)

		indicator := ""
		if th.ID == currentID {
			indicator = " " + colGreen + "●" + colReset
		}

		sb.WriteString(cursor + name + " " + desc + indicator + "\n")
		_ = i
		idx++
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CCCCCC")).Render("  Style Variant") + "\n")
	idx++

	for _, v := range variants {
		active := idx == m.cursor
		cursor := listCursor(active, colMagenta)

		nameColor := colWhite
		if active {
			nameColor = colMagenta
		}
		name := fixedCol(v.id, 18, nameColor)
		desc := dimText(v.desc)

		indicator := ""
		if v.id == currentVariant {
			indicator = " " + colGreen + "●" + colReset
		}

		sb.WriteString(cursor + name + " " + desc + indicator + "\n")
		idx++
	}

	// Live preview: show a mini sample with the hovered theme/variant
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CCCCCC")).Render("  Preview") + "\n")
	preview := m.themePreviewSample()
	sb.WriteString(preview)

	sb.WriteString("\n\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
		"  [Enter] Select  [Esc] Back"))

	return sb.String()
}

// themePreviewSample generates a small MOTD sample using the currently
// hovered theme/variant so the user can see how it looks before selecting.
func (m Model) themePreviewSample() string {
	themes := theme.All()
	variantIDs := []string{"default", "compact", "boxed", "powerline", "cards", "minimal"}

	// Determine which theme/variant is hovered
	previewTheme := theme.MustGet(m.config.ThemeID())
	previewVariant := render.Variant(m.config.GlobalVariant())

	if m.cursor < len(themes) {
		previewTheme = themes[m.cursor]
	} else if m.cursor > len(themes) {
		varIdx := m.cursor - len(themes) - 1
		if varIdx >= 0 && varIdx < len(variantIDs) {
			previewVariant = render.Variant(variantIDs[varIdx])
		}
	}

	opts := render.Options{Theme: previewTheme, Variant: previewVariant}
	r := render.New(opts)

	// Render a small sample
	var sb strings.Builder
	sb.WriteString("  " + r.Header("System", "system") + "\n")
	sb.WriteString("  " + r.KeyValue("host", "myserver") + "\n")
	sb.WriteString("  " + r.KeyValue("uptime", "3d 14h") + "\n")
	sb.WriteString("  " + fmt.Sprintf("    %-10s  %s", "ram", r.ProgressBar(67.3, 16, "")))
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
