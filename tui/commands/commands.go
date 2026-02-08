package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bkenks/lazymux/constants"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type CommandsMsg interface {
	isCommandMsg()
}


///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// State Management

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

// Implementations (p.s. setting "isCommandMsg" func on Msg gives it the type CommandsMsg so we can switch on it in ModelManager)
func (MsgSetState) isCommandMsg() {}

func SetState(state SessionState) tea.Cmd {
	return func() tea.Msg {
		return MsgSetState{
			State: state,
		}
	}
	
}
// End "State Management"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////




///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// External Cmds

type (
	MsgCmdErrorHandler struct { Err error } // used for error handling
	
	MsgReposRefreshed struct {
		RepoList []list.Item
	}
)

// Implementations (p.s. setting "isCommandMsg" func on Msg gives it the type CommandsMsg so we can switch on it in ModelManager)
func (MsgCmdErrorHandler) isCommandMsg() {}
func (MsgReposRefreshed) isCommandMsg() {}

/////////////////////////////////////
// Helper Functions
var MsgCmdCompleted tea.ExecCallback =
	func(err error) tea.Msg { return MsgCmdErrorHandler{Err: err} }

func TeaCmdBuilder(name string, arg ...string) tea.Cmd {
	cmdBuilder := exec.Command(name, arg...)
	
	cmd := tea.ExecProcess(
		cmdBuilder, // insert prior command
		MsgCmdCompleted,  // run this function when done (i.e. emit Msg)
	)

	return cmd
}

func RefreshReposCmd() tea.Cmd {
	
	cmd := exec.Command("ghq", "list") // Call ghq to list repositories
	out, err := cmd.Output()

	if err != nil { // Fail-Safe
		fmt.Println("Error getting Repo List:", err)
		os.Exit(1)
	}


	// string(out) → "github.com/user/Repo1\ngithub.com/user/Repo2\n"
	// strings.TrimSpace(...) → removes the final \n, giving "github.com/user/Repo1\ngithub.com/user/Repo2"
	// strings.Split(..., "\n") → splits into strings on "\n"
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	repos := make([]list.Item, 0, len(lines)) // preallocate slice (i.e. set array size)


	/////////////////////////////////////
	// Format lines (list of repos like "github.com/user/Repo1") into a []list.Item
	for _, line := range lines {
		if line == "" { // Fail-Safe
			continue
		}
		
		parts := strings.Split(line, "/") // split path ("github.com/user/Repo1") by "/"
		nameFromSplit := parts[len(parts)-1] // grab the last element from the split (which is the repo name)

		repos = append(repos, constants.Repo{ // Add to a []list.Item (array of `list.Item`s)
			Name: nameFromSplit,
			Path: line,
		})
	}
	//
	/////////////////////////////////////

	return func() tea.Msg {
		return MsgReposRefreshed{ RepoList: repos }
	}
}

type (
	MsgStartRepoClone struct { RepoUrls []string }
	MsgRepoCloneInit struct { TotalRepos int }
	MsgRepoCloned struct { Err error }
)
func (MsgStartRepoClone) isCommandMsg() {}
func (MsgRepoCloneInit) isCommandMsg() {}
func (MsgRepoCloned) isCommandMsg() {}

func StartCloneReposCmd(repoUrlsChunk string) tea.Cmd {
	repoUrls := strings.Split(strings.TrimSpace(repoUrlsChunk), "\n")

	return func() tea.Msg {
		return MsgStartRepoClone{
			RepoUrls: repoUrls,
		}
	}
}

func CloneReposExecCmd(repoUrls []string) tea.Cmd {
	var cmds []tea.Cmd

	for _, r := range repoUrls {
		cmds = append(cmds, tea.ExecProcess(
			exec.Command("ghq", "get", r),
			func(err error) tea.Msg {
				return MsgRepoCloned{Err: err}
			},
		))
	}

	return tea.Batch(cmds...)
}

type MsgRepoDeleted struct { Err error }
func (MsgRepoDeleted) isCommandMsg() {}

func DeleteRepoCmd(repoGhqPath string) tea.Cmd {
	cmdBuilder := exec.Command("ghq", "rm", repoGhqPath) // build shell command
	cmdBuilder.Stdin = strings.NewReader("y") // pipe "y" to terminal to accept ghq prompt asking if sure to remove repo

	cmd := tea.ExecProcess(
		cmdBuilder, // insert prior command
		func(err error) tea.Msg { return MsgRepoDeleted{ Err: err } }, // run this function when done (i.e. emit Msg)
	)

	return cmd
}
// End "Helper Functions"
/////////////////////////////////////


// End "External Cmds"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////