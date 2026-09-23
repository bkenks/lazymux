package commands

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/repomgr"
)

// RefreshReposCmd walks the lazymux base dir and rebuilds the repo list,
// ordered by the current domain.Sort mode.
func RefreshReposCmd() tea.Cmd {
	snapshot := cfg().Clone()
	sortMode := domain.Sort
	return func() tea.Msg {
		found, err := repomgr.List(snapshot)
		if err != nil {
			return events.Toast{Level: events.ToastError, Msg: fmt.Sprintf("scan failed: %v", err)}
		}

		repos := make([]list.Item, 0, len(found))
		for _, r := range found {
			repos = append(repos, r)
		}

		domain.SortRepos(repos, sortMode)

		return events.ReposRefreshed{RepoList: repos}
	}
}
