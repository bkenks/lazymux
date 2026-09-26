package commands

import (
	"context"
	"os"
	"os/exec"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/gitkeeper/internal/events"
	"github.com/bkenks/gitkeeper/internal/repomgr"
)

// pullTimeout bounds one repo's `git pull`, so a hung network connection or a
// credential prompt can't keep a pull-all from finishing.
const pullTimeout = 2 * time.Minute

// PullAllReposCmd scans every managed repo and kicks off `git pull --ff-only`
// against each in parallel (capped at 8 concurrent network ops), streaming one
// PullResult per repo down a channel. It returns immediately with a
// PullAllStarted carrying the total and that channel; the UI drains it via
// WaitForPullCmd to drive a live progress bar. --ff-only guarantees we never
// leave a repo half-merged, so repos that can't fast-forward are just skipped.
func PullAllReposCmd() tea.Cmd {
	snapshot := cfg().Clone()
	return func() tea.Msg {
		found, err := repomgr.List(snapshot)
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
					ch <- pullRepo(path)
				}(p)
			}
			wg.Wait()
			close(ch)
		}()

		return events.PullAllStarted{Total: len(paths), Results: ch}
	}
}

// pullRepo fast-forwards one repo. Git is told not to prompt for credentials,
// since the TUI owns the terminal while a pull-all runs.
func pullRepo(path string) events.PullResult {
	ctx, cancel := context.WithTimeout(context.Background(), pullTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", path, "pull", "--ff-only")
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	output, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return events.PullResult{RepoPath: path, Reason: "timed out after " + pullTimeout.String()}
	}
	if err != nil {
		return events.PullResult{RepoPath: path, Reason: repomgr.FirstLine(string(output))}
	}
	return events.PullResult{RepoPath: path}
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
