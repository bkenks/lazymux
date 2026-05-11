package commands

import (
	"os/exec"

	"github.com/bkenks/lazymux/internal/events"
	tea "github.com/charmbracelet/bubbletea"
)

func CloneReposExecCmd(repoUrls []string) tea.Cmd {
	var cmds []tea.Cmd
	ghqBin := cfg().Tools.Ghq

	for _, r := range repoUrls {
		cmds = append(cmds, tea.ExecProcess(
			exec.Command(ghqBin, "get", r),
			func(err error) tea.Msg {
				return events.CloneRepoExec{Err: err}
			},
		))
	}

	return tea.Batch(cmds...)
}
