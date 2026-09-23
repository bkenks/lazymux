package commands

import (
	"os/exec"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/events"
)

func OpenInVSCode(repoFullPath string) tea.Cmd {
	cmdBuilder := exec.Command(cfg().Tools.Editor, repoFullPath)

	return tea.ExecProcess(
		cmdBuilder,
		func(err error) tea.Msg { return events.OpenInVSCodeComplete{Err: err} },
	)
}
