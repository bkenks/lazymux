package repomgr

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bkenks/lazymux/internal/config"
)

func initRepo(t *testing.T, base, key string) string {
	t.Helper()
	dir := RepoDir(base, key)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	mustGit(t, dir, "init", "-q")
	return dir
}

func TestParseRepoURLRejectsUnsafePaths(t *testing.T) {
	for _, raw := range []string{
		"https://github.com/../../etc/repo",
		"https://github.com/a//b",
		"https://github.com/a/./b",
		"git@github.com:../repo.git",
	} {
		if u, err := ParseRepoURL(raw); err == nil {
			t.Errorf("%s: parsed as %+v, want an error", raw, u)
		}
	}
}

func TestRenderGitConfigLeavesRepoAloneWhenOriginIsMissing(t *testing.T) {
	base := t.TempDir()
	dir := initRepo(t, base, "me/demo")
	cfg := config.Config{
		BaseDir:         base,
		PlaceholderHost: config.DefaultPlaceholderHost,
		Forges:          []config.Forge{{Name: "github", Host: "github.com"}},
	}
	link := config.RepoLink{Upstreams: []string{"github"}, Origin: "github"}
	if err := RenderGitConfig(cfg, "me/demo", link); err != nil {
		t.Fatal(err)
	}

	missing := config.RepoLink{Upstreams: []string{"gone"}, Origin: "gone"}
	err := RenderGitConfig(cfg, "me/demo", missing)
	if err == nil {
		t.Fatal("expected an error for an origin forge missing from the registry")
	}
	got := gitCfg(t, dir, "url.https://github.com/.insteadOf")
	if got != "https://lazymux-placeholder/" {
		t.Errorf("working insteadOf rule removed, now %q", got)
	}
}

func TestRenderGitConfigKeepsUnrelatedInsteadOfRules(t *testing.T) {
	base := t.TempDir()
	dir := initRepo(t, base, "me/demo")
	mustGit(t, dir, "config", "--local",
		"url.https://mirror.example/.insteadOf", "https://lazymux-placeholder2.example/")
	cfg := config.Config{
		BaseDir:         base,
		PlaceholderHost: config.DefaultPlaceholderHost,
		Forges:          []config.Forge{{Name: "github", Host: "github.com"}},
	}

	link := config.RepoLink{Upstreams: []string{"github"}, Origin: "github"}
	if err := RenderGitConfig(cfg, "me/demo", link); err != nil {
		t.Fatal(err)
	}
	if got := gitCfg(t, dir, "url.https://mirror.example/.insteadOf"); got == "" {
		t.Error("a rule that merely contains the placeholder host was removed")
	}
}

func TestRenderGitConfigReportsGitFailures(t *testing.T) {
	cfg := config.Config{
		BaseDir:         t.TempDir(),
		PlaceholderHost: config.DefaultPlaceholderHost,
		Forges:          []config.Forge{{Name: "github", Host: "github.com"}},
	}
	link := config.RepoLink{Upstreams: []string{"github"}, Origin: "github"}
	err := RenderGitConfig(cfg, "me/missing", link)
	if err == nil {
		t.Fatal("expected an error for a repo directory that doesn't exist")
	}
}

func TestRemoveRefusesPathsOutsideBaseDir(t *testing.T) {
	root := t.TempDir()
	base := filepath.Join(root, "lazy")
	outside := filepath.Join(root, "lazymux-other", "repo")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	escaping := filepath.Join(base, "..", "lazymux-other", "repo")
	for _, target := range []string{outside, base, escaping} {
		if err := Remove(base, target); err == nil {
			t.Errorf("Remove(%q, %q) succeeded", base, target)
		}
	}
	if _, err := os.Stat(outside); err != nil {
		t.Errorf("directory outside base was deleted: %v", err)
	}
}

func TestRemovePrunesEmptyNamespaces(t *testing.T) {
	base := t.TempDir()
	dir := initRepo(t, base, "group/sub/repo")

	if err := Remove(base, dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(base, "group")); !os.IsNotExist(err) {
		t.Errorf("empty namespace dirs left behind: %v", err)
	}
	if _, err := os.Stat(base); err != nil {
		t.Errorf("base dir removed: %v", err)
	}
}

func TestListFollowsSymlinkedBaseDir(t *testing.T) {
	target := t.TempDir()
	initRepo(t, target, "me/demo")
	link := filepath.Join(t.TempDir(), "lazymux")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	repos, err := ListMeta(config.Config{BaseDir: link})
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 || repos[0].Path != "me/demo" {
		t.Fatalf("repos = %+v", repos)
	}
	if want := filepath.Join(link, "me", "demo"); repos[0].AbsPath != want {
		t.Errorf("AbsPath = %q, want it under the configured base %q", repos[0].AbsPath, want)
	}
}
