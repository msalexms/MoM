package module

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ams/mom/internal/module/render"
)

// GitStatusModule displays git repos with uncommitted changes.
type GitStatusModule struct {
	// Paths lists the directories to scan for git repos.
	// Each entry may start with ~/ which is expanded to $HOME at scan time.
	// When empty, defaults to the most common developer directory names.
	Paths    []string
	MaxRepos int // maximum dirty repos to display; 0 means use default (5)
}

func (m *GitStatusModule) Name() string           { return "git" }
func (m *GitStatusModule) Title() string          { return "Git Status" }
func (m *GitStatusModule) Description() string    { return "Repos with uncommitted changes" }
func (m *GitStatusModule) Dependencies() []string { return []string{"git"} }
func (m *GitStatusModule) Available() bool        { return CheckDependency("git") }
func (m *GitStatusModule) DefaultEnabled() bool   { return false }
func (m *GitStatusModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantCompact, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *GitStatusModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *GitStatusModule) Settings() []SettingDef {
	return []SettingDef{
		{
			Key:         "paths",
			Label:       "Scan Paths",
			Type:        SettingString,
			Default:     "~/projects, ~/repos, ~/src, ~/code, ~/dev, ~/workspace, ~/git",
			Description: "Directories to scan for git repos (configure via Git Paths in the TUI)",
		},
		{
			Key:         "max_repos",
			Label:       "Max Repos",
			Type:        SettingInt,
			Default:     5,
			Description: "Maximum number of dirty repos to display",
		},
	}
}

func (m *GitStatusModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

type repoStatus struct {
	name   string
	dirty  int
	branch string
}

func (m *GitStatusModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
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
		lines = append(lines, fmt.Sprintf("%-16s  %s  %s",
			th.Color(repo.name, th.Palette.Warning),
			th.Color(repo.branch, th.Palette.Accent),
			th.Color(fmt.Sprintf("%d changes", repo.dirty), th.Palette.Danger)))
	}

	compact := fmt.Sprintf("%d dirty repos", len(repos))
	return r.Section("Git Status", "git", compact, lines), nil
}

func (m *GitStatusModule) scanRepos(ctx context.Context) []repoStatus {
	paths := m.Paths
	if len(paths) == 0 {
		home, _ := os.UserHomeDir()
		paths = []string{
			filepath.Join(home, "projects"),
			filepath.Join(home, "repos"),
			filepath.Join(home, "src"),
			filepath.Join(home, "code"),
			filepath.Join(home, "dev"),
			filepath.Join(home, "workspace"),
			filepath.Join(home, "git"),
		}
	}

	maxRepos := m.MaxRepos
	if maxRepos <= 0 {
		maxRepos = 5
	}

	var repos []repoStatus
	for _, base := range paths {
		base = expandHomePath(base)
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

// expandHomePath replaces a leading ~/ with the user's home directory.
func expandHomePath(p string) string {
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
