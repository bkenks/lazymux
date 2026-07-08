package commands

import (
	"os/exec"
	"strings"
	"sync"

	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/repomgr"
	tea "github.com/charmbracelet/bubbletea"
)

// PullAllReposCmd runs `git pull --ff-only` against every managed repo in
// parallel. Repos that can't be fast-forwarded (merge conflicts, divergent
// branches, dirty working tree, missing upstream) are skipped and reported
// back via PullAllReposComplete. --ff-only guarantees we never leave a repo
// in a half-merged state, so no cleanup is needed on failure.
func PullAllReposCmd() tea.Cmd {
	return func() tea.Msg {
		found, err := repomgr.List(cfg())
		if err != nil {
			return events.PullAllReposComplete{
				Skipped: []events.SkippedPull{{Reason: "scan failed: " + err.Error()}},
			}
		}

		paths := make([]string, 0, len(found))
		for _, r := range found {
			paths = append(paths, r.AbsPath)
		}

		var (
			mu      sync.Mutex
			wg      sync.WaitGroup
			pulled  int
			skipped []events.SkippedPull
		)

		sem := make(chan struct{}, 8) // cap concurrent network ops

		for _, p := range paths {
			if p == "" {
				continue
			}
			wg.Add(1)
			sem <- struct{}{}
			go func(path string) {
				defer wg.Done()
				defer func() { <-sem }()

				output, err := exec.Command("git", "-C", path, "pull", "--ff-only").CombinedOutput()
				if err != nil {
					mu.Lock()
					skipped = append(skipped, events.SkippedPull{
						RepoPath: path,
						Reason:   firstLine(string(output)),
					})
					mu.Unlock()
					return
				}
				mu.Lock()
				pulled++
				mu.Unlock()
			}(p)
		}

		wg.Wait()
		return events.PullAllReposComplete{Pulled: pulled, Skipped: skipped}
	}
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
