package commands

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// uiCloneRepo: Cmds & Msgs

// Msgs
type (
	MsgCloneRepoDialog struct {}
	MsgQuitRepoDialog struct {}
	MsgGhqGet struct { err error }
)

// UI Cmds
func QuitRepoDialog() tea.Cmd {
	return func() tea.Msg {
		return MsgQuitRepoDialog{}
	}
}

// External Action Cmds
func CloneRepoAction(repo string) tea.Cmd {
	c := exec.Command("ghq", "get", repo)

	cmd := tea.ExecProcess(c, func(err error) tea.Msg {
		return MsgGhqGet{err: err}
	})

	return tea.Cmd(cmd)
}

// End "uiCloneRepo: Cmds & Msgs"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////




///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// uiMain: Cmds & Msgs

// Msgs
type (
	MsgUpdateProjectList struct{}
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

// End "uiMain: Cmds & Msgs"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////