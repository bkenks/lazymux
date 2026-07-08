package commands

import (
	"fmt"
	"sort"

	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/repomgr"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// RefreshReposCmd walks the lazymux base dir and rebuilds the repo list,
// most-recently-interacted first.
func RefreshReposCmd() tea.Cmd {
	return func() tea.Msg {
		found, err := repomgr.List(cfg())
		if err != nil {
			return events.Toast{Level: events.ToastError, Msg: fmt.Sprintf("scan failed: %v", err)}
		}

		repos := make([]list.Item, 0, len(found))
		for _, r := range found {
			repos = append(repos, r)
		}

		// Most recently interacted first; never-interacted fall to the end.
		sort.SliceStable(repos, func(i, j int) bool {
			ri := repos[i].(domain.Repo)
			rj := repos[j].(domain.Repo)
			return ri.LastInteracted.After(rj.LastInteracted)
		})

		return events.ReposRefreshed{RepoList: repos}
	}
}
