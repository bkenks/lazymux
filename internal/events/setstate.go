package events

import (
	"github.com/bkenks/lazymux/internal/domain"
)

type SetState struct{ State domain.SessionState }

// Implementations (p.s. setting "isCommandMsg" func on Msg gives it the type CommandsMsg so we can switch on it in ModelManager)
func (SetState) isEvent() {}
