package repomgr

import (
	"os/exec"
	"slices"
	"strings"
	"testing"
)

// newCommittedRepo returns a repo with one commit and, when withRemote is
// set, a bare repo wired up as its origin.
func newCommittedRepo(t *testing.T, withRemote bool) (dir, remote string) {
	t.Helper()
	dir = t.TempDir()
	mustGit(t, dir, "init")
	mustGit(t, dir, "config", "user.email", "t@example.com")
	mustGit(t, dir, "config", "user.name", "T")
	mustGit(t, dir, "commit", "--allow-empty", "-m", "first")
	if withRemote {
		remote = t.TempDir()
		mustGit(t, remote, "init", "--bare")
		mustGit(t, dir, "remote", "add", "origin", remote)
	}
	return dir, remote
}

func remoteTags(t *testing.T, remote string) []string {
	t.Helper()
	out, err := exec.Command("git", "-C", remote, "tag", "--list").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.Fields(string(out))
}

func TestReleaseTagCreatesAndPushesAnnotatedTag(t *testing.T) {
	dir, remote := newCommittedRepo(t, true)

	if err := ReleaseTag(dir, "mypkg/v1.2.3"); err != nil {
		t.Fatal(err)
	}

	if got := remoteTags(t, remote); !slices.Equal(got, []string{"mypkg/v1.2.3"}) {
		t.Errorf("remote tags = %v, want the new tag", got)
	}
	out, err := exec.Command("git", "-C", dir, "cat-file", "-t", "mypkg/v1.2.3").Output()
	if err != nil || strings.TrimSpace(string(out)) != "tag" {
		t.Errorf("tag object type = %q (%v), want an annotated tag", out, err)
	}
}

func TestReleaseTagKeepsLocalTagWhenPushFails(t *testing.T) {
	dir, _ := newCommittedRepo(t, false)

	err := ReleaseTag(dir, "v0.0.1")

	if err == nil || !strings.Contains(err.Error(), "created v0.0.1 locally") {
		t.Fatalf("err = %v, want a push failure naming the local tag", err)
	}
	if tags, _ := ListTags(dir); !slices.Equal(tags, []string{"v0.0.1"}) {
		t.Errorf("local tags = %v, want the created tag kept", tags)
	}
}

func TestReleaseTagRefusesExistingTag(t *testing.T) {
	dir, remote := newCommittedRepo(t, true)
	mustGit(t, dir, "tag", "v1.0.0")

	if err := ReleaseTag(dir, "v1.0.0"); err == nil {
		t.Fatal("err = nil, want a refusal to overwrite the tag")
	}
	if got := remoteTags(t, remote); len(got) != 0 {
		t.Errorf("remote tags = %v, want nothing pushed", got)
	}
}

func TestReleaseTagFailsWithoutCommits(t *testing.T) {
	dir := t.TempDir()
	mustGit(t, dir, "init")

	if err := ReleaseTag(dir, "v0.0.1"); err == nil {
		t.Fatal("err = nil, want a failure tagging a repo with no commits")
	}
}

func TestListTags(t *testing.T) {
	dir, _ := newCommittedRepo(t, false)
	if tags, err := ListTags(dir); err != nil || len(tags) != 0 {
		t.Fatalf("ListTags = %v, %v; want none", tags, err)
	}
	mustGit(t, dir, "tag", "v1.0.0")
	mustGit(t, dir, "tag", "pkg/v2.0.0")

	tags, err := ListTags(dir)
	if err != nil || !slices.Equal(tags, []string{"pkg/v2.0.0", "v1.0.0"}) {
		t.Errorf("ListTags = %v, %v", tags, err)
	}
}

func TestCheckTagName(t *testing.T) {
	for _, ok := range []string{"v0.0.0", "mypkg/v0.0.0", "0.0.0-mypkg"} {
		if err := CheckTagName(ok); err != nil {
			t.Errorf("CheckTagName(%q) = %v, want nil", ok, err)
		}
	}
	for _, bad := range []string{"my pkg/0.0.0", "v..0.0.0", "0.0.0.lock", "v0.0.0/", "~v0.0.0"} {
		if err := CheckTagName(bad); err == nil {
			t.Errorf("CheckTagName(%q) = nil, want an error", bad)
		}
	}
}
