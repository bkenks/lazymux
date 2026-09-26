package commands

import (
	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/gitkeeper/internal/events"
	"github.com/bkenks/gitkeeper/internal/repomgr"
)

// DeleteRepoCmd removes the repo stored under key at absPath, along with any
// namespace parents it leaves empty.
func DeleteRepoCmd(key, absPath string) tea.Cmd {
	baseDir := cfg().RepoRoot()
	return func() tea.Msg {
		err := repomgr.Remove(baseDir, absPath)
		return events.RepoDeleted{Key: key, Err: err}
	}
}
