package repomgr

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/domain"
)

// RepoDir returns the on-disk path for a repo key under baseDir.
func RepoDir(baseDir, key string) string {
	return filepath.Join(baseDir, filepath.FromSlash(key))
}

// Clone clones from the real URL into <baseDir>/<key>, then rewrites the repo
// to use a placeholder origin resolved to the primary forge via insteadOf.
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
// at the placeholder host, and a single local insteadOf rewrites the
// placeholder to the primary forge. It's idempotent — any stale
// lazymux-managed insteadOf rules are cleared first — so it can be re-run
// whenever the primary forge or scheme changes.
func RenderGitConfig(cfg config.Config, key string, link config.RepoLink) error {
	dir := RepoDir(cfg.BaseDir, key)
	scheme := normalizeScheme(link.Scheme)

	if err := clearManagedInsteadOf(dir, cfg.PlaceholderHost); err != nil {
		return err
	}

	forge, ok := cfg.ForgeByName(link.Primary)
	if !ok {
		return fmt.Errorf("primary forge %q not in registry", link.Primary)
	}

	phBase := hostBase(scheme, cfg.PlaceholderHost)
	forgeBase := hostBase(scheme, forge.Host)

	// url.<forgeBase>.insteadOf = <placeholderBase>
	if err := gitConfig(dir, "url."+forgeBase+".insteadOf", phBase); err != nil {
		return err
	}
	// origin stores the stable placeholder URL.
	placeholderURL := phBase + key + ".git"
	return gitConfig(dir, "remote.origin.url", placeholderURL)
}

// clearManagedInsteadOf removes every url.<base>.insteadOf whose value points
// at our placeholder host, so switching primaries never leaves two rules
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
		line := sc.Text()
		sp := strings.IndexByte(line, ' ')
		if sp < 0 {
			continue
		}
		keyName, value := line[:sp], line[sp+1:]
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

func gitConfig(dir, key, value string) error {
	out, err := exec.Command("git", "-C", dir, "config", "--local", key, value).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git config %s: %s", key, firstLine(string(out)))
	}
	return nil
}

// List walks baseDir and returns every git repo found, annotated with its
// forge link from config. A repo is any directory containing a .git entry;
// walking stops descending once one is found, so nested namespaces work.
func List(cfg config.Config) ([]domain.Repo, error) {
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
		repos = append(repos, domain.Repo{
			Name:           filepath.Base(path),
			Path:           key,
			AbsPath:        path,
			LastInteracted: interactions[key],
			Forges:         link.Forges,
			Primary:        link.Primary,
			Scheme:         link.Scheme,
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

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
