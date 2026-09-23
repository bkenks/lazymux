package commands

import (
	"os/exec"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/events"
)

func TeaCmdBuilder(name string, arg ...string) tea.Cmd {
	cmdBuilder := exec.Command(name, arg...)

	cmdComplete := func(err error) tea.Msg { return events.CmdComplete{Err: err} }

	cmd := tea.ExecProcess(
		cmdBuilder,  // insert prior command
		cmdComplete, // run this function when done (i.e. emit Msg)
	)

	return cmd
}
