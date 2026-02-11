package commands

import (
	"os/exec"
	"strings"

	"github.com/bkenks/lazymux/internal/events"
	tea "github.com/charmbracelet/bubbletea"
)

func DeleteRepoCmd(repoGhqPath string) tea.Cmd {
	cmdBuilder := exec.Command("ghq", "rm", repoGhqPath) // build shell command
	cmdBuilder.Stdin = strings.NewReader("y")            // pipe "y" to terminal to accept ghq prompt asking if sure to remove repo

	cmd := tea.ExecProcess(
		cmdBuilder, // insert prior command
		func(err error) tea.Msg { return events.RepoDeleted{Err: err} }, // run this function when done (i.e. emit Msg)
	)

	return cmd
}
