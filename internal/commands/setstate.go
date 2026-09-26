package commands

import (
	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/gitkeeper/internal/domain"
	"github.com/bkenks/gitkeeper/internal/events"
)

func SetState(state domain.SessionState) tea.Cmd {
	return func() tea.Msg {
		return events.SetState{
			State: state,
		}
	}

}
