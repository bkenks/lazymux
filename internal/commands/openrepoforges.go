package commands

import (
	"github.com/bkenks/lazymux/internal/events"
	tea "github.com/charmbracelet/bubbletea"
)

// OpenRepoForgesCmd opens the per-repo forge editor for a repo key.
func OpenRepoForgesCmd(key string) tea.Cmd {
	return func() tea.Msg {
		return events.OpenRepoForges{Key: key}
	}
}
