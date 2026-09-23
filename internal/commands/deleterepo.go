package commands

import (
	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/repomgr"
)

// DeleteRepoCmd removes a repo directory (and now-empty namespace parents).
func DeleteRepoCmd(absPath string) tea.Cmd {
	baseDir := cfg().BaseDir
	return func() tea.Msg {
		err := repomgr.Remove(baseDir, absPath)
		return events.RepoDeleted{Err: err}
	}
}
