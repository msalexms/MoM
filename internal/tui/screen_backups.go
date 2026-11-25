package tui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/msalexms/MoM/internal/backup"
	"github.com/msalexms/MoM/internal/config"
	"github.com/msalexms/MoM/internal/generator"
	c "github.com/msalexms/MoM/internal/tui/components"
)

// backupPreview is what the right-hand panel shows for one backup.
type backupPreview struct {
	loading bool
	content string // rendered MOTD, or the stored text for text backups
	text    bool   // content is stored text (original or older version), not a config
	err     error
}

// reloadBackups refreshes the list and finds which entry is installed now.
func (m *Model) reloadBackups() {
	list, err := m.backups.List()
	if err != nil {
		m.fail(err)
	}
	m.backupList = list
	m.setCursor(min(m.cursor(), max(0, len(list)-1)))

	m.currentBackup = ""
	installed := m.writer.InstalledConfig()
	if installed == nil {
		return
	}
	for _, b := range list {
		if b.IsOriginal {
			continue
		}
		if data, err := m.backups.Read(&b); err == nil && bytes.Equal(data, installed) {
			m.currentBackup = b.Path
			return
		}
	}
}

// loadBackupPreview renders the highlighted backup in the background,
// once per backup: a config is rendered with fresh modules exactly as the
// login hook would, a text backup is shown as it is.
func (m Model) loadBackupPreview() tea.Cmd {
	if len(m.backupList) == 0 {
		return nil
	}
	b := m.backupList[m.cursor()]
	if _, ok := m.backupPreviews[b.Path]; ok {
		return nil
	}
	m.backupPreviews[b.Path] = backupPreview{loading: true}

	w, bm, newRegistry := m.writer, m.backups, m.newRegistry
	return func() tea.Msg {
		msg := backupPreviewMsg{path: b.Path}
		cfg, err := w.BackupConfig(&b)
		switch {
		case errors.Is(err, generator.ErrLegacyBackup):
			data, err := bm.Read(&b)
			msg.content, msg.text, msg.err = string(data), true, err
		case err != nil:
			msg.err = err
		default:
			reg := newRegistry()
			reg.SyncConfig(cfg)
			msg.content, msg.err = generator.NewGenerator(reg, cfg).Generate(context.Background())
		}
		return msg
	}
}

func (m Model) updateBackups(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.moveCursor(msg, len(m.backupList)) {
		return m, m.loadBackupPreview()
	}
	if !key.Matches(msg, m.keys.Enter) || len(m.backupList) == 0 {
		return m, nil
	}

	b := m.backupList[m.cursor()]
	w := m.writer
	if b.IsOriginal {
		return m.privileged(T("Restoring the original MOTD…"), func(ctx context.Context) result {
			if err := w.Uninstall(ctx); err != nil {
				return result{err: err}
			}
			return result{text: T("mom removed from login: the original MOTD is back."), reload: true}
		})
	}
	if b.Path == m.currentBackup {
		m.ok("%s", T("This is the MOTD installed now."))
		return m, nil
	}
	if _, err := w.BackupConfig(&b); err != nil {
		m.fail(errors.New(T("This backup comes from an older version of mom: it can be viewed but not installed.")))
		return m, nil
	}

	when := b.Timestamp.Format("02/01/2006 15:04")
	return m.privileged(T("Installing the backup…"), func(ctx context.Context) result {
		cfg, err := w.Reinstall(ctx, &b)
		if err != nil {
			return result{err: err}
		}
		if err := config.Save(cfg); err != nil {
			return result{err: err}
		}
		return result{text: Tf("Installed the MOTD from %s.", when), restored: cfg, saved: true, reload: true}
	})
}

func (m Model) viewBackups() string {
	if len(m.backupList) == 0 {
		return c.Muted.Render(T("No backups yet: every install is saved here."))
	}

	// Left: the history.
	const listW = 34
	var list strings.Builder
	start, end := c.Window(len(m.backupList), m.cursor(), m.listHeight(2))
	for i := start; i < end; i++ {
		b := m.backupList[i]
		active := i == m.cursor()
		tag := ""
		switch {
		case b.IsOriginal:
			tag = c.Yellow.Render(T("original"))
		case b.Path == m.currentBackup:
			tag = c.OK.Render(T("● installed"))
		case len(b.Modules) > 0:
			tag = c.Muted.Render(Tf("%d modules", len(b.Modules)))
		}
		list.WriteString(c.Cursor(active) + c.Pad(c.Item(b.Timestamp.Format("02/01/2006 15:04"), active, true), 18) + tag + "\n")
	}
	if info := c.ScrollInfo(start, end, len(m.backupList)); info != "" {
		list.WriteString("\n" + info)
	}

	// Right: what the highlighted backup is, and how it looks.
	b := m.backupList[m.cursor()]
	p := m.backupPreviews[b.Path]
	rightW := max(20, m.contentWidth()-listW-2)
	var about, action string
	switch {
	case b.IsOriginal:
		about = T("The /etc/motd from before mom.")
		action = T("Enter removes mom from login and puts it back.")
	case b.Path == m.currentBackup:
		about = T("Installed now.")
	case p.text:
		about = T("Copy from an older version of mom (view only).")
	default:
		about = backupSummary(b)
		action = T("Enter installs this one again.")
	}

	var right strings.Builder
	right.WriteString(c.Text2.Width(rightW).Render(about) + "\n")
	if action != "" {
		right.WriteString(c.Muted.Width(rightW).Render(action) + "\n")
	}
	right.WriteString("\n")
	body := ""
	switch {
	case p.loading:
		body = c.Muted.Render(T("Generating the preview…"))
	case p.err != nil:
		body = c.Bad.Render(p.err.Error())
	case strings.TrimSpace(p.content) == "":
		body = c.Muted.Render(T("(empty)"))
	default:
		body = p.content
	}
	right.WriteString(lipgloss.NewStyle().MaxWidth(rightW).MaxHeight(max(3, m.listHeight(4))).Render(body))

	return lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(listW).Render(list.String()), "  ", right.String())
}

func backupSummary(b backup.Backup) string {
	return fmt.Sprintf(T("Installed on %s · %d modules"), b.Timestamp.Format("02/01/2006 15:04"), len(b.Modules))
}
