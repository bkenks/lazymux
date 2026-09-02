package repomgr

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/domain"
)

// RepoDir returns the on-disk path for a repo key under baseDir.
func RepoDir(baseDir, key string) string {
	return filepath.Join(baseDir, filepath.FromSlash(key))
}

// Clone clones from the real URL into <baseDir>/<key>, then rewrites the repo
// to use a placeholder origin resolved to the origin forge via insteadOf.
// Cloning happens against the real URL so existing credentials work; the
// placeholder is applied only afterwards.
func Clone(cfg config.Config, realURL string, u RepoURL, link config.RepoLink) error {
	dest := RepoDir(cfg.BaseDir, u.Key())
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("%s already exists", dest)
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	out, err := exec.Command("git", "clone", realURL, dest).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone: %s", firstLine(string(out)))
	}
	return RenderGitConfig(cfg, u.Key(), link)
}

// RenderGitConfig makes a repo's git config match its RepoLink: origin points
// at the placeholder host, a single local insteadOf rewrites the placeholder to
// the origin forge (so fetch/pull go there), and one remote.origin.pushurl is
// written per upstream forge, which makes a push fan out to all of them. It's
// idempotent — stale lazymux-managed insteadOf rules and every existing pushurl
// are cleared first — so it can be re-run whenever the links or scheme change.
func RenderGitConfig(cfg config.Config, key string, link config.RepoLink) error {
	dir := RepoDir(cfg.BaseDir, key)
	scheme := normalizeScheme(link.Scheme)

	if err := clearManagedInsteadOf(dir, cfg.PlaceholderHost); err != nil {
		return err
	}

	origin, ok := cfg.ForgeByName(link.Origin)
	if !ok {
		return fmt.Errorf("origin forge %q not in registry", link.Origin)
	}

	phBase := hostBase(scheme, cfg.PlaceholderHost)
	originBase := hostBase(scheme, origin.Host)

	// url.<originBase>.insteadOf = <placeholderBase>
	if err := gitConfig(dir, "url."+originBase+".insteadOf", phBase); err != nil {
		return err
	}
	// origin stores the stable placeholder URL.
	placeholderURL := phBase + key + ".git"
	if err := gitConfig(dir, "remote.origin.url", placeholderURL); err != nil {
		return err
	}
	return renderPushURLs(cfg, dir, key, scheme, link)
}

// renderPushURLs replaces remote.origin.pushurl with one concrete URL per
// upstream forge. Upstreams that aren't in the registry are skipped. With a
// single upstream (the origin forge) no pushurl is written at all, so push
// follows the placeholder origin like it always has.
func renderPushURLs(cfg config.Config, dir, key, scheme string, link config.RepoLink) error {
	if err := unsetAll(dir, "remote.origin.pushurl"); err != nil {
		return err
	}
	if len(link.Upstreams) < 2 {
		return nil
	}
	for _, name := range link.Upstreams {
		forge, ok := cfg.ForgeByName(name)
		if !ok {
			continue
		}
		if err := gitConfigAdd(dir, "remote.origin.pushurl", hostBase(scheme, forge.Host)+key+".git"); err != nil {
			return err
		}
	}
	return nil
}

// clearManagedInsteadOf removes every url.<base>.insteadOf whose value points
// at our placeholder host, so switching the origin forge never leaves two rules
// competing for the same placeholder prefix.
func clearManagedInsteadOf(dir, placeholderHost string) error {
	out, err := exec.Command("git", "-C", dir, "config", "--local",
		"--get-regexp", `^url\..*\.insteadof$`).Output()
	if err != nil {
		// Exit status 1 just means no matching keys — not an error for us.
		return nil
	}
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		keyName, value, ok := strings.Cut(sc.Text(), " ")
		if !ok {
			continue
		}
		if !strings.Contains(value, placeholderHost) {
			continue
		}
		// keyName is url.<base>.insteadof — strip to the url.<base> section.
		section := strings.TrimSuffix(keyName, ".insteadof")
		_ = exec.Command("git", "-C", dir, "config", "--local",
			"--remove-section", section).Run()
	}
	return nil
}

// gitConfigAdd appends a value to a multi-valued config key.
func gitConfigAdd(dir, key, value string) error {
	out, err := exec.Command("git", "-C", dir, "config", "--local", "--add", key, value).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git config --add %s: %s", key, firstLine(string(out)))
	}
	return nil
}

// unsetAll drops every value of a config key. Exit status 5 means the key was
// already absent, which is the desired end state.
func unsetAll(dir, key string) error {
	cmd := exec.Command("git", "-C", dir, "config", "--local", "--unset-all", key)
	out, err := cmd.CombinedOutput()
	if err != nil && cmd.ProcessState.ExitCode() != 5 {
		return fmt.Errorf("git config --unset-all %s: %s", key, firstLine(string(out)))
	}
	return nil
}

func gitConfig(dir, key, value string) error {
	out, err := exec.Command("git", "-C", dir, "config", "--local", key, value).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git config %s: %s", key, firstLine(string(out)))
	}
	return nil
}

// List walks baseDir and returns every git repo found, annotated with its
// forge link from config and its local git stats. A repo is any directory
// containing a .git entry; walking stops descending once one is found, so
// nested namespaces work.
func List(cfg config.Config) ([]domain.Repo, error) {
	return list(cfg, true)
}

// ListMeta is List without the per-repo git stats, which cost three git
// subprocesses each. Callers that only need locations and forge links (the
// MCP server) should use this.
func ListMeta(cfg config.Config) ([]domain.Repo, error) {
	return list(cfg, false)
}

func list(cfg config.Config, withStats bool) ([]domain.Repo, error) {
	base := cfg.BaseDir
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return nil, nil
	}
	interactions := domain.LoadInteractions()

	var repos []domain.Repo
	err := filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable dirs rather than aborting the whole walk
		}
		if !d.IsDir() {
			return nil
		}
		if path == base {
			return nil
		}
		if _, statErr := os.Stat(filepath.Join(path, ".git")); statErr != nil {
			return nil // not a repo root; keep descending
		}
		rel, relErr := filepath.Rel(base, path)
		if relErr != nil {
			return filepath.SkipDir
		}
		key := filepath.ToSlash(rel)
		link := cfg.Repos[key]
		var stats repoStats
		if withStats {
			stats = gitStats(path)
		}
		repos = append(repos, domain.Repo{
			Name:             filepath.Base(path),
			Path:             key,
			AbsPath:          path,
			LastInteracted:   interactions[key],
			Upstreams:        link.Upstreams,
			Origin:           link.Origin,
			Scheme:           link.Scheme,
			LocalBranches:    stats.branches,
			UnpushedCommits:  stats.unpushed,
			UncommittedFiles: stats.uncommitted,
		})
		return filepath.SkipDir // don't descend into a repo
	})
	return repos, err
}

// Remove deletes a repo directory and prunes now-empty namespace parents up to
// (but not including) baseDir.
func Remove(baseDir, absPath string) error {
	if err := os.RemoveAll(absPath); err != nil {
		return err
	}
	dir := filepath.Dir(absPath)
	for dir != baseDir && strings.HasPrefix(dir, baseDir) {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			break
		}
		if err := os.Remove(dir); err != nil {
			break
		}
		dir = filepath.Dir(dir)
	}
	return nil
}

// repoStats holds the local git signals shown in the repo list.
type repoStats struct {
	branches    int
	unpushed    int
	uncommitted int
}

// gitStats inspects a repo's local state without touching the network. Failures
// (not a git repo, git missing) yield a zero-value result rather than aborting
// the whole listing.
func gitStats(dir string) repoStats {
	var s repoStats
	// Local branch count: one line per ref under refs/heads.
	if out, err := exec.Command("git", "-C", dir, "for-each-ref",
		"--format=%(refname)", "refs/heads").Output(); err == nil {
		s.branches = countLines(string(out))
	}
	// Unpushed commits: reachable from any local branch but no remote-tracking
	// ref. With no remotes configured, --remotes is empty, so every commit on a
	// branch counts — correct, since nothing is backed up anywhere.
	if out, err := exec.Command("git", "-C", dir, "rev-list", "--count",
		"--branches", "--not", "--remotes").Output(); err == nil {
		s.unpushed, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	}
	// Uncommitted files: one porcelain status line per changed path, counting
	// staged, unstaged, and untracked entries alike.
	if out, err := exec.Command("git", "-C", dir, "status", "--porcelain").Output(); err == nil {
		s.uncommitted = countLines(string(out))
	}
	return s
}

func countLines(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

func firstLine(s string) string {
	// Cut returns the whole string when there's no newline, which is what we
	// want for single-line output.
	first, _, _ := strings.Cut(strings.TrimSpace(s), "\n")
	return first
}
