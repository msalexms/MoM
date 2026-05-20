package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ams/mom/internal/config"
	"github.com/ams/mom/internal/tui/components"
)

func (m Model) viewProfiles() string {
	var sb strings.Builder

	sb.WriteString(viewTitleStyle("#FFD700").Render("  :: Profiles") + "\n")
	sb.WriteString(viewSeparator() + "\n\n")

	// If in save mode, show the text input
	if m.profileSaving {
		sb.WriteString("  Enter a name for this profile:\n\n")
		sb.WriteString("  " + m.profileInput.View() + "\n\n")
		sb.WriteString(components.HelpStyle.Render("  [Enter] Save  [Esc] Cancel"))
		return sb.String()
	}

	profiles := config.ListProfiles()

	// First item is always "Save current as..."
	items := []string{"Save current config as new profile"}
	for _, p := range profiles {
		items = append(items, p)
	}

	for i, item := range items {
		active := i == m.cursor
		cursor := listCursor(active, colYellow)

		nameColor := colWhite
		if active {
			nameColor = colYellow
		}

		prefix := fixedCol("  ", 2, colGray)
		if i == 0 {
			prefix = fixedCol("+ ", 2, colYellow)
		}
		sb.WriteString(cursor + prefix + col(item, nameColor) + "\n")
	}

	if len(profiles) == 0 {
		sb.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true).Render(
			"  No saved profiles yet"))
	}

	sb.WriteString("\n\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
		"  [Enter] Load/Save  [d] Delete  [Esc] Back"))

	return sb.String()
}

func (m Model) updateProfiles(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If in save mode, handle text input
	if m.profileSaving {
		if key.Matches(msg, m.keys.Back) {
			m.profileSaving = false
			m.profileInput.Blur()
			return m, nil
		}
		if msg.Type == tea.KeyEnter {
			name := strings.TrimSpace(m.profileInput.Value())
			if name == "" {
				name = fmt.Sprintf("profile-%d", len(config.ListProfiles())+1)
			}
			// Sanitize: only allow alphanumeric, dash, underscore
			name = sanitizeProfileName(name)
			if err := config.SaveProfile(m.config, name); err != nil {
				m.errMsg = "Save failed: " + err.Error()
			} else {
				m.status = fmt.Sprintf("Profile '%s' saved!", name)
			}
			m.profileSaving = false
			m.profileInput.Blur()
			m.profileInput.SetValue("")
			m.state = StateDashboard
			m.cursor = 0
			return m, nil
		}
		var cmd tea.Cmd
		m.profileInput, cmd = m.profileInput.Update(msg)
		return m, cmd
	}

	profiles := config.ListProfiles()
	maxIdx := len(profiles) // index 0 = save, 1..n = profiles

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.cursor < maxIdx {
			m.cursor++
		}
	case key.Matches(msg, m.keys.Enter):
		if m.cursor == 0 {
			// Enter save mode — show text input for name
			m.profileSaving = true
			m.profileInput.SetValue("")
			m.profileInput.Focus()
			return m, m.profileInput.Cursor.BlinkCmd()
		} else {
			// Load profile
			profileName := profiles[m.cursor-1]
			loaded, err := config.LoadProfile(profileName)
			if err != nil {
				m.errMsg = "Load failed: " + err.Error()
			} else {
				*m.config = *loaded
				m.unsaved = true
				m.status = fmt.Sprintf("Profile '%s' loaded!", profileName)
			}
			m.state = StateDashboard
			m.cursor = 0
		}
	case msg.String() == "d":
		if m.cursor > 0 && m.cursor <= len(profiles) {
			name := profiles[m.cursor-1]
			config.DeleteProfile(name)
			m.status = fmt.Sprintf("Profile '%s' deleted", name)
			if m.cursor > len(profiles)-1 {
				m.cursor--
			}
		}
	}

	return m, nil
}

func sanitizeProfileName(name string) string {
	var sb strings.Builder
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_' {
			sb.WriteRune(ch)
		} else if ch == ' ' {
			sb.WriteRune('-')
		}
	}
	result := sb.String()
	if result == "" {
		return "unnamed"
	}
	return result
}
