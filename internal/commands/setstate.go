package commands

import (
	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
)

func SetState(state domain.SessionState) tea.Cmd {
	return func() tea.Msg {
		return events.SetState{
			State: state,
		}
	}

}
