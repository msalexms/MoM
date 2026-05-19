package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ams/mom/internal/distro"
)

func (m Model) viewRollback() string {
	var sb strings.Builder

	sb.WriteString(viewTitleStyle("#FFD700").Render("  :: Rollback") + "\n")
	sb.WriteString(viewSeparator() + "\n\n")

	if len(m.backups) == 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true).Render(
			"  No backups available") + "\n")
	} else {
		for i, b := range m.backups {
			cursor := "  "
			if i == m.cursor {
				cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true).Render("▸ ")
			}

			label := b.Timestamp.Format("2006-01-02 15:04:05")
			labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
			if i == m.cursor {
				labelStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFD700"))
			}

			extra := ""
			if b.IsOriginal {
				extra = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Render(" [ORIGINAL]")
			}
			if b.Distro != "" && b.Distro != "pre-rollback" {
				extra += lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
					fmt.Sprintf(" (%s)", b.Distro))
			}

			sb.WriteString(fmt.Sprintf("%s%s%s\n", cursor, labelStyle.Render(label), extra))
		}
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
		"  [Enter] Restore  [Esc] Back"))

	return sb.String()
}

func (m Model) updateRollback(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxIdx := len(m.backups) - 1

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
		if m.cursor < len(m.backups) {
			b := m.backups[m.cursor]
			paths := distro.GetPaths(m.distroInfo.Family)
			err := m.backupMgr.Restore(context.Background(), &b, paths.MotdFile)
			if err != nil {
				m.errMsg = "Rollback failed: " + err.Error()
			} else {
				m.status = "Rollback successful!"
			}
			m.state = StateDashboard
			m.cursor = 0
		}
	}

	return m, nil
}
