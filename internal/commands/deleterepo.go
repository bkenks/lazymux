package commands

import (
	"os/exec"
	"strings"

	"github.com/bkenks/lazymux/internal/events"
	tea "github.com/charmbracelet/bubbletea"
)

func DeleteRepoCmd(repoGhqPath string) tea.Cmd {
	cmdBuilder := exec.Command(cfg().Tools.Ghq, "rm", repoGhqPath)
	cmdBuilder.Stdin = strings.NewReader("y") // ghq prompts to confirm; pipe "y"

	return tea.ExecProcess(
		cmdBuilder,
		func(err error) tea.Msg { return events.RepoDeleted{Err: err} },
	)
}
