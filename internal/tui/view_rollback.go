package tui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/msalexms/MoM/internal/distro"
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
			active := i == m.cursor
			cursor := listCursor(active, colYellow)

			labelColor := colWhite
			if active {
				labelColor = colYellow
			}
			label := fixedCol(b.Timestamp.Format("2006-01-02 15:04:05"), 21, labelColor)

			extra := ""
			if b.IsOriginal {
				extra = " " + colGreen + colBold + "[ORIGINAL]" + colReset
			}
			if b.Distro != "" && b.Distro != "pre-rollback" {
				extra += " " + dimText(b.Distro)
			}

			sb.WriteString(cursor + label + extra + "\n")
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
