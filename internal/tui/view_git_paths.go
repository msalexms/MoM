package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/msalexms/MoM/internal/module"
)

func (m Model) viewGitPaths() string {
	var sb strings.Builder

	sb.WriteString(viewTitleStyle("#FFD700").Render("  :: Git Paths") + "\n")
	sb.WriteString(viewSeparator() + "\n\n")

	paths := m.config.Modules.GitConfig.Paths

	if len(paths) == 0 {
		home, _ := os.UserHomeDir()
		sb.WriteString(dimText(fmt.Sprintf("  Using defaults: %s, %s, %s\n",
			home+"/projects",
			home+"/repos",
			home+"/src")) + "\n")
	} else {
		for i, p := range paths {
			active := i == m.gitPathCursor && !m.gitPathAdding
			cursor := listCursor(active, colYellow)

			// Check if the path exists on disk after ~ expansion
			expanded := gitPathExpand(p)
			existsMark := col(" ✗", colGray)
			if _, err := os.Stat(expanded); err == nil {
				existsMark = col(" ✓", colGreen)
			}

			pathColor := colWhite
			if active {
				pathColor = colYellow
			}

			sb.WriteString(cursor + col(p, pathColor) + existsMark + "\n")
		}
		sb.WriteString("\n")
	}

	// Max repos line
	maxRepos := m.config.Modules.GitConfig.MaxRepos
	if maxRepos <= 0 {
		maxRepos = 5
	}
	sb.WriteString(fmt.Sprintf("  Max repos shown: %s\n\n",
		col(fmt.Sprintf("%d", maxRepos), colCyan)))

	// Add-path input field
	if m.gitPathAdding {
		sb.WriteString("  New path: " + m.gitPathInput.View() + "\n\n")
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
			"  [Enter] Confirm  [Esc] Cancel"))
	} else {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
			"  [a] Add path  [x] Remove selected  [+] More repos  [-] Fewer repos  [Esc] Back"))
	}

	return sb.String()
}

func (m Model) updateGitPaths(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// --- Add-path input mode ---
	if m.gitPathAdding {
		if key.Matches(msg, m.keys.Back) {
			m.gitPathAdding = false
			m.gitPathInput.Blur()
			return m, nil
		}
		if msg.Type == tea.KeyEnter {
			val := strings.TrimSpace(m.gitPathInput.Value())
			if val != "" {
				m.config.Modules.GitConfig.Paths = append(
					m.config.Modules.GitConfig.Paths, val)
				m.syncGitModule()
				m.unsaved = true
			}
			m.gitPathAdding = false
			m.gitPathInput.Blur()
			m.gitPathInput.SetValue("")
			return m, nil
		}
		var cmd tea.Cmd
		m.gitPathInput, cmd = m.gitPathInput.Update(msg)
		return m, cmd
	}

	// --- Normal navigation mode ---
	paths := m.config.Modules.GitConfig.Paths
	maxIdx := len(paths) - 1

	maxRepos := m.config.Modules.GitConfig.MaxRepos
	if maxRepos <= 0 {
		maxRepos = 5
	}

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.gitPathCursor > 0 {
			m.gitPathCursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.gitPathCursor < maxIdx {
			m.gitPathCursor++
		}
	case msg.String() == "a":
		m.gitPathAdding = true
		m.gitPathInput.Focus()
		return m, m.gitPathInput.Cursor.BlinkCmd()
	case msg.String() == "x":
		if len(paths) > 0 && m.gitPathCursor <= maxIdx {
			m.config.Modules.GitConfig.Paths = append(
				paths[:m.gitPathCursor],
				paths[m.gitPathCursor+1:]...)
			if m.gitPathCursor > 0 && m.gitPathCursor >= len(m.config.Modules.GitConfig.Paths) {
				m.gitPathCursor--
			}
			m.syncGitModule()
			m.unsaved = true
		}
	case msg.String() == "+":
		m.config.Modules.GitConfig.MaxRepos = maxRepos + 1
		m.syncGitModule()
		m.unsaved = true
	case msg.String() == "-":
		if maxRepos > 1 {
			m.config.Modules.GitConfig.MaxRepos = maxRepos - 1
			m.syncGitModule()
			m.unsaved = true
		}
	}

	return m, nil
}

// syncGitModule propagates the current config values into the live module
// instance so previews reflect the changes immediately.
func (m *Model) syncGitModule() {
	if mod, ok := m.registry.Get("git"); ok {
		if gitMod, ok := mod.(*module.GitStatusModule); ok {
			gitMod.Paths = m.config.Modules.GitConfig.Paths
			gitMod.MaxRepos = m.config.Modules.GitConfig.MaxRepos
		}
	}
}

// gitPathExpand replaces a leading ~/ with the user's home directory.
func gitPathExpand(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return home + "/" + p[2:]
		}
	}
	return p
}
