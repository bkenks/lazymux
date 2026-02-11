package commands

import (
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	tea "github.com/charmbracelet/bubbletea"
)

func SetState(state domain.SessionState) tea.Cmd {
	return func() tea.Msg {
		return events.SetState{
			State: state,
		}
	}

}
