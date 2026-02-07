package commands

import (
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// uiCloneRepo: Cmds & Msgs

// Msgs
type (
	MsgCloneRepoDialog struct {}
	MsgQuitRepoDialog struct {}
	MsgGhqGet struct { err error }
	MsgGhqRm struct {err error}
)

// UI Cmds
func QuitRepoDialog() tea.Cmd {
	return func() tea.Msg {
		return MsgQuitRepoDialog{}
	}
}

/////////////////////////////////////
// External Action Cmds
func CloneRepoAction(repoUrl string) tea.Cmd {
	c := exec.Command("ghq", "get", repoUrl)
	cmd := tea.ExecProcess(c, func(err error) tea.Msg {
		return MsgGhqGet{err: err}
	})

	return cmd
}
// End "External Action Cmds"
/////////////////////////////////////

// End "uiCloneRepo: Cmds & Msgs"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////




///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// uiMain: Cmds & Msgs

// Msgs
type (
	MsgUpdateProjectList struct{}
	MsgConfirmDeleteDialog struct{
		repoPath string
	}
	MsgConfirmDeleteDialogQuit struct{}
)


/////////////////////////////////////
// UI Cmds
func CloneRepoDialog() (tea.Cmd) {
	return func() tea.Msg {
		return MsgCloneRepoDialog{}
	}
}

func UpdateRepoList() (tea.Cmd) {
	return func() tea.Msg {
		return MsgUpdateProjectList{}
	}
}

func ConfirmDeleteDialog() (tea.Cmd) {
	return func() tea.Msg {
		return MsgConfirmDeleteDialog{}
	}
}

func ConfirmDeleteDialogQuit() (tea.Cmd) {
	return func() tea.Msg {
		return MsgConfirmDeleteDialogQuit{}
	}
}

// End "UI Cmds"
/////////////////////////////////////


// External Action Cmds
func OpenLazygitAction(path string) tea.Cmd {
	c := exec.Command("lazygit", "-p", path)

	type lgFinishedMsg struct { err error }

	cmd := tea.ExecProcess(c, func(err error) tea.Msg {
		return lgFinishedMsg{err: err}
	})
	
	return tea.Cmd(cmd)
}

func DeleteRepoAction(repoGhqPath string) tea.Cmd {
	c := exec.Command("ghq", "rm", repoGhqPath)

	c.Stdin = strings.NewReader("y")

	cmd := tea.ExecProcess(c, func(err error) tea.Msg {
		return MsgGhqRm{err: err}
	})

	return cmd
}

// End "uiMain: Cmds & Msgs"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////