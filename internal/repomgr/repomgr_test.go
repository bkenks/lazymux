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
