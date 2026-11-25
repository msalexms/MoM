package module

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/msalexms/MoM/internal/i18n"

	"github.com/msalexms/MoM/internal/config"
	"github.com/msalexms/MoM/internal/module/render"
)

// GitStatusModule displays git repos with uncommitted changes.
type GitStatusModule struct {
	// Paths lists the directories to scan for git repos.
	// Each entry may start with ~/ which is expanded to $HOME at scan time.
	// When empty, defaults to the most common developer directory names.
	Paths    []string
	MaxRepos int // maximum dirty repos to display; 0 means use default (5)
}

func (m *GitStatusModule) Name() string  { return "git" }
func (m *GitStatusModule) Title() string { return "Git status" }

// UpdateFromConfig implements ConfigSync.
func (m *GitStatusModule) UpdateFromConfig(cfg *config.Config) {
	m.Paths = cfg.Modules.GitConfig.Paths
	m.MaxRepos = cfg.Modules.GitConfig.MaxRepos
}
func (m *GitStatusModule) Description() string    { return "Repos with uncommitted changes" }
func (m *GitStatusModule) Dependencies() []string { return []string{"git"} }
func (m *GitStatusModule) Available() bool        { return CheckDependency("git") }
func (m *GitStatusModule) Settings() []Setting {
	return []Setting{
		listSetting("paths", "Paths", "Folders to scan for repos (empty: ~/projects, ~/repos, ~/src…)",
			func(c *config.Config) *[]string { return &c.Modules.GitConfig.Paths }),
		intSetting("max_repos", "Max repos", "Maximum number of repos with changes to show",
			func(c *config.Config) *int { return &c.Modules.GitConfig.MaxRepos }),
	}
}

// defaultGitPaths are scanned when no path is configured.
var defaultGitPaths = []string{"~/projects", "~/repos", "~/src", "~/code", "~/dev", "~/workspace", "~/git"}

type repoStatus struct {
	name   string
	dirty  int
	branch string
}

func (m *GitStatusModule) Generate(ctx context.Context, opts render.Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	repos := m.scanRepos(ctx)
	if len(repos) == 0 {
		return "", nil
	}

	r := render.New(opts)
	th := r.Theme()

	var lines []string
	for _, repo := range repos {
		lines = append(lines, fmt.Sprintf("%s  %s  %s",
			r.Col(repo.name, 16, th.Palette.Warning),
			th.Color(repo.branch, th.Palette.Accent),
			th.Color(fmt.Sprintf(i18n.T("%d changes"), repo.dirty), th.Palette.Danger)))
	}

	compact := fmt.Sprintf(i18n.T("%d dirty repos"), len(repos))
	return r.Section("Git", "git", compact, lines), nil
}

func (m *GitStatusModule) scanRepos(ctx context.Context) []repoStatus {
	paths := m.Paths
	if len(paths) == 0 {
		paths = defaultGitPaths
	}

	maxRepos := m.MaxRepos
	if maxRepos <= 0 {
		maxRepos = 5
	}

	var repos []repoStatus
	for _, base := range paths {
		base = ExpandHomePath(base)
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			dir := filepath.Join(base, e.Name())
			if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
				continue
			}
			dirty := gitDirtyCount(ctx, dir)
			if dirty == 0 {
				continue
			}
			branch := gitBranch(ctx, dir)
			repos = append(repos, repoStatus{e.Name(), dirty, branch})
			if len(repos) >= maxRepos {
				return repos
			}
		}
	}
	return repos
}

// ExpandHomePath replaces a leading ~/ with the user's home directory.
// It is the single ~-expansion helper, used by the git scan and by the TUI
// when validating configured paths.
func ExpandHomePath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

func gitDirtyCount(ctx context.Context, dir string) int {
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0
	}
	return len(lines)
}

func gitBranch(ctx context.Context, dir string) string {
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "branch", "--show-current")
	out, err := cmd.Output()
	if err != nil {
		return "?"
	}
	return strings.TrimSpace(string(out))
}
