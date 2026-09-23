package commands

import (
	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
)

// SortModeChangedCmd reports a new repo list sort order so it gets persisted.
func SortModeChangedCmd(mode domain.SortMode) tea.Cmd {
	return func() tea.Msg { return events.SortModeChanged{Mode: mode} }
}
