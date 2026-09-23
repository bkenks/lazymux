package commands

import (
	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/events"
)

// OpenRepoForgesCmd opens the per-repo forge editor for a repo key.
func OpenRepoForgesCmd(key string) tea.Cmd {
	return func() tea.Msg {
		return events.OpenRepoForges{Key: key}
	}
}
