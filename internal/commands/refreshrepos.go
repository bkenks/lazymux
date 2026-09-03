package commands

import (
	"fmt"

	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/repomgr"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// RefreshReposCmd walks the lazymux base dir and rebuilds the repo list,
// ordered by the current domain.Sort mode.
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

		domain.SortRepos(repos, domain.Sort)

		return events.ReposRefreshed{RepoList: repos}
	}
}
