package commands

import (
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type CommandsMsg interface {
	isCommandMsg()
}

type SessionState int

const (
	StateMain SessionState = iota
	StateConfirmDelete
	StateCloneRepo
	StateBulkCloneRepos
)

type (
	MsgSetState struct{ State SessionState }
)

func (MsgSetState) isCommandMsg() {}
func SetState(state SessionState) tea.Cmd {
	return func() tea.Msg {
		return MsgSetState{
			State: state,
		}
	}
	
}


///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// uiCloneRepo: Cmds & Msgs

// Msgs
type (
	MsgQuitRepoDialog struct {}
	MsgGhqGet struct { err error }
	MsgGhqBulkCount struct { err error }
	MsgGhqRm struct {err error}
)

func (MsgQuitRepoDialog) isCommandMsg() {}
func (MsgGhqGet) isCommandMsg() {}
func (MsgGhqBulkCount) isCommandMsg() {}
func (MsgGhqRm) isCommandMsg() {}



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

func CmdFinishedCloningRepos() tea.Cmd {
	return func() tea.Msg {
		return MsgGhqGet{}
	}
}

func BulkCloneRepoAction(repoUrlsChunk string) tea.Cmd {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	repoUrls := strings.Split(strings.TrimSpace(repoUrlsChunk), "\n")

	for _, r := range repoUrls {
		c := exec.Command("ghq", "get", r)
		cmd = tea.ExecProcess(c, func(err error) tea.Msg {
			return MsgGhqBulkCount{ err: err }
		})
		cmds = append(cmds, cmd)
	}
	
	return tea.Batch(cmds...)
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
	MsgBulkCloneRepoDialog struct{}
)

func (MsgUpdateProjectList) isCommandMsg() {}
func (MsgConfirmDeleteDialog) isCommandMsg() {}
func (MsgBulkCloneRepoDialog) isCommandMsg() {}


/////////////////////////////////////
// UI Cmds
func BulkCloneRepoDialog() (tea.Cmd) {
	return func() tea.Msg {
		return MsgBulkCloneRepoDialog{}
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