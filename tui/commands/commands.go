package commands

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

type (
	MsgUpdateProjectList struct{}
	MsgCloneRepoDialog struct {}
	MsgQuitRepoDialog struct {}
	MsgGhqGet struct { err error }
)

func CloneRepoDialog() (tea.Cmd) {
	return func() tea.Msg {
		return MsgCloneRepoDialog{}
	}
}

func QuitRepoDialog() tea.Cmd {
	return func() tea.Msg {
		return MsgQuitRepoDialog{}
	}
}

func UpdateRepoList() (tea.Cmd) {
	return func() tea.Msg {
		return MsgUpdateProjectList{}
	}
}

func CloneRepoAction (repo string) tea.Cmd {
	c := exec.Command("ghq", "get", repo)

	cmd := tea.ExecProcess(c, func(err error) tea.Msg {
		return MsgGhqGet{err: err}
	})

	return tea.Cmd(cmd)
}