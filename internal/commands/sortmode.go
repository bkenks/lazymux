package commands

import (
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	tea "github.com/charmbracelet/bubbletea"
)

// SortModeChangedCmd reports a new repo list sort order so it gets persisted.
func SortModeChangedCmd(mode domain.SortMode) tea.Cmd {
	return func() tea.Msg { return events.SortModeChanged{Mode: mode} }
}
