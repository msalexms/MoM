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
	Paths []string // directories to scan; defaults to ~/projects, ~/repos
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
func (m *GitStatusModule) Settings() []SettingDef         { return nil }

func (m *GitStatusModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

type repoStatus struct {
	name    string
	dirty   int
	branch  string
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
	var sb strings.Builder

	switch r.Variant() {
	case render.VariantCompact:
		sb.WriteString(r.Header("Git", "git"))
		sb.WriteString(fmt.Sprintf("\n    %d dirty repos", len(repos)))
	case render.VariantBoxed:
		var c strings.Builder
		for _, repo := range repos {
			c.WriteString(fmt.Sprintf("%-16s  %s  %s\n", repo.name,
				th.Color(repo.branch, th.Palette.Accent),
				th.Color(fmt.Sprintf("%d changes", repo.dirty), th.Palette.Warning)))
		}
		sb.WriteString(render.Indent(r.Box(strings.TrimRight(c.String(), "\n"), "Git Status"), "  "))
	default:
		sb.WriteString(r.Header("Git Status", "git"))
		sb.WriteString("\n\n")
		for _, repo := range repos {
			sb.WriteString(fmt.Sprintf("    %-16s  %s  %s\n",
				th.Color(repo.name, th.Palette.Warning),
				th.Color(repo.branch, th.Palette.Accent),
				th.Color(fmt.Sprintf("%d changes", repo.dirty), th.Palette.Danger)))
		}
	}
	return sb.String(), nil
}

func (m *GitStatusModule) scanRepos(ctx context.Context) []repoStatus {
	paths := m.Paths
	if len(paths) == 0 {
		home, _ := os.UserHomeDir()
		paths = []string{
			filepath.Join(home, "projects"),
			filepath.Join(home, "repos"),
			filepath.Join(home, "src"),
		}
	}

	var repos []repoStatus
	for _, base := range paths {
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
			if len(repos) >= 5 {
				return repos
			}
		}
	}
	return repos
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
