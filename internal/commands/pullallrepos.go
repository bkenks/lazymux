package commands

import (
	"os/exec"
	"strings"
	"sync"

	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/repomgr"
	tea "github.com/charmbracelet/bubbletea"
)

// PullAllReposCmd scans every managed repo and kicks off `git pull --ff-only`
// against each in parallel (capped at 8 concurrent network ops), streaming one
// PullResult per repo down a channel. It returns immediately with a
// PullAllStarted carrying the total and that channel; the UI drains it via
// WaitForPullCmd to drive a live progress bar. --ff-only guarantees we never
// leave a repo half-merged, so repos that can't fast-forward are just skipped.
func PullAllReposCmd() tea.Cmd {
	return func() tea.Msg {
		found, err := repomgr.List(cfg())
		if err != nil {
			ch := make(chan events.PullResult, 1)
			ch <- events.PullResult{Reason: "scan failed: " + err.Error()}
			close(ch)
			return events.PullAllStarted{Total: 1, Results: ch}
		}

		paths := make([]string, 0, len(found))
		for _, r := range found {
			if r.AbsPath != "" {
				paths = append(paths, r.AbsPath)
			}
		}

		ch := make(chan events.PullResult, len(paths))
		go func() {
			var wg sync.WaitGroup
			sem := make(chan struct{}, 8) // cap concurrent network ops
			for _, p := range paths {
				wg.Add(1)
				sem <- struct{}{}
				go func(path string) {
					defer wg.Done()
					defer func() { <-sem }()

					output, err := exec.Command("git", "-C", path, "pull", "--ff-only").CombinedOutput()
					if err != nil {
						ch <- events.PullResult{RepoPath: path, Reason: firstLine(string(output))}
						return
					}
					ch <- events.PullResult{RepoPath: path}
				}(p)
			}
			wg.Wait()
			close(ch)
		}()

		return events.PullAllStarted{Total: len(paths), Results: ch}
	}
}

// WaitForPullCmd blocks on the next PullResult from the pull-all channel,
// returning PullAllDrained once the channel is closed.
func WaitForPullCmd(ch <-chan events.PullResult) tea.Cmd {
	return func() tea.Msg {
		r, ok := <-ch
		if !ok {
			return events.PullAllDrained{}
		}
		return r
	}
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
