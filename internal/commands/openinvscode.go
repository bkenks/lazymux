package commands

import (
	"os/exec"

	"github.com/bkenks/lazymux/internal/events"
	tea "github.com/charmbracelet/bubbletea"
)

func OpenInVSCode(repoFullPath string) tea.Cmd {
	cmdBuilder := exec.Command(cfg().Tools.Editor, repoFullPath)

	return tea.ExecProcess(
		cmdBuilder,
		func(err error) tea.Msg { return events.OpenInVSCodeComplete{Err: err} },
	)
}
