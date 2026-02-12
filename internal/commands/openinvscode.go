package commands

import (
	"os/exec"

	"github.com/bkenks/lazymux/internal/events"
	tea "github.com/charmbracelet/bubbletea"
)

func OpenInVSCode(repoFullPath string) tea.Cmd {
	cmdBuilder := exec.Command("code", repoFullPath) // build shell command

	cmd := tea.ExecProcess(
		cmdBuilder, // insert prior command
		func(err error) tea.Msg { return events.OpenInVSCodeComplete{Err: err} }, // run this function when done (i.e. emit Msg)
	)

	return cmd
}
