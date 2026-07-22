package repomgr

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bkenks/lazymux/internal/config"
)

func TestParseRepoURL(t *testing.T) {
	cases := []struct {
		in                          string
		scheme, host, ns, name, key string
	}{
		{"https://github.com/bkenks/lazymux.git", "https", "github.com", "bkenks", "lazymux", "bkenks/lazymux"},
		{"https://github.com/bkenks/lazymux", "https", "github.com", "bkenks", "lazymux", "bkenks/lazymux"},
		{"git@github.com:bkenks/lazymux.git", "ssh", "github.com", "bkenks", "lazymux", "bkenks/lazymux"},
		{"ssh://git@fj.homektb.com/bkenks/lazymux.git", "ssh", "fj.homektb.com", "bkenks", "lazymux", "bkenks/lazymux"},
		{"ssh://git@fj.homektb.com:2222/bkenks/lazymux.git", "ssh", "fj.homektb.com", "bkenks", "lazymux", "bkenks/lazymux"},
		{"https://gitlab.com/group/subgroup/proj.git", "https", "gitlab.com", "group/subgroup", "proj", "group/subgroup/proj"},
	}
	for _, c := range cases {
		u, err := ParseRepoURL(c.in)
		if err != nil {
			t.Fatalf("%s: %v", c.in, err)
		}
		if u.Scheme != c.scheme || u.Host != c.host || u.Namespace != c.ns || u.Name != c.name {
			t.Errorf("%s: got %+v", c.in, u)
		}
		if u.Key() != c.key {
			t.Errorf("%s: key=%q want %q", c.in, u.Key(), c.key)
		}
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	full := append([]string{"-C", dir}, args...)
	if out, err := exec.Command("git", full...).CombinedOutput(); err != nil {
		t.Fatalf("git %s: %s", strings.Join(args, " "), out)
	}
}

func TestGitStats(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@example.com")
	runGit(t, dir, "config", "user.name", "T")

	// Fresh repo, no commits yet: no branches, nothing unpushed, clean.
	if s := gitStats(dir); s.branches != 0 || s.unpushed != 0 || s.uncommitted != 0 {
		t.Fatalf("empty repo: got %+v", s)
	}

	// One commit on the default branch, no remote: 1 branch, 1 unpushed, clean.
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "a.txt")
	runGit(t, dir, "commit", "-m", "first")
	if s := gitStats(dir); s.branches != 1 || s.unpushed != 1 || s.uncommitted != 0 {
		t.Fatalf("after commit: got %+v", s)
	}

	// Second branch, one modified file and one untracked file: 2 branches, 2
	// uncommitted paths. Untracked files count too — a wipe loses them as well.
	runGit(t, dir, "branch", "feature")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	if s := gitStats(dir); s.branches != 2 || s.uncommitted != 2 {
		t.Fatalf("two branches + two uncommitted: got %+v", s)
	}
}

func gitCfg(t *testing.T, dir, key string) string {
	t.Helper()
	out, _ := exec.Command("git", "-C", dir, "config", "--local", "--get", key).Output()
	return strings.TrimSpace(string(out))
}

func TestRenderGitConfig(t *testing.T) {
	base := t.TempDir()
	key := "bkenks/lazymux"
	dir := filepath.Join(base, "bkenks", "lazymux")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", dir, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %s", out)
	}

	cfg := config.Config{
		BaseDir:         base,
		PlaceholderHost: config.DefaultPlaceholderHost,
		Forges: []config.Forge{
			{Name: "github", Host: "github.com"},
			{Name: "forgejo", Host: "fj.homektb.com"},
		},
	}

	// Primary = forgejo, https.
	link := config.RepoLink{Forges: []string{"github", "forgejo"}, Primary: "forgejo", Scheme: "https"}
	if err := RenderGitConfig(cfg, key, link); err != nil {
		t.Fatal(err)
	}
	if got := gitCfg(t, dir, "remote.origin.url"); got != "https://lazymux-placeholder/bkenks/lazymux.git" {
		t.Errorf("origin = %q", got)
	}
	if got := gitCfg(t, dir, "url.https://fj.homektb.com/.insteadOf"); got != "https://lazymux-placeholder/" {
		t.Errorf("forgejo insteadOf = %q", got)
	}

	// Switch primary to github: the forgejo rule must be gone, github rule set.
	link.Primary = "github"
	if err := RenderGitConfig(cfg, key, link); err != nil {
		t.Fatal(err)
	}
	if got := gitCfg(t, dir, "url.https://github.com/.insteadOf"); got != "https://lazymux-placeholder/" {
		t.Errorf("github insteadOf = %q", got)
	}
	if got := gitCfg(t, dir, "url.https://fj.homektb.com/.insteadOf"); got != "" {
		t.Errorf("stale forgejo insteadOf still present: %q", got)
	}

	// Switch to ssh scheme: origin + insteadOf both flip to scp form, https gone.
	link.Scheme = "ssh"
	if err := RenderGitConfig(cfg, key, link); err != nil {
		t.Fatal(err)
	}
	if got := gitCfg(t, dir, "remote.origin.url"); got != "git@lazymux-placeholder:bkenks/lazymux.git" {
		t.Errorf("ssh origin = %q", got)
	}
	if got := gitCfg(t, dir, "url.git@github.com:.insteadOf"); got != "git@lazymux-placeholder:" {
		t.Errorf("ssh insteadOf = %q", got)
	}
	if got := gitCfg(t, dir, "url.https://github.com/.insteadOf"); got != "" {
		t.Errorf("stale https insteadOf still present: %q", got)
	}
}
