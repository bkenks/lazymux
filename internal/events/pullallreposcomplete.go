package events

import (
	"fmt"
	"path/filepath"
	"strings"
)

// PullResult reports one repo's `git pull` outcome as it lands. An empty Reason
// means the pull succeeded; otherwise it carries the first line of git's output.
// An empty RepoPath means the pull-all failed before reaching any repo.
type PullResult struct {
	RepoPath string
	Reason   string
}

func (PullResult) isEvent() {}

// PullAllStarted hands the UI the total repo count and the channel that streams
// each PullResult, so the repo list can drain it and drive a progress bar.
type PullAllStarted struct {
	Total   int
	Results <-chan PullResult
}

func (PullAllStarted) isEvent() {}

// PullAllDrained is emitted once the results channel is closed — every repo has
// reported in.
type PullAllDrained struct{}

func (PullAllDrained) isEvent() {}

// PullAllReposComplete is the terminal summary the repo list emits after
// draining every PullResult; the app refreshes and toasts from it.
type PullAllReposComplete struct {
	Pulled  int
	Skipped []PullResult
}

func (PullAllReposComplete) isEvent() {}

// Summary is the one-line status shown when a pull-all finishes.
func (e PullAllReposComplete) Summary() string {
	if len(e.Skipped) == 0 {
		return fmt.Sprintf("Pulled %d repos.", e.Pulled)
	}
	names := make([]string, 0, len(e.Skipped))
	for _, s := range e.Skipped {
		if s.RepoPath == "" {
			names = append(names, s.Reason)
		} else {
			names = append(names, filepath.Base(s.RepoPath))
		}
	}
	return fmt.Sprintf("Pulled %d, skipped %d: %s",
		e.Pulled, len(e.Skipped), strings.Join(names, ", "))
}
