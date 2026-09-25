package repomgr

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// pushTagTimeout bounds `git push` of a release tag, so a hung connection or a
// credential prompt can't leave the release pending forever.
const pushTagTimeout = 2 * time.Minute

// ListTags returns every tag name in the repo at dir.
func ListTags(dir string) ([]string, error) {
	out, err := runGit(dir, "tag", "--list")
	if err != nil {
		return nil, err
	}
	return strings.Fields(out), nil
}

// CheckTagName reports why name can't be a git tag, or nil if it can.
func CheckTagName(name string) error {
	err := exec.Command("git", "check-ref-format", "refs/tags/"+name).Run()
	if err != nil {
		return fmt.Errorf("%q isn't a valid git tag name", name)
	}
	return nil
}

// ReleaseTag creates an annotated tag at HEAD and pushes just that tag to
// origin, which fans out to every upstream through the origin's push URLs.
// Git is told not to prompt for credentials, since the TUI owns the terminal.
// When the push fails the tag stays in the local repo, and the error says so.
func ReleaseTag(dir, tag string) error {
	if _, err := runGit(dir, "tag", "--annotate", "--message", tag, tag); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), pushTagTimeout)
	defer cancel()
	push := exec.CommandContext(ctx, "git", "-C", dir, "push", "origin", "refs/tags/"+tag)
	push.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := push.CombinedOutput()
	if ctx.Err() != nil {
		return fmt.Errorf("created %s locally, but the push timed out after %s", tag, pushTagTimeout)
	}
	if err != nil {
		return fmt.Errorf("created %s locally, but the push failed: %s", tag, FirstLine(string(out)))
	}
	return nil
}
