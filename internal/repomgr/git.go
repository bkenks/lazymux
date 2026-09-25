package repomgr

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/domain"
)

// RepoDir returns the on-disk path for a repo key under baseDir.
func RepoDir(baseDir, key string) string {
	return filepath.Join(baseDir, filepath.FromSlash(key))
}

// RenderGitConfig makes a repo's git config match its RepoLink: origin points
// at the placeholder host, a single local insteadOf rewrites the placeholder to
// the origin forge (so fetch/pull go there), and one remote.origin.pushurl is
// written per upstream forge, which makes a push fan out to all of them. It's
// idempotent — stale lazymux-managed insteadOf rules and every existing pushurl
// are cleared first — so it can be re-run whenever the links or scheme change.
// The origin forge is looked up before anything is changed, so a missing forge
// leaves the repo's existing config working.
func RenderGitConfig(cfg config.Config, key string, link config.RepoLink) error {
	dir := RepoDir(cfg.RepoRoot(), key)
	scheme := config.NormalizeScheme(link.Scheme)

	origin, ok := cfg.ForgeByName(link.Origin)
	if !ok {
		return fmt.Errorf("origin forge %q not in registry", link.Origin)
	}

	if err := clearManagedInsteadOf(dir, cfg.PlaceholderHost); err != nil {
		return err
	}

	phBase := hostBase(scheme, cfg.PlaceholderHost)
	originBase := hostBase(scheme, origin.Host)

	// url.<originBase>.insteadOf = <placeholderBase>
	insteadOfKey := "url." + originBase + ".insteadOf"
	if _, err := runGit(dir, "config", "--local", insteadOfKey, phBase); err != nil {
		return err
	}
	// origin stores the stable placeholder URL.
	placeholderURL := RemoteURL(scheme, cfg.PlaceholderHost, key)
	if _, err := runGit(dir, "config", "--local", "remote.origin.url", placeholderURL); err != nil {
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
		pushURL := RemoteURL(scheme, forge.Host, key)
		_, err := runGit(dir, "config", "--local", "--add", "remote.origin.pushurl", pushURL)
		if err != nil {
			return err
		}
	}
	return nil
}

// clearManagedInsteadOf removes every url.<base>.insteadOf whose value is our
// placeholder prefix (for either scheme), so switching the origin forge never
// leaves two rules competing for the same placeholder prefix.
func clearManagedInsteadOf(dir, placeholderHost string) error {
	managed := []string{
		hostBase(config.SchemeHTTPS, placeholderHost),
		hostBase(config.SchemeSSH, placeholderHost),
	}
	out, err := runGit(dir, "config", "--local", "--get-regexp", `^url\..*\.insteadof$`)
	if exitCode(err) == 1 {
		return nil // no insteadOf rules at all
	}
	if err != nil {
		return err
	}
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		keyName, value, ok := strings.Cut(sc.Text(), " ")
		if !ok || !slices.Contains(managed, value) {
			continue
		}
		// keyName is url.<base>.insteadof — strip to the url.<base> section.
		section := strings.TrimSuffix(keyName, ".insteadof")
		if _, err := runGit(dir, "config", "--local", "--remove-section", section); err != nil {
			return err
		}
	}
	return nil
}

// unsetAll drops every value of a config key. Exit status 5 means the key was
// already absent, which is the desired end state.
func unsetAll(dir, key string) error {
	_, err := runGit(dir, "config", "--local", "--unset-all", key)
	if exitCode(err) == 5 {
		return nil
	}
	return err
}

// runGit runs git with args in dir and returns its combined output. A failure
// is reported as the git subcommand plus the first line git printed; the
// underlying *exec.ExitError stays reachable through errors.As.
func runGit(dir string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		subcommand := strings.Join(args[:min(len(args), 3)], " ")
		return string(out), fmt.Errorf("git %s: %s: %w", subcommand, FirstLine(string(out)), err)
	}
	return string(out), nil
}

// exitCode returns the exit status carried by an error from runGit, or -1 if
// err is nil or git didn't run to completion.
func exitCode(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
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
	if err := cfg.ValidateRepoRoot(); err != nil {
		return nil, err
	}
	base := cfg.RepoRoot()
	// WalkDir doesn't follow a symlinked root, so walk its target and report
	// paths under base as configured.
	root, err := filepath.EvalSymlinks(base)
	if err != nil {
		return nil, fmt.Errorf("resolving %s: %w", base, err)
	}
	interactions := domain.LoadInteractions()

	var repos []domain.Repo
	err = filepath.WalkDir(root, func(walked string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable dirs rather than aborting the whole walk
		}
		if !d.IsDir() || walked == root {
			return nil
		}
		if _, statErr := os.Stat(filepath.Join(walked, ".git")); statErr != nil {
			return nil // not a repo root; keep descending
		}
		rel, relErr := filepath.Rel(root, walked)
		if relErr != nil {
			return filepath.SkipDir
		}
		path := filepath.Join(base, rel)
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
// (but not including) baseDir. It refuses any path that isn't strictly inside
// baseDir.
func Remove(baseDir, absPath string) error {
	if !isInside(baseDir, absPath) {
		return fmt.Errorf("refusing to delete %s: not inside %s", absPath, baseDir)
	}
	if err := os.RemoveAll(absPath); err != nil {
		return err
	}
	dir := filepath.Dir(absPath)
	for isInside(baseDir, dir) {
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

// isInside reports whether path is strictly below dir.
func isInside(dir, path string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	return rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
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

// FirstLine returns the first line of s with surrounding whitespace trimmed,
// for turning a command's output into a one-line error.
func FirstLine(s string) string {
	first, _, _ := strings.Cut(strings.TrimSpace(s), "\n")
	return first
}
