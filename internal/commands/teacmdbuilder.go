package commands

import (
	"os/exec"

	"github.com/bkenks/lazymux/internal/events"
	tea "github.com/charmbracelet/bubbletea"
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
